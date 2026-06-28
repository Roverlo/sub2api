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
- 如果上游再次引入或修改类似“登录后强制部署/运营/合规确认、必须输入指定短语或字符才能继续使用控制台”的门禁，不要合入本地定制分支；合并时应保留本地禁用状态，避免恢复前端强制弹窗、`admin-compliance-required` 事件或后端 `423 ADMIN_COMPLIANCE_ACK_REQUIRED` 阻断链路。
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

### 当前新 VPS（HNCloud，美国 CN2）

- VPS 地址：`177.3.32.116`
- SSH 端口：`20002`
- SSH 用户：`root`
- 本机私钥：`C:\Users\胡文雨\.ssh\sub2api_hncloud_177_3_32_116_ed25519`
- 本机公钥：`C:\Users\胡文雨\.ssh\sub2api_hncloud_177_3_32_116_ed25519.pub`
- 不要把私钥内容写入仓库、日志、Issue、PR 或聊天回复。文档中只允许引用本机路径。

连接命令：

```powershell
ssh -i "$env:USERPROFILE\.ssh\sub2api_hncloud_177_3_32_116_ed25519" -p 20002 root@177.3.32.116
```

当前新 VPS 已在 2026-06-28 完成并行部署验证：

- 系统：Debian GNU/Linux 12，1 vCPU，约 `960MiB` 内存，`50G` 系统盘。
- 已启用 `4G` swap，Docker、Docker Compose v2、Nginx 均已安装并设为开机自启。
- 项目目录：`/opt/sub2api`
- Compose 文件：`/opt/sub2api/docker-compose.yml`
- Compose services：`sub2api`、`postgres`、`redis`
- 当前容器：
  - `sub2api`：镜像 `sub2api:ai-sdk-compat`，端口 `127.0.0.1:18080->8080/tcp`，应为 `healthy`
  - `sub2api-postgres`：镜像 `postgres:18-alpine`，应为 `healthy`
  - `sub2api-redis`：镜像 `redis:8-alpine`，应为 `healthy`
- Nginx 当前监听 `80/443`，公网流量路径为 `Nginx 443/80` -> `http://127.0.0.1:18080` -> 容器 `8080`。
- 当前公网入口：`https://luo-codex.260213.xyz`，Cloudflare DNS 为灰云 DNS-only，A 记录指向 `177.3.32.116`。
- 当前公开设置里的 `api_base_url`、`frontend_url`、`balance_low_notify_recharge_url` 已切到 `luo-codex.260213.xyz`；如果首页 `window.__APP_CONFIG__` 仍出现 `codex.260213.xyz`，通常是应用设置缓存，重启 `sub2api` 应用容器即可刷新，不要重启数据库。
- `http://luo-codex.260213.xyz/health` 应 `301` 跳转到 HTTPS，`https://luo-codex.260213.xyz/health` 应返回 `{"status":"ok"}`。
- Let's Encrypt 证书名：`luo-codex.260213.xyz`，当前为 RSA 证书；后续排障可用 `certbot certificates -d luo-codex.260213.xyz` 复核。
- Nginx 已对 `https://luo-codex.260213.xyz/assets/` 开启 `gzip` 和长期缓存，对 `/logo.png` 开启 7 天缓存；API 和普通页面仍保持 `proxy_buffering off`，避免影响流式接口。
- 如果 443 突然无法访问或 Nginx 无法监听，先用 `ss -lntup | grep ':443'` 和 `docker ps` 检查是否有无关 Docker 服务占用了 443；确认业务归属前不要删除数据卷。

新 VPS 迁移验证快照（2026-06-28）：

- `http://177.3.32.116/health` 返回 `200`。
- `https://luo-codex.260213.xyz/health` 返回 `200`；HTTP 会 `301` 跳转到 HTTPS。
- 本机直连新 VPS 的 20 次 ping 快照：`0%` 丢包，平均约 `146ms`，范围约 `139-185ms`；`/health` HTTPS 总耗时约 `0.43-0.50s`。网页慢时优先区分网络链路、HTML 首包、静态资源压缩/缓存和浏览器缓存。
- `POST http://177.3.32.116/v1/v1/responses` 未带 API key 返回 `401 API_KEY_REQUIRED`，说明兼容兜底路由进入鉴权链路。
- `POST https://luo-codex.260213.xyz/v1/v1/responses` 未带 API key 返回 `401 API_KEY_REQUIRED`，说明 HTTPS 域名入口同样进入鉴权链路。
- Postgres 迁移后关键计数：`74` 张 public 表，`accounts=3`，`api_keys=5`。
- 新 VPS 出站访问 `https://api.openai.com/v1/models` 返回未带认证的 `401`，Anthropic 返回未带 API key 的 `401`，Gemini 返回未带 API key 的 `403`，说明地区和基础网络可进入 API 服务链路。

### 旧 VPS（RackNerd/ColoCrossing，保留待切换）

