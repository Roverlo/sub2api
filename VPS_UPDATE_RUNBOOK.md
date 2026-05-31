# VPS 低中断更新 Runbook

这份文档记录本次 Sub2API VPS 更新经验。核心思路是把“镜像交付”和“服务切换”拆开：

1. 先在本机构建镜像，并提前加载到 VPS。
2. 确认新镜像已经在 VPS 上后，再切换应用容器。
3. 切换动作放到 VPS 后台脚本里执行，脚本负责健康检查和失败回滚。这样即使当前 SSH、代理或 Codex 会话在容器重建时短暂断开，VPS 也能自己完成切换或回滚。

## 部署形态

本次观察到的生产形态：

- Compose 项目目录：`/opt/sub2api`
- Compose 文件：`/opt/sub2api/docker-compose.yml`
- 应用容器：`sub2api`
- Compose 使用的线上镜像标签：`sub2api:ai-sdk-compat`
- 公网流量：Nginx `443/80` -> 宿主机本地端口 -> 容器 `8080`
- 数据持久化目录：
  - `/opt/sub2api/data`
  - `/opt/sub2api/postgres_data`
  - `/opt/sub2api/redis_data`

这是单应用容器部署，不是蓝绿或滚动发布。替换应用容器时，正在进行的请求、流式响应或 WebSocket 连接可能短暂中断；PostgreSQL 和 Redis 不需要跟着重启。

## 本机构建镜像

使用带版本和 commit 的显式标签，方便后续排查。

```powershell
git describe --tags --always --dirty
git rev-parse --short HEAD

docker build `
  --build-arg PNPM_VERSION=11.0.0 `
  --build-arg VERSION=<version-from-git-describe> `
  --build-arg COMMIT=<short-commit> `
  -t sub2api:<new-image-tag> .

docker run --rm sub2api:<new-image-tag> --version
```

经验点：Docker 构建前端依赖时，必须在 `pnpm install --frozen-lockfile` 前复制 `frontend/pnpm-workspace.yaml`。否则 pnpm 看不到 `allowBuilds` 配置，可能因为 `esbuild`、`vue-demi` 等 approved build scripts 失败。

## 传输镜像到 VPS

不要在 PowerShell 里直接用管道传二进制 tar 流，例如 `docker save ... | ssh ... docker load`。PowerShell 可能破坏二进制流，导致远端 `docker load` 报 `invalid tar header`。

推荐先保存成 tar 文件，再用 `scp` 上传。

```powershell
$tag = "sub2api:<new-image-tag>"
$tar = "$env:TEMP\sub2api-<new-image-tag>.tar"

docker save -o $tar $tag

scp -i <ssh-key> -P <ssh-port> `
  $tar `
  <user>@<vps-host>:/opt/sub2api/backups/sub2api-<new-image-tag>.tar

ssh -i <ssh-key> -p <ssh-port> <user>@<vps-host> `
  "docker load -i /opt/sub2api/backups/sub2api-<new-image-tag>.tar"
```

加载镜像不会影响当前正在运行的服务。

## 后台切换与自动回滚

切换脚本应放在 VPS 上后台执行，例如使用 `nohup`。这样当前操作会话即使因为代理短暂中断，也不会影响远程切换流程。

脚本至少应做这些事情：

- 确认新镜像和线上镜像标签都存在。
- 切换前确认当前服务健康。
- 把旧线上镜像标签另存为回滚镜像。
- 切换前导出 PostgreSQL 备份。
- 把新镜像打到 Compose 正在使用的线上标签。
- 只重建 `sub2api` 应用容器，不动 PostgreSQL 和 Redis。
- 等待容器 healthcheck 和 HTTP `/health` 恢复。
- 如果健康检查失败，自动把线上标签改回旧镜像并重建应用容器。

核心切换命令示例：

```bash
cd /opt/sub2api

ts="$(date +%Y%m%d%H%M%S)"
old_backup="sub2api:backup-before-update-$ts"
db_backup="/opt/sub2api/backups/pre-switch-db-$ts.sql"

docker tag sub2api:ai-sdk-compat "$old_backup"
docker exec sub2api-postgres pg_dump -U sub2api -d sub2api > "$db_backup"
chmod 600 "$db_backup"

docker tag sub2api:<new-image-tag> sub2api:ai-sdk-compat
docker compose up -d --no-deps --force-recreate sub2api
```

健康检查：

```bash
docker inspect sub2api --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:<local-port>/health
docker exec sub2api /app/sub2api --version
```

本次实际切换中，应用容器从开始重建到 `healthy` 大约 8 秒；有效不可用窗口只有几秒。

## 切换后验证

```bash
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' | grep sub2api
docker exec sub2api /app/sub2api --version
docker image inspect sub2api:ai-sdk-compat --format '{{.Id}}'
docker image inspect sub2api:<new-image-tag> --format '{{.Id}}'
curl -sS -o /dev/null -w 'health_http=%{http_code} total=%{time_total}\n' http://127.0.0.1:<local-port>/health
```

注意：`docker ps` 仍会显示 Compose 配置里的镜像标签，比如 `sub2api:ai-sdk-compat`。判断是否真的切到新镜像，应比较线上标签和新镜像的 image ID，或直接看 `/app/sub2api --version`。

如需确认数据库迁移：

```bash
docker exec -i sub2api-postgres psql -U sub2api -d sub2api -At <<'SQL'
SELECT filename, applied_at
FROM schema_migrations
WHERE filename = '<migration-file-name>';
SQL
```

## 回滚

如果新容器健康检查失败，切换脚本应自动执行：

```bash
cd /opt/sub2api
docker tag sub2api:backup-before-update-YYYYMMDDHHMMSS sub2api:ai-sdk-compat
docker compose up -d --no-deps --force-recreate sub2api
```

然后再次验证：

```bash
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:<local-port>/health
docker exec sub2api /app/sub2api --version
```

数据库回滚要单独评估。即使迁移看起来低风险，也建议切换前保留一次 `pg_dump`。

## 经验总结

- 构建和传输镜像可以提前完成，不影响线上服务。
- 真正有中断风险的是 `docker compose up -d --no-deps --force-recreate sub2api`。
- 应用镜像更新时，只重建应用容器，PostgreSQL 和 Redis 保持运行。
- 如果当前操作会话依赖正在更新的代理服务，切换必须由 VPS 后台脚本独立完成。
- 镜像构建时注入明确的版本和 commit，方便切换后确认。
- PowerShell 不适合直接透传 `docker save` 的二进制流，使用 `docker save -o` 加 `scp` 更稳。
- 切换前保留旧镜像标签和数据库备份，回滚成本会低很多。
