# 构建说明

本项目提供了多种构建方式，支持不同的操作系统和开发环境。

## 构建工具

### Linux/macOS - 使用 Makefile

项目包含一个功能完整的 `Makefile`，支持以下操作：

```bash
# 查看所有可用命令
make help

# 构建应用程序
make build

# 构建多平台版本
make build-all          # 构建所有平台
make build-linux        # 构建Linux版本
make build-windows      # 构建Windows版本
make build-darwin       # 构建macOS版本

# 测试
make test               # 运行测试
make test-coverage      # 运行测试并生成覆盖率报告
make test-race          # 运行竞态检测测试
make benchmark          # 运行性能测试

# 代码质量
make fmt                # 格式化代码
make vet                # 运行go vet检查
make lint               # 运行golangci-lint检查
make check              # 运行所有代码检查

# 运行
make run                # 运行应用程序
make run-dev            # 以开发模式运行

# Docker
make docker-build       # 构建Docker镜像
make docker-run         # 运行Docker容器
make compose-up         # 启动Docker Compose服务
make compose-down       # 停止Docker Compose服务

# 清理
make clean              # 清理构建文件
make clean-data         # 清理数据文件

# 开发工具
make dev-setup          # 设置开发环境
make dev-watch          # 热重载开发

# 发布
make release            # 准备发布版本
make info               # 显示项目信息
```

### Windows - 使用批处理文件

Windows 用户可以使用 `make.bat` 批处理文件：

```cmd
# 查看所有可用命令
make.bat help

# 构建应用程序
make.bat build

# 构建多平台版本
make.bat build-all
make.bat build-windows
make.bat build-linux

# 测试
make.bat test
make.bat test-coverage

# 运行
make.bat run
make.bat run-dev

# Docker
make.bat docker-build
make.bat docker-run
make.bat compose-up
make.bat compose-down

# 清理
make.bat clean

# 项目信息
make.bat info
```

## 版本信息注入

构建时会自动注入以下版本信息：

- **Version**: 版本号（可通过 `VERSION` 环境变量设置）
- **BuildTime**: 构建时间
- **GitCommit**: Git提交哈希
- **GitBranch**: Git分支名

### 设置版本号

```bash
# Linux/macOS
make build VERSION=v1.2.3

# Windows
set VERSION=v1.2.3 && make.bat build
```

### 查看版本信息

运行应用程序时会显示版本信息：

```
=== 传感器日志服务器 ===
版本: v1.0.0
构建时间: 2025-07-06T02:45:07Z
Git提交: ab780df
Git分支: main
```

## Docker 构建

### 使用 Makefile 构建

Dockerfile 已更新为使用 Makefile 进行构建：

```dockerfile
# 使用 Makefile 构建应用程序
RUN make build-linux VERSION=${VERSION}
```

### 构建 Docker 镜像

```bash
# 基本构建
docker build -t sensor-logger-server:latest .

# 指定版本
docker build --build-arg VERSION=v1.2.3 -t sensor-logger-server:v1.2.3 .

# 使用 Makefile
make docker-build

# 使用 Docker Compose
docker-compose build
```

### 开发环境

项目提供了开发环境的 Docker Compose 配置：

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f
```

## 开发工具

### 热重载开发

安装 Air 进行热重载开发：

```bash
# 安装开发工具
make dev-setup

# 启动热重载
make dev-watch
```

或者手动安装：

```bash
go install github.com/cosmtrek/air@latest
air
```

### 代码质量工具

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
make lint
```

## 环境配置

### 开发环境

```bash
# 设置环境变量
export ENVIRONMENT=dev

# 或者使用 make 命令
make run-dev
```

### 生产环境

```bash
# 设置环境变量
export ENVIRONMENT=production

# 构建生产版本
make build VERSION=v1.0.0
```

## 测试

### 运行测试

```bash
# 基本测试
make test

# 覆盖率测试
make test-coverage

# 竞态检测
make test-race

# 性能测试
make benchmark
```

### 查看覆盖率报告

测试完成后会生成 `coverage.html` 文件，可以在浏览器中查看详细的覆盖率报告。

## 部署

### 系统安装

```bash
# 安装到系统（需要 sudo）
make install

# 卸载
make uninstall
```

### Docker 部署

```bash
# 生产环境
docker-compose up -d

# 开发环境
docker-compose -f docker-compose.dev.yml up -d
```

## 故障排除

### Windows 环境

如果在 Windows 上遇到编码问题，可以：

1. 使用 PowerShell 而不是 CMD
2. 设置正确的代码页：`chcp 65001`
3. 使用 Git Bash 或 WSL

### 构建失败

如果构建失败，请检查：

1. Go 版本是否为 1.21+
2. 依赖是否正确下载：`go mod download`
3. 代码是否通过检查：`make check`

### Docker 构建失败

如果 Docker 构建失败：

1. 确保 Docker 版本支持多阶段构建
2. 检查网络连接（Go 代理设置）
3. 清理 Docker 缓存：`docker system prune` 