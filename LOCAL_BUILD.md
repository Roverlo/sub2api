# Sub2API 本机构建记录

本文记录当前 Windows 机器上已经验证过的本机构建方式。目标是以后不要在小内存 VPS 上构建，只在本机生成镜像或产物，再发布到服务器。

## 已验证环境

- Windows PowerShell
- Go `1.26.3`
- Node.js `v24.15.0`
- pnpm 使用项目匹配版本 `9.15.9`
- Docker Desktop `29.4.3`
- Docker Compose `v5.1.3`

如果新开终端后找不到 `go` 或 `make`，先刷新当前 PowerShell 的 PATH：

```powershell
$env:Path = "C:\Program Files\Go\bin;$env:LOCALAPPDATA\Microsoft\WinGet\Links;$env:Path"
```

## 首次准备

Go 代理建议使用国内可访问配置：

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

本机全局 `pnpm` 可能是 winget 安装的新版。这个项目按 Dockerfile 使用 `pnpm 9.15.9`，所以本项目命令优先用 `corepack pnpm`：

```powershell
corepack prepare pnpm@9.15.9 --activate
corepack pnpm --version
```

PowerShell 如果报 `npm.ps1 cannot be loaded`，优先使用 `corepack pnpm` 或 `.cmd` 入口，不需要为此改系统执行策略。

## 安装依赖

在项目根目录执行：

```powershell
corepack pnpm --dir frontend install --frozen-lockfile
```

后端依赖：

```powershell
Push-Location backend
go mod download
Pop-Location
```

## 本地前后端构建验证

前端构建：

```powershell
corepack pnpm --dir frontend run build
```

前端产物会输出到：

```text
backend/internal/web/dist
```

后端 Windows 二进制构建：

```powershell
Push-Location backend
$version = (Get-Content .\cmd\server\VERSION -Raw).Trim()
New-Item -ItemType Directory -Force -Path .\bin | Out-Null
go build -ldflags="-s -w -X main.Version=$version" -trimpath -o .\bin\server.exe .\cmd\server
Pop-Location
```

产物位置：

```text
backend/bin/server.exe
```

## 生产 Docker 镜像构建

默认 Docker Hub 在本机可能会访问 `auth.docker.io` 超时。已验证可用的方式是通过 Dockerfile 的 build arg 切换基础镜像源，不需要修改 Dockerfile：

```powershell
docker build --progress=plain `
  --build-arg NODE_IMAGE=docker.m.daocloud.io/library/node:24-alpine `
  --build-arg GOLANG_IMAGE=docker.m.daocloud.io/library/golang:1.26.3-alpine `
  --build-arg ALPINE_IMAGE=docker.m.daocloud.io/library/alpine:3.21 `
  --build-arg POSTGRES_IMAGE=docker.m.daocloud.io/library/postgres:18-alpine `
  -t sub2api:local .
```

已验证成功产物：

```text
sub2api:local
linux/amd64
约 140MB
```

如果想看构建进度，保留 `--progress=plain`。首次构建会慢，因为要下载 Node、Go、Postgres、Alpine 基础镜像和前端依赖；后续有缓存会明显快很多。

## 可选：保存镜像用于上传服务器

```powershell
docker save sub2api:local -o sub2api-local.tar
```

上传和加载到服务器时，使用实际服务器地址、端口和 SSH key：

```powershell
scp -P <ssh-port> -i <private-key-path> .\sub2api-local.tar root@<server-ip>:/root/sub2api-local.tar
ssh -p <ssh-port> -i <private-key-path> root@<server-ip> "docker load -i /root/sub2api-local.tar"
```

## 注意事项

- 不建议再在小内存 VPS 上执行 Go/前端构建；之前 VPS 出现过 OOM kill。
- 根目录 `Makefile` 和后端 `Makefile` 更偏 Linux/POSIX 环境；Windows 本机优先使用本文中的 PowerShell 命令或 Docker 构建命令。
- 生产镜像构建会自动把前端 dist 嵌入 Go 后端，最终运行用 Docker 镜像即可。
