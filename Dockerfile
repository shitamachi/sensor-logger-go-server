# 使用官方 Go 镜像作为构建阶段
FROM golang:1.24.4-alpine AS builder

# 构建参数
ARG VERSION=v1.0.0

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git ca-certificates tzdata make

# 设置 Go 代理（中国地区优化）
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.google.cn

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 使用 Makefile 构建应用程序
RUN make build-linux VERSION=${VERSION}

# 使用最小的基础镜像作为运行阶段
FROM alpine:latest

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/sensor-logger-server_unix ./sensor-logger-server

# 创建必要的目录
RUN mkdir -p data/logs temp && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 18000

# 设置环境变量
ENV SERVER_PORT=18000
ENV SERVER_HOST=0.0.0.0
ENV ENVIRONMENT=production
ENV DATA_DIR=/app/data
ENV ENABLE_FILE_LOG=true
ENV LOG_LEVEL=info

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:18000/ || exit 1

# 启动命令
CMD ["./sensor-logger-server"] 