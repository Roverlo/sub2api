# Codex 项目提示

本文件给 Codex / 自动化代理使用，范围是整个 `C:\ALL_in_H\sub2api` 仓库。用户的最新明确指令优先于本文件。

## 本地环境

- 本地终端是 Windows PowerShell。
- 如果编译、测试或构建过程中缺失工具，先询问用户是否需要下载安装；用户同意后再安装。
- 不要清理、重置或删除与当前任务无关的本地改动。当前仓库可能存在旧的未跟踪目录，例如 `.playwright-cli/`，不要因为它未跟踪就顺手删除。

## Git 和 GitHub

- 每次代码或文档更新后，更新 Git，并推送到用户 GitHub fork。
- 如果还没有 fork，先为用户创建 fork，再推送。
- 当前常用 remotes：
  - `origin`: `https://github.com/Wei-Shaw/sub2api.git`
  - `fork`: `https://github.com/Roverlo/sub2api.git`
- 本机访问 GitHub 通常需要本地代理 `127.0.0.1:10808`。Git 默认可能没有读取 Windows 用户代理配置；WinHTTP、环境变量、Git 配置都可能是直连。
- 对 GitHub fetch / push 使用 per-command 代理，避免污染全局配置：

```powershell
git -c http.proxy=http://127.0.0.1:10808 `
    -c https.proxy=http://127.0.0.1:10808 `
    fetch --tags origin

git -c http.proxy=http://127.0.0.1:10808 `
    -c https.proxy=http://127.0.0.1:10808 `
    push fork <branch>
```

## 上游合并和验证

- 检查原项目更新时，先取 `origin` 的分支和 tags，再比较当前分支、本地自定义提交、上游 `main` 或目标 tag。
- 合并上游时要保守解决冲突，优先保留本地已经上线的定制能力，避免大范围重构。
- 常用验证：

```powershell
git diff --check

Push-Location backend
go test ./...
Pop-Location

Push-Location frontend
pnpm typecheck
Pop-Location
```

- 如果冲突涉及定价、模型或 `codex-auto-review`，额外跑相关后端测试。历史上该区域容易出现 pricing JSON 和测试期望不一致。

## VPS 连接信息

- VPS 地址：`192.3.89.62`
- SSH 端口：`20002`
- SSH 用户：`root`
- 本机私钥：`C:\Users\胡文雨\.ssh\sub2api_vps_ed25519`
- 本机公钥：`C:\Users\胡文雨\.ssh\sub2api_vps_ed25519.pub`
- 注意：用户曾误写过 `200002`，这是无效端口；实际可用端口是 `20002`。
- 不要把私钥内容写入仓库、日志、Issue、PR 或聊天回复。文档中只允许引用本机路径。

连接命令：

```powershell
ssh -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" -p 20002 root@192.3.89.62
```

上传文件命令：

```powershell
scp -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" -P 20002 `
  <local-file> `
  root@192.3.89.62:/opt/sub2api/backups/
```

## VPS 部署边界

- 除非用户明确要求部署，否则不要推送、发布或切换 VPS。
- 用户曾明确要求过“先合并更新，但是先不要推送到我的 VPS”；后续相同场景默认只合并、验证、推送 GitHub fork，不部署 VPS。
- VPS 上的项目目录：`/opt/sub2api`
- Compose 文件：`/opt/sub2api/docker-compose.yml`
- 应用容器：`sub2api`
- Compose 线上镜像标签：`sub2api:ai-sdk-compat`
- 公网流量路径：Nginx `443/80` -> 宿主机本地端口 -> 容器 `8080`
- 常用本机健康检查端口：`http://127.0.0.1:18080/health`
- PostgreSQL 和 Redis 是独立容器；应用镜像更新时只重建 `sub2api`，不要无故重启数据库或 Redis。

## VPS 更新流程

推荐流程是“本机构建镜像 -> 上传 tar -> VPS 加载镜像 -> VPS 后台脚本切换并自动回滚”。不要在 PowerShell 里使用 `docker save | ssh docker load` 直传二进制流，历史上会导致远端 `invalid tar header`。

本机构建：

```powershell
$version = git describe --tags --always --dirty
$commit = git rev-parse --short HEAD
$tag = "sub2api:<new-image-tag>"

docker build `
  --build-arg PNPM_VERSION=11.0.0 `
  --build-arg VERSION=$version `
  --build-arg COMMIT=$commit `
  -t $tag .

docker run --rm $tag --version
```

保存并上传：

```powershell
$tar = "$env:TEMP\sub2api-<new-image-tag>.tar"
docker save -o $tar sub2api:<new-image-tag>

scp -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" -P 20002 `
  $tar `
  root@192.3.89.62:/opt/sub2api/backups/
```

VPS 加载：

```bash
docker load -i /opt/sub2api/backups/sub2api-<new-image-tag>.tar
```

切换时使用 VPS 侧后台脚本，例如 `nohup`，这样当前 SSH / Codex 会话断开也不会影响切换或回滚。脚本至少要做：

- 确认新镜像存在。
- 确认当前服务健康。
- 把当前 `sub2api:ai-sdk-compat` 另存为回滚镜像。
- 切换前导出 PostgreSQL 备份。
- 将新镜像 tag 成 `sub2api:ai-sdk-compat`。
- 只重建 `sub2api` 应用容器。
- 轮询 Docker health 和 HTTP `/health`。
- 如果健康检查失败，自动恢复旧镜像标签并重建应用容器。

核心切换命令：

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

切换后验证：

```bash
docker inspect sub2api --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
curl -sS -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18080/health
docker exec sub2api /app/sub2api --version
docker image inspect sub2api:ai-sdk-compat --format '{{.Id}}'
docker image inspect sub2api:<new-image-tag> --format '{{.Id}}'
```

`docker ps` 仍会显示 Compose 配置里的 `sub2api:ai-sdk-compat` 标签；判断是否真的切到新镜像，要比较 image ID 或查看容器内 `/app/sub2api --version`。
