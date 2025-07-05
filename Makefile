# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# 应用信息
BINARY_NAME=sensor-logger-server
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# 版本信息
VERSION?=v1.0.0
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

# 构建标志
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.GitBranch=$(GIT_BRANCH)"

# Docker相关
DOCKER_IMAGE=sensor-logger-server
DOCKER_TAG?=latest
DOCKER_REGISTRY?=localhost:5000

# 测试覆盖率
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# 默认目标
.DEFAULT_GOAL := help

# 显示帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "可用的make命令："
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 构建相关
.PHONY: build
build: ## 构建应用程序
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v .

.PHONY: build-linux
build-linux: ## 构建Linux版本
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) -v .

.PHONY: build-windows
build-windows: ## 构建Windows版本
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) -v .

.PHONY: build-darwin
build-darwin: ## 构建macOS版本
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DARWIN) -v .

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## 构建所有平台版本

# 测试相关
.PHONY: test
test: ## 运行测试
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "覆盖率报告已生成: $(COVERAGE_HTML)"

.PHONY: test-race
test-race: ## 运行竞态检测测试
	$(GOTEST) -v -race ./...

.PHONY: benchmark
benchmark: ## 运行性能测试
	$(GOTEST) -bench=. -benchmem ./...

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	$(GOFMT) ./...

.PHONY: vet
vet: ## 运行go vet检查
	$(GOCMD) vet ./...

.PHONY: lint
lint: ## 运行golangci-lint检查
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint未安装，跳过lint检查"; \
	fi

.PHONY: check
check: fmt vet lint test ## 运行所有代码检查

# 依赖管理
.PHONY: deps
deps: ## 下载依赖
	$(GOMOD) download

.PHONY: deps-update
deps-update: ## 更新依赖
	$(GOMOD) tidy
	$(GOGET) -u ./...

.PHONY: deps-verify
deps-verify: ## 验证依赖
	$(GOMOD) verify

# 清理
.PHONY: clean
clean: ## 清理构建文件
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_DARWIN)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

.PHONY: clean-data
clean-data: ## 清理数据文件
	rm -rf data/
	rm -rf temp/
	rm -f sensor_messages_*.json

# 运行
.PHONY: run
run: ## 运行应用程序
	$(GOCMD) run .

.PHONY: run-dev
run-dev: ## 以开发模式运行应用程序
	ENVIRONMENT=dev $(GOCMD) run .

# Docker相关
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-build-no-cache
docker-build-no-cache: ## 无缓存构建Docker镜像
	docker build --no-cache -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## 运行Docker容器
	docker run -p 18000:18000 --name $(DOCKER_IMAGE) $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-run-bg
docker-run-bg: ## 后台运行Docker容器
	docker run -d -p 18000:18000 --name $(DOCKER_IMAGE) $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: ## 停止Docker容器
	docker stop $(DOCKER_IMAGE) || true

.PHONY: docker-clean
docker-clean: ## 清理Docker容器和镜像
	docker stop $(DOCKER_IMAGE) || true
	docker rm $(DOCKER_IMAGE) || true
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true

.PHONY: docker-push
docker-push: ## 推送Docker镜像
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose相关
.PHONY: compose-up
compose-up: ## 启动Docker Compose服务
	docker-compose up -d

.PHONY: compose-down
compose-down: ## 停止Docker Compose服务
	docker-compose down

.PHONY: compose-logs
compose-logs: ## 查看Docker Compose日志
	docker-compose logs -f

.PHONY: compose-build
compose-build: ## 构建Docker Compose服务
	docker-compose build

.PHONY: compose-restart
compose-restart: ## 重启Docker Compose服务
	docker-compose restart

# 部署相关
.PHONY: install
install: build ## 安装应用程序到系统
	sudo cp $(BINARY_NAME) /usr/local/bin/

.PHONY: uninstall
uninstall: ## 从系统卸载应用程序
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# 开发工具
.PHONY: dev-setup
dev-setup: ## 设置开发环境
	$(GOGET) -u golang.org/x/tools/cmd/goimports
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -u github.com/cosmtrek/air@latest

.PHONY: dev-watch
dev-watch: ## 使用air进行热重载开发
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air未安装，请先运行 make dev-setup"; \
	fi

# 发布相关
.PHONY: release
release: clean build-all test-coverage ## 准备发布版本
	@echo "版本 $(VERSION) 已准备发布"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Git分支: $(GIT_BRANCH)"

# 信息显示
.PHONY: info
info: ## 显示项目信息
	@echo "项目信息:"
	@echo "  名称: $(BINARY_NAME)"
	@echo "  版本: $(VERSION)"
	@echo "  构建时间: $(BUILD_TIME)"
	@echo "  Git提交: $(GIT_COMMIT)"
	@echo "  Git分支: $(GIT_BRANCH)"
	@echo "  Docker镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)" 