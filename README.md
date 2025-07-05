# 传感器日志服务器

这是一个用Go语言编写的传感器数据接收和解析服务器，专门用于接收和展示来自[Sensor Logger](https://www.tszheichoi.com/sensorlogger)应用的传感器数据。

## 🎯 项目概述

本项目成功实现了一个完整的传感器数据接收和解析服务器，可以接收来自Sensor Logger应用的传感器数据，并将其转换为人类友好的格式进行展示。

## ✅ 功能特点

- ✅ **完整的数据解析** - 支持解析所有常见的传感器类型
- ✅ **人类友好的展示** - 将原始数据转换为易于理解的格式
- ✅ **实时数据仪表板** - 提供美观的Web界面查看传感器数据
- ✅ **数据持久化** - 自动保存所有接收到的数据为JSON文件
- ✅ **MongoDB支持** - 支持将数据存储到MongoDB数据库
- ✅ **多传感器支持** - 支持加速度计、陀螺仪、磁力计等多种传感器
- ✅ **中文界面** - 完全中文化的用户界面
- ✅ **RESTful API** - 提供完整的API接口访问数据
- ✅ **结构化日志** - 使用Go 1.21+ slog结构化日志系统
- ✅ **环境配置** - 支持开发/生产环境区分
- ✅ **Docker支持** - 提供完整的容器化部署方案
- ✅ **构建系统** - 支持多平台构建和版本管理

## 🚀 快速开始

### 使用构建工具

项目提供了完整的构建系统，支持多种操作系统：

**Linux/macOS:**
```bash
# 查看所有可用命令
make help

# 构建应用程序
make build

# 运行应用程序
make run

# 运行测试
make test
```

**Windows:**
```cmd
# 查看所有可用命令
make.bat help

# 构建应用程序
make.bat build

# 运行应用程序
make.bat run

# 运行测试
make.bat test
```

### 传统方式

```bash
# 直接运行
go run main.go

# 构建后运行
go build -o sensor-logger-server .
./sensor-logger-server
```

## 📊 支持的传感器类型

| 传感器类型 | 描述 | 数据示例 |
|-----------|------|----------|
| 加速度计 (accelerometer) | 测量设备在三个轴上的加速度 | X/Y/Z轴加速度 (m/s²) |
| 陀螺仪 (gyroscope) | 测量设备的角速度 | X/Y/Z轴角速度 (rad/s) |
| 磁力计 (magnetometer) | 测量磁场强度和方向 | 磁方位角 (度) 或 X/Y/Z轴磁场 (μT) |
| 重力传感器 (gravity) | 测量重力矢量 | X/Y/Z轴重力分量 (m/s²) |
| 方向传感器 (orientation) | 设备方向四元数 | 四元数 W/X/Y/Z 分量 |
| 指南针 (compass) | 指南针方位 | 磁方位角 (度) |
| 计步器 (pedometer) | 步数统计 | 累计步数 |
| 未校准磁力计 (magnetometeruncalibrated) | 原始磁场数据 | X/Y/Z轴未校准磁场 (μT) |
| 位置 (location) | GPS位置信息 | 经纬度、海拔、速度等 |
| 气压计 (barometer) | 大气压力 | 气压 (hPa)、气压高度 (米) |

## 🔧 配置说明

### 环境配置

项目支持通过`.env`文件进行配置：

```bash
# 复制配置文件模板
cp env.example .env

# 编辑配置文件
# 修改 .env 文件中的配置项
```

**配置选项说明：**

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_PORT` | 18000 | 服务器端口 |
| `SERVER_HOST` | (空) | 服务器主机，空表示监听所有接口 |
| `ENVIRONMENT` | dev | 运行环境 (dev/development/prod/production) |
| `LOG_LEVEL` | info | 日志级别 (debug/info/warn/error) |
| `ENABLE_FILE_LOG` | true | 是否启用文件日志 |
| `DATA_DIR` | ./data | 数据文件存储目录 |
| `MAX_DATA_STORE` | 100 | 内存中最大数据存储条数 |
| `MONGO_URI` | mongodb://localhost:27017 | MongoDB连接URI |
| `MONGO_DATABASE` | sensor_logger | MongoDB数据库名称 |
| `MONGO_TIMEOUT` | 10 | MongoDB连接超时（秒） |

### 日志系统

项目使用Go 1.21+的结构化日志系统，支持：

**开发环境:**
- 文本格式输出，易于阅读
- 包含源码文件和行号信息
- 人类友好的时间格式

**生产环境:**
- JSON格式输出，便于日志收集
- 标准RFC3339时间格式
- 结构化字段便于查询

## 🐳 Docker部署

### 快速部署

```bash
# 使用Docker Compose（推荐）
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 开发环境

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f
```

### 手动Docker构建

```bash
# 构建镜像
docker build -t sensor-logger-server:latest .

# 运行容器
docker run -p 18000:18000 sensor-logger-server:latest
```

## 📱 配置Sensor Logger应用

1. 在手机上打开Sensor Logger应用
2. 进入设置页面（点击齿轮图标）
3. 找到"推送URL"设置
4. 输入：`http://[你的服务器IP]:18000/data`
5. 点击"Tap to Test Pushing"按钮测试连接

## 🌐 Web界面和API

### 访问地址

- **主页**: `http://[你的IP]:18000/` - 服务器状态和配置信息
- **数据仪表板**: `http://[你的IP]:18000/dashboard` - 实时传感器数据展示
- **API接口**: `http://[你的IP]:18000/api/data` - 获取JSON格式的所有数据

### 数据展示示例

#### 加速度计数据
```
传感器类型: accelerometer
时间: 2025-07-05 15:30:25.123
精度: 高精度

X轴加速度: -0.032849 m/s² (X轴方向的加速度)
Y轴加速度: -0.004899 m/s² (Y轴方向的加速度)
Z轴加速度: 0.089095 m/s² (Z轴方向的加速度)
```

#### 陀螺仪数据
```
传感器类型: gyroscope
时间: 2025-07-05 15:30:25.147
精度: 高精度

X轴角速度: -0.051131 rad/s (绕X轴的角速度)
Y轴角速度: -0.031957 rad/s (绕Y轴的角速度)
Z轴角速度: 0.016911 rad/s (绕Z轴的角速度)
```

#### 磁力计数据
```
传感器类型: magnetometer
时间: 2025-07-05 15:30:25.186
精度: 高精度

磁方位角: 137.28 度 (相对于磁北的方位角)
```

#### 位置数据
```
传感器类型: location
时间: 2025-07-05 15:30:25.200
精度: 高精度

纬度: 39.90419800 度 (地理纬度)
经度: 116.40739600 度 (地理经度)
海拔: 43.20 米 (海拔高度)
速度: 0.00 m/s (移动速度)
方位角: 0.00 度 (移动方位角)
```

## 📁 项目结构

```
sensor-logger-server/
├── main.go                          # 主程序入口
├── types.go                         # 数据结构定义
├── config.go                        # 配置管理
├── parser.go                        # 传感器数据解析
├── handlers.go                      # HTTP处理程序
├── utils.go                         # 工具函数
├── logger.go                        # 日志系统
├── database.go                      # MongoDB数据库操作
├── *_test.go                        # 测试文件
├── Makefile                         # 构建脚本（Linux/macOS）
├── make.bat                         # 构建脚本（Windows）
├── .air.toml                        # 热重载配置
├── Dockerfile                       # Docker构建文件
├── docker-compose.yml               # Docker Compose配置
├── docker-compose.dev.yml           # 开发环境配置
├── go.mod                           # Go模块文件
├── README.md                        # 说明文档
├── BUILD.md                         # 构建说明文档
├── env.example                      # 配置文件模板
├── .gitignore                       # Git忽略文件
├── .gitattributes                   # Git属性文件
├── data/                            # 数据存储目录
│   ├── logs/                        # 日志文件目录
│   └── sensor_data_*.json           # 传感器数据文件
└── temp/                            # 测试数据目录
    └── sensor_data_*.json           # 测试用传感器数据
```

## 🔌 API接口

### POST /data
接收来自Sensor Logger应用的传感器数据。

**请求格式:**
```json
{
    "messageId": 27,
    "sessionId": "f08abd14-4a86-4913-8c33-f75269e862b9",
    "deviceId": "0e35011f-e2fe-482e-b9a1-1625bb039f37",
    "payload": [
        {
            "name": "accelerometer",
            "values": {
                "x": -0.032849,
                "y": -0.004899,
                "z": 0.089095
            },
            "accuracy": 3,
            "time": 1751729987417535200
        }
    ]
}
```

### GET /dashboard
显示传感器数据仪表板，包含：
- 统计信息（总消息数、总读数、传感器类型、设备数量）
- 最新传感器数据的人类友好展示
- 自动刷新功能

### GET /api/data
返回内存中的解析后传感器数据（JSON格式）。

### GET /api/db/data
从MongoDB数据库获取传感器数据。

**查询参数:**
- `limit`: 限制返回的记录数量（默认50）
- `device`: 按设备ID过滤
- `sensor`: 按传感器类型过滤

**示例:**
```
GET /api/db/data?limit=100&device=test-device&sensor=accelerometer
```

### GET /api/db/devices
获取所有设备信息，包括：
- 设备ID
- 首次和最后访问时间
- 总记录数
- 支持的传感器类型
- 会话列表

### GET /api/db/stats
获取数据库统计信息，包括：
- 总记录数
- 设备数量
- 传感器类型数量
- 最新数据时间

## 🏗️ 技术架构

### 数据流程
1. **接收**: Sensor Logger应用通过HTTP POST发送JSON数据到`/data`端点
2. **解析**: 服务器解析JSON数据，提取传感器信息
3. **转换**: 将原始数据转换为人类可读格式，添加单位和描述
4. **存储**: 原始数据保存为文件，解析后数据存储在内存中
5. **展示**: 通过Web界面和API提供数据访问

### 核心组件
- **SensorMessage**: 完整消息结构体
- **SensorReading**: 单个传感器读数结构体
- **ParsedSensorData**: 解析后的数据结构体
- **HumanReadableSensorData**: 人类可读的传感器数据结构体

### 数据解析特性

#### 时间戳处理
- 自动将纳秒时间戳转换为可读的日期时间格式
- 支持时区转换和本地化显示

#### 精度标识
- 0: 不可靠
- 1: 低精度
- 2: 中等精度
- 3: 高精度

#### 单位转换
所有数据都包含适当的单位标识：
- 加速度: m/s²
- 角速度: rad/s
- 磁场强度: μT
- 角度: 度
- 距离: 米
- 压力: hPa

## 💾 数据存储

### 文件存储
- 所有接收到的原始数据都会保存为JSON文件（可通过`ENABLE_FILE_LOG`配置）
- 文件存储在`DATA_DIR`目录中（默认为`./data`）
- 文件命名格式: `sensor_data_YYYYMMDD_HHMMSS.json`

### 内存存储
- 解析后的数据存储在内存中，支持最近N条记录的快速访问（可通过`MAX_DATA_STORE`配置，默认100条）
- 用于快速响应API请求和仪表板显示

### MongoDB存储
- 支持将数据自动存储到MongoDB数据库（通过`MONGO_URI`等配置项设置）
- 数据库结构：
  - `sensor_data` 集合：存储传感器读数数据
  - `device_info` 集合：存储设备信息和统计数据
- 自动创建索引以优化查询性能
- 支持设备信息的自动更新和统计

## 🧪 测试

### 运行测试

```bash
# 基本测试
make test           # Linux/macOS
make.bat test       # Windows

# 覆盖率测试
make test-coverage  # Linux/macOS
make.bat test-coverage  # Windows

# 竞态检测
make test-race      # Linux/macOS

# 性能测试
make benchmark      # Linux/macOS
```

### 测试覆盖率

当前测试覆盖率为 **68.4%**，包括：

- ✅ 传感器数据解析功能
- ✅ 人类可读格式转换
- ✅ HTTP处理程序
- ✅ 并发处理
- ✅ 错误处理
- ✅ 真实数据文件解析
- ✅ 配置管理功能
- ✅ 数据库操作功能
- ✅ 日志系统功能

### 基准测试结果

```
BenchmarkParseSensorMessage-16    181866    6088 ns/op    1705 B/op    37 allocs/op
```

- 每次解析操作耗时约 6 微秒
- 每次操作分配约 1.7KB 内存
- 内存分配次数为 37 次

## 🚀 GitHub Actions CI/CD

项目配置了完整的GitHub Actions CI/CD工作流，支持自动化测试和Docker镜像构建推送。

### 📋 工作流概述

CI/CD工作流包含以下阶段：
1. **🧪 测试阶段** - 运行Go测试套件，代码质量检查
2. **🐳 构建和推送** - 构建Docker镜像并推送到私有镜像仓库

### 🎯 触发条件

- **Push到主分支** (`main`, `master`)
- **创建标签** (格式：`v*`，如 `v1.0.0`)
- **Pull Request** 到主分支（仅运行测试，不构建推送）
- **手动触发** - 在GitHub Actions页面点击 "Run workflow" 按钮

### 🔐 配置要求

#### 必需的GitHub Secrets

在GitHub仓库的 **Settings > Secrets and variables > Actions** 中配置：

```
DOCKER_USERNAME     # Docker镜像仓库用户名
DOCKER_PASSWORD     # Docker镜像仓库密码或访问令牌
```

#### 可选的GitHub Variables

```
DOCKER_REGISTRY     # 私有镜像仓库地址（如：harbor.company.com）
```

如果不设置，将使用默认值 `your-private-registry.com`

### 🏷️ 镜像标签策略

- **标签推送**: `your-registry.com/sensor-logger-server:v1.0.0`
- **分支推送**: `your-registry.com/sensor-logger-server:v2024.07.06-1a2b3c4`
- **主分支**: 额外添加 `latest` 标签

### 🔧 支持的镜像仓库

- **Harbor私有仓库**: `harbor.company.com`
- **阿里云容器镜像服务**: `registry.cn-hangzhou.aliyuncs.com`
- **腾讯云容器镜像服务**: `ccr.ccs.tencentyun.com`
- **华为云容器镜像服务**: `swr.cn-north-1.myhuaweicloud.com`

### 📊 监控和日志

- 在GitHub Actions页面查看工作流运行状态
- 查看详细日志和测试结果
- 获取构建版本和镜像标签信息

### 🔒 安全最佳实践

1. **定期轮换密钥**：定期更新Docker密码
2. **最小权限原则**：只给必要的权限
3. **使用专用账户**：为CI/CD创建专用的Docker账户
4. **监控访问日志**：定期检查镜像仓库的访问日志
5. **环境隔离**：生产环境使用独立的镜像仓库

### 🔍 故障排除

#### 常见问题：

**Docker登录失败**
- 检查 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` 是否正确设置

**镜像拉取失败**
- 确认镜像仓库地址正确
- 检查Docker登录凭据是否有效
- 确认镜像仓库权限设置

**测试失败**
- 检查代码是否有语法错误
- 确保所有依赖正确安装
- 查看测试日志获取详细错误信息

详细配置说明请参考 [GitHub Actions设置文档](GITHUB_ACTIONS_SETUP.md)。

## 🎨 界面特色

### 响应式设计
- 适配桌面和移动设备
- 现代化的卡片布局
- 清晰的数据分组

### 实时更新
- 30秒自动刷新
- 实时统计信息
- 最新数据优先显示

### 用户友好
- 完全中文界面
- 直观的数据展示
- 详细的使用说明

## 🔧 开发工具

### 热重载开发

```bash
# 安装开发工具
make dev-setup

# 启动热重载
make dev-watch

# 或者手动安装Air
go install github.com/cosmtrek/air@latest
air
```

### 代码质量

```bash
# 格式化代码
make fmt

# 静态检查
make vet

# Lint检查
make lint

# 运行所有检查
make check
```

## 📦 构建和部署

### 版本信息

构建时会自动注入版本信息：
- Version: 版本号
- BuildTime: 构建时间
- GitCommit: Git提交哈希
- GitBranch: Git分支名

### 多平台构建

```bash
# 构建当前平台
make build

# 构建所有平台
make build-all

# 构建特定平台
make build-linux
make build-windows
make build-darwin
```

### 系统安装

```bash
# 安装到系统（Linux/macOS）
make install

# 卸载
make uninstall
```

## 🚨 故障排除

### 连接问题
1. 确保手机和服务器在同一网络中
2. 检查防火墙设置，确保端口18000开放
3. 验证IP地址是否正确

### 数据不显示
1. 检查Sensor Logger应用是否正在发送数据
2. 查看服务器控制台输出是否有错误信息
3. 尝试刷新仪表板页面

### Windows环境问题
1. 使用PowerShell而不是CMD
2. 设置正确的代码页：`chcp 65001`
3. 使用Git Bash或WSL

### 构建失败
1. 检查Go版本是否为1.21+
2. 确保依赖正确下载：`go mod download`
3. 运行代码检查：`make check`

## 🎯 项目亮点

1. **完整性**: 支持Sensor Logger应用的所有常见传感器类型
2. **易用性**: 中文界面，人类友好的数据展示
3. **实时性**: 实时接收和展示传感器数据
4. **可扩展性**: 易于添加新的传感器类型支持
5. **稳定性**: 容错设计，即使数据解析失败也不影响接收
6. **性能**: 高效的Go语言实现，支持并发处理
7. **现代化**: 使用Go 1.21+特性，结构化日志，完整的构建系统
8. **部署友好**: 支持Docker容器化部署，多环境配置

## 📚 参考资料

- [Sensor Logger官方网站](https://www.tszheichoi.com/sensorlogger)
- [Awesome Sensor Logger项目](https://github.com/tszheichoi/awesome-sensor-logger)
- [传感器数据格式文档](https://github.com/tszheichoi/awesome-sensor-logger/blob/main/UNITS.md)
- [构建说明文档](BUILD.md)

## 📄 许可证

本项目基于MIT许可证开源。 