- VPS 地址：`192.3.89.62`
- SSH 端口：`20002`
- SSH 用户：`root`
- 本机私钥：`C:\Users\胡文雨\.ssh\sub2api_vps_ed25519`
- 本机公钥：`C:\Users\胡文雨\.ssh\sub2api_vps_ed25519.pub`
- 不要把私钥内容写入仓库、日志、Issue、PR 或聊天回复。文档中只允许引用本机路径。

连接命令：

```powershell
ssh -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" -p 20002 root@192.3.89.62
```

代理连接经验：

- 如果直连 SSH 出现 `Connection timed out during banner exchange`、`Timeout, server 192.3.89.62 not responding`，但本地 `ping` 或 `Test-NetConnection -Port 20002` 能通，不要先判断为密钥失效；这更可能是本机直连到 VPS SSH 端口的链路质量差。
- `127.0.0.1:10808` 只是本机历史上常见的代理端口，不要无检查地直接使用。先确认代理是否开启、当前端口是否仍是 `10808`，再决定是否走代理 SSH。
- 远程检查、部署前健康检查、长命令执行和 `scp` 上传，如果已确认本机代理可用，优先通过代理 SSH；如果代理未开启或端口不确定，先直连做轻量测试，仍不稳定时再询问用户当前代理端口。
- 代理 SSH 可使用 Git for Windows 自带的 `connect.exe` 作为 OpenSSH `ProxyCommand`。常见路径：`C:\Program Files\Git\mingw64\bin\connect.exe`。

代理可用性检查：

```powershell
$proxyPort = 10808
Test-NetConnection -ComputerName 127.0.0.1 -Port $proxyPort -InformationLevel Quiet
Get-NetTCPConnection -LocalAddress 127.0.0.1 -LocalPort $proxyPort -State Listen -ErrorAction SilentlyContinue |
  Select-Object LocalAddress,LocalPort,OwningProcess
```

只有上述端口确实在监听时，再使用代理 SSH 示例：

```powershell
$proxyCommand = '"C:\Progra~1\Git\mingw64\bin\connect.exe" -S 127.0.0.1:10808 %h %p'

ssh -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" `
  -p 20002 `
  -o "ProxyCommand=$proxyCommand" `
  root@192.3.89.62
```

上传文件命令：

```powershell
scp -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" -P 20002 `
  <local-file> `
  root@192.3.89.62:/opt/sub2api/backups/
```

确认代理端口监听后，可使用代理上传示例：

```powershell
$proxyCommand = '"C:\Progra~1\Git\mingw64\bin\connect.exe" -S 127.0.0.1:10808 %h %p'

scp -i "$env:USERPROFILE\.ssh\sub2api_vps_ed25519" `
  -P 20002 `
  -o "ProxyCommand=$proxyCommand" `
  <local-file> `
  root@192.3.89.62:/opt/sub2api/backups/
```

## VPS 部署边界

- 除非用户明确要求部署，否则不要推送、发布或切换 VPS。
- 用户曾明确要求过“先合并更新，但是先不要推送到我的 VPS”；后续相同场景默认只合并、验证、推送 GitHub fork，不部署 VPS。
- 用户要求部署前，先检查 VPS 资源情况。部署方式应尽量减少中断影响，缩短中断时长。
- VPS 上的项目目录：`/opt/sub2api`
- Compose 文件：`/opt/sub2api/docker-compose.yml`
- 应用容器：`sub2api`
- Compose 线上镜像标签：`sub2api:ai-sdk-compat`
- 公网流量路径：Nginx `443/80` -> 宿主机本地端口 -> 容器 `8080`
- 常用本机健康检查端口：`http://127.0.0.1:18080/health`
- PostgreSQL 和 Redis 是独立容器；应用镜像更新时只重建 `sub2api`，不要无故重启数据库或 Redis。

## VPS 基础信息快照

以下信息是 2026-06-15 通过只读 SSH 检查得到的现场快照；资源占用、容器状态、Docker 版本和 Nginx 配置以后可能变化。部署、排障或扩容前必须重新核对。

- 主机名：`lfw2400`
- 系统：Debian GNU/Linux 12，Linux `6.1.0-9-amd64`，`x86_64`
- 资源快照：2 vCPU，约 `1.9GiB` 内存；检查时可用内存约 `951MiB`，根分区约 `29G`，可用约 `9.2G`
- Docker：`Docker version 29.1.4`
- Docker Compose：`Docker Compose version v5.0.1`
- Compose services：`sub2api`、`postgres`、`redis`
- 当前容器基线：
  - `sub2api`：镜像 `sub2api:ai-sdk-compat`，端口 `127.0.0.1:18080->8080/tcp`，应为 `healthy`
  - `sub2api-postgres`：镜像 `postgres:18-alpine`，应为 `healthy`
  - `sub2api-redis`：镜像 `redis:8-alpine`，应为 `healthy`
