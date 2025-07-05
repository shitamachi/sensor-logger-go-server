# Docker 部署指南

本项目已配置 Docker 支持，可以通过 Docker 容器运行传感器日志服务器。

## 文件说明

- `Dockerfile` - 多阶段构建配置，创建优化的生产镜像（已针对中国地区配置 Go proxy）
- `docker-compose.yml` - 包含应用程序和 MongoDB 数据库的完整服务栈
- `.dockerignore` - 排除不必要的文件以优化构建过程

## 快速开始

### 使用 Docker Compose（推荐）

1. **构建并启动服务**：
   ```bash
   docker-compose up -d
   ```

2. **查看日志**：
   ```bash
   docker-compose logs -f sensor-logger-server
   ```

3. **停止服务**：
   ```bash
   docker-compose down
   ```

### 仅使用 Docker

1. **构建镜像**：
   ```bash
   docker build -t sensor-logger-server .
   ```

2. **运行容器**：
   ```bash
   docker run -d \
     --name sensor-logger-server \
     -p 18000:18000 \
     -e SERVER_HOST=0.0.0.0 \
     -e MONGO_URI=mongodb://your-mongodb-host:27017 \
     sensor-logger-server
   ```

## 配置选项

### 环境变量

| 变量名 | 默认值 | 描述 |
|--------|--------|------|
| SERVER_PORT | 18000 | 服务器监听端口 |
| SERVER_HOST | 0.0.0.0 | 服务器监听地址 |
| MONGO_URI | mongodb://mongodb:27017 | MongoDB 连接字符串 |
| MONGO_DATABASE | sensor_logger | MongoDB 数据库名 |
| MONGO_TIMEOUT | 10 | MongoDB 连接超时（秒） |
| MAX_DATA_STORE | 1000 | 内存中最大数据存储量 |
| ENABLE_LOGGING | true | 是否启用日志记录 |
| LOG_LEVEL | info | 日志级别 |
| ENVIRONMENT | production | 运行环境 |
| DATA_DIR | /app/data | 数据存储目录 |
| ENABLE_FILE_LOG | true | 是否启用文件日志 |

### 端口映射

- **18000** - 主应用程序端口
- **27017** - MongoDB 数据库端口（仅在使用 docker-compose 时）

### 数据卷

- `sensor_messages` - 应用程序数据存储
- `sensor_logs` - 日志文件存储
- `mongodb_data` - MongoDB 数据存储
- `mongodb_config` - MongoDB 配置存储

## 访问服务

服务启动后，可以通过以下地址访问：

- **主页面**: http://localhost:18000
- **数据仪表板**: http://localhost:18000/dashboard
- **API 端点**:
  - 内存数据: http://localhost:18000/api/data
  - 数据库数据: http://localhost:18000/api/db/data
  - 设备信息: http://localhost:18000/api/db/devices
  - 统计信息: http://localhost:18000/api/db/stats

## 生产部署建议

1. **修改 MongoDB 密码**：
   ```yaml
   environment:
     - MONGO_INITDB_ROOT_USERNAME=your_username
     - MONGO_INITDB_ROOT_PASSWORD=your_secure_password
   ```

2. **使用外部 MongoDB**：
   ```yaml
   environment:
     - MONGO_URI=mongodb://username:password@your-mongodb-host:27017/sensor_logger?authSource=admin
   ```

3. **配置日志级别**：
   ```yaml
   environment:
     - LOG_LEVEL=warn  # 生产环境建议使用 warn 或 error
   ```

4. **数据备份**：
   定期备份 Docker 数据卷中的数据：
   ```bash
   docker run --rm -v sensor-logger-server_mongodb_data:/data -v $(pwd):/backup alpine tar czf /backup/mongodb_backup.tar.gz /data
   ```

## 中国地区优化

Dockerfile 已针对中国地区进行了优化：

1. **Go 代理设置**：
   - 使用 `goproxy.cn` 作为主要代理
   - 备用 `proxy.golang.org` 官方代理
   - 配置 `sum.golang.google.cn` 作为校验服务

2. **自定义 Go 代理**：
   如果需要使用其他代理，可以在构建时覆盖：
   ```bash
   docker build --build-arg GOPROXY=https://mirrors.aliyun.com/goproxy/ -t sensor-logger-server .
   ```

3. **企业内网环境**：
   如果在企业内网环境中，可以设置：
   ```dockerfile
   ENV GOPROXY=https://your-internal-proxy.com,direct
   ENV GOSUMDB=off
   ```

## 故障排除

1. **查看容器状态**：
   ```bash
   docker-compose ps
   ```

2. **查看容器日志**：
   ```bash
   docker-compose logs sensor-logger-server
   docker-compose logs mongodb
   ```

3. **进入容器调试**：
   ```bash
   docker-compose exec sensor-logger-server sh
   ```

4. **重启服务**：
   ```bash
   docker-compose restart sensor-logger-server
   ```

## 健康检查

Dockerfile 包含健康检查配置，会定期检查服务是否正常运行。可以通过以下命令查看健康状态：

```bash
docker inspect --format='{{.State.Health.Status}}' sensor-logger-server
```

## 清理

完全清理所有容器、镜像和数据卷：

```bash
# 停止并删除容器
docker-compose down

# 删除数据卷（注意：这会删除所有数据）
docker-compose down -v

# 删除镜像
docker rmi sensor-logger-server
``` 