- 项目和数据目录：
  - `/opt/sub2api`
  - `/opt/sub2api/backups`
  - `/opt/sub2api/data`
  - `/opt/sub2api/postgres_data`
  - `/opt/sub2api/redis_data`
- Nginx：`nginx -t` 当前应通过；已观察到的站点包括 `codex.260213.xyz`、`default`、`update.xiaohulp.sbs`，`conf.d` 下有 `sub2api-upgrade.conf`
- `codex.260213.xyz` 当前监听 `80` 和 `443 ssl http2`，反代到 `http://127.0.0.1:18080`；`/assets/` 有单独的静态资源优化配置，曾用于改善前端 chunk 加载速度

部署或排障前建议先跑：

```bash
hostname
uname -srmo
. /etc/os-release && printf '%s %s\n' "$NAME" "$VERSION_ID"
nproc
free -h
df -h / /opt /opt/sub2api
docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
cd /opt/sub2api && docker compose ps
docker inspect sub2api --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'
curl -sS -o /dev/null -w 'health_http=%{http_code} total=%{time_total}\n' http://127.0.0.1:18080/health
nginx -t
```

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

Docker Desktop 代理注意事项：

- Docker daemon 可能不会读取 Windows 用户代理或 Git/PowerShell 代理配置。构建时如果卡在拉取 Docker Hub 基础镜像，常见报错类似 `Docker Desktop has no HTTPS proxy`、`failed to resolve source metadata for docker.io/library/...`、`connectex` 超时。
- 如果只是后端代码小改，且本机已有当前线上运行时镜像 `sub2api:ai-sdk-compat`，可以用备用流程绕过 Docker Hub 拉取：在临时目录从当前提交构建前端 dist 和 Linux 后端二进制，再基于已有 `sub2api:ai-sdk-compat` 只替换 `/app/sub2api` 生成新镜像。
- 备用流程只适合未修改 Dockerfile、基础镜像、系统依赖、运行时依赖和入口脚本的场景。如果这些内容有变化，应先修复 Docker Desktop 代理或基础镜像拉取问题，再走完整 `docker build`。
- PowerShell 下不要优先使用 `docker commit --change 'ENTRYPOINT [...]'` 改入口点，历史上容易因引号传递变成错误的 shell 字符串。备用流程应使用临时 Dockerfile，保留基底镜像的 `ENTRYPOINT` / `CMD`。

备用本机构建示例：

```powershell
$commit = git rev-parse --short HEAD
$version = git describe --tags --always --dirty
$buildDate = (Get-Date).ToUniversalTime().ToString('yyyy-MM-ddTHH:mm:ssZ')
$tmpRoot = Join-Path $env:TEMP "sub2api-build-$commit"
$srcTar = Join-Path $env:TEMP "sub2api-src-$commit.tar"
$tag = "sub2api:<new-image-tag>"

if (Test-Path $tmpRoot) { Remove-Item -LiteralPath $tmpRoot -Recurse -Force }
if (Test-Path $srcTar) { Remove-Item -LiteralPath $srcTar -Force }
New-Item -ItemType Directory -Path $tmpRoot | Out-Null
git archive --format=tar -o $srcTar HEAD
tar -xf $srcTar -C $tmpRoot

Push-Location (Join-Path $tmpRoot 'frontend')
pnpm install --frozen-lockfile
pnpm run build
Pop-Location

Push-Location (Join-Path $tmpRoot 'backend')
$env:CGO_ENABLED='0'
$env:GOOS='linux'
$env:GOARCH='amd64'
go build -tags embed `
  -ldflags "-s -w -X main.Version=$version -X main.Commit=$commit -X main.Date=$buildDate -X main.BuildType=release" `
  -trimpath `
  -o (Join-Path $tmpRoot 'sub2api') `
  ./cmd/server
Pop-Location

$dockerfile = Join-Path $tmpRoot 'Dockerfile.runtime'
@'
FROM sub2api:ai-sdk-compat
USER root
COPY sub2api /app/sub2api
RUN chown sub2api:sub2api /app/sub2api && chmod 755 /app/sub2api
'@ | Set-Content -LiteralPath $dockerfile -Encoding ASCII

docker build -f $dockerfile -t $tag $tmpRoot
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

特定兼容修复的验证经验：

- 对 `/v1/v1/responses` 这类客户端重复拼接 `/v1` 的兜底路由，未带 API key 的验证请求应返回 `401 API_KEY_REQUIRED`，而不是 `404 page not found`。这说明请求已经进入 API 鉴权链路。
- 可同时验证本地和公网路径：

```bash
curl -sS -o /tmp/body.txt -w '%{http_code}\n' \
  -X POST http://127.0.0.1:18080/v1/v1/responses \
  -H 'Content-Type: application/json' \
  --data '{}'

curl -sS -o /tmp/body.txt -w '%{http_code}\n' \
  -X POST https://codex.260213.xyz/v1/v1/responses \
  -H 'Content-Type: application/json' \
  --data '{}'
```
