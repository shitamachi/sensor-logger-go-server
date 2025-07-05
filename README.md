# 传感器日志服务器

这是一个用Go语言编写的传感器数据接收和解析服务器，专门用于接收和展示来自[Sensor Logger](https://www.tszheichoi.com/sensorlogger)应用的传感器数据。

## 功能特点

✅ **完整的数据解析** - 支持解析所有常见的传感器类型
✅ **人类友好的展示** - 将原始数据转换为易于理解的格式
✅ **实时数据仪表板** - 提供美观的Web界面查看传感器数据
✅ **数据持久化** - 自动保存所有接收到的数据为JSON文件
✅ **MongoDB支持** - 支持将数据存储到MongoDB数据库
✅ **多传感器支持** - 支持加速度计、陀螺仪、磁力计等多种传感器
✅ **中文界面** - 完全中文化的用户界面
✅ **RESTful API** - 提供完整的API接口访问数据

## 支持的传感器类型

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

## 快速开始

### 1. 配置应用（可选）

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
| `MONGO_URI` | mongodb://localhost:27017 | MongoDB连接URI |
| `MONGO_DATABASE` | sensor_logger | MongoDB数据库名称 |
| `MONGO_TIMEOUT` | 10 | MongoDB连接超时（秒） |
| `MAX_DATA_STORE` | 100 | 内存中最大数据存储条数 |
| `ENABLE_LOGGING` | true | 是否启用日志输出 |
| `LOG_LEVEL` | info | 日志级别 (debug/info/warn/error) |
| `DATA_DIR` | ./data | 数据文件存储目录 |
| `ENABLE_FILE_LOG` | true | 是否启用文件日志 |

### 2. 安装MongoDB（可选）

如果要使用MongoDB数据库存储功能：

**Windows:**
1. 下载并安装[MongoDB Community Server](https://www.mongodb.com/try/download/community)
2. 启动MongoDB服务：`net start MongoDB`

**Linux/macOS:**
```bash
# Ubuntu/Debian
sudo apt-get install mongodb

# macOS (使用Homebrew)
brew install mongodb-community

# 启动MongoDB
sudo systemctl start mongod  # Linux
brew services start mongodb-community  # macOS
```

**Docker:**
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### 3. 启动服务器

```bash
go run main.go
```

服务器将在端口 18000 上启动，你会看到类似以下的输出：

```
=== 传感器日志服务器 ===
服务器启动在端口 18000

本机IP地址:
  192.168.1.100 (以太网)
  192.168.1.101 (Wi-Fi)

请在Sensor Logger应用中设置推送URL为: http://[你的IP地址]:18000/data
使用 'Tap to Test Pushing' 按钮测试连接
访问 http://[你的IP地址]:18000/dashboard 查看数据仪表板
========================
```

### 4. 配置Sensor Logger应用

1. 在手机上打开Sensor Logger应用
2. 进入设置页面（点击齿轮图标）
3. 找到"推送URL"设置
4. 输入：`http://[你的服务器IP]:18000/data`
5. 点击"Tap to Test Pushing"按钮测试连接

### 5. 查看数据

- **主页**: `http://[你的IP]:18000/` - 服务器状态和配置信息
- **数据仪表板**: `http://[你的IP]:18000/dashboard` - 实时传感器数据展示
- **API接口**: `http://[你的IP]:18000/api/data` - 获取JSON格式的所有数据

## 数据展示示例

### 加速度计数据
```
传感器类型: accelerometer
时间: 2025-07-05 15:30:25.123
精度: 高精度

X轴加速度: -0.032849 m/s² (X轴方向的加速度)
Y轴加速度: -0.004899 m/s² (Y轴方向的加速度)
Z轴加速度: 0.089095 m/s² (Z轴方向的加速度)
```

### 陀螺仪数据
```
传感器类型: gyroscope
时间: 2025-07-05 15:30:25.147
精度: 高精度

X轴角速度: -0.051131 rad/s (绕X轴的角速度)
Y轴角速度: -0.031957 rad/s (绕Y轴的角速度)
Z轴角速度: 0.016911 rad/s (绕Z轴的角速度)
```

### 磁力计数据
```
传感器类型: magnetometer
时间: 2025-07-05 15:30:25.186
精度: 高精度

磁方位角: 137.28 度 (相对于磁北的方位角)
```

## 文件结构

```
sensor-logger-server/
├── main.go                          # 主程序入口
├── types.go                         # 数据结构定义
├── config.go                        # 配置管理
├── parser.go                        # 传感器数据解析
├── handlers.go                      # HTTP处理程序
├── utils.go                         # 工具函数
├── main_test.go                     # 核心功能测试
├── handlers_test.go                 # HTTP处理程序测试
├── config_test.go                   # 配置功能测试
├── database.go                      # MongoDB数据库操作
├── database_test.go                 # 数据库功能测试
├── go.mod                           # Go模块文件
├── README.md                        # 说明文档
├── env.example                      # 配置文件模板
├── .env                             # 配置文件（可选）
├── run_tests.bat                    # 测试运行脚本
├── data/                            # 数据存储目录
│   └── sensor_data_*.json           # 传感器数据文件
└── temp/                            # 测试数据目录
    └── sensor_data_*.json           # 测试用传感器数据
```

## API接口

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

## 数据解析特性

### 时间戳处理
- 自动将纳秒时间戳转换为可读的日期时间格式
- 支持时区转换和本地化显示

### 精度标识
- 0: 不可靠
- 1: 低精度
- 2: 中等精度
- 3: 高精度

### 单位转换
所有数据都包含适当的单位标识：
- 加速度: m/s²
- 角速度: rad/s
- 磁场强度: μT
- 角度: 度
- 距离: 米
- 压力: hPa

## 数据存储

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

## 技术特点

- **高性能**: 使用Go语言编写，处理速度快
- **并发安全**: 支持多个设备同时发送数据
- **容错性强**: 即使数据解析失败也不会影响数据接收
- **扩展性好**: 易于添加新的传感器类型支持
- **跨平台**: 支持Windows、Linux、macOS

## 故障排除

### 连接问题
1. 确保手机和服务器在同一网络中
2. 检查防火墙设置，确保端口18000开放
3. 验证IP地址是否正确

### 数据不显示
1. 检查Sensor Logger应用是否正在发送数据
2. 查看服务器控制台输出是否有错误信息
3. 尝试刷新仪表板页面

### 性能优化
- 服务器会自动限制内存中存储的数据量（最多100条记录）
- 所有历史数据都保存在JSON文件中，可以离线分析

## 参考资料

- [Sensor Logger官方网站](https://www.tszheichoi.com/sensorlogger)
- [Awesome Sensor Logger项目](https://github.com/tszheichoi/awesome-sensor-logger)
- [传感器数据格式文档](https://github.com/tszheichoi/awesome-sensor-logger/blob/main/UNITS.md)

## 测试

项目包含完整的测试套件，涵盖了核心功能和HTTP处理程序。

### 运行测试

```bash
# 运行所有测试
go test -v

# 运行测试覆盖率分析
go test -cover

# 运行基准测试
go test -bench=. -benchmem

# 运行竞态检测
go test -race

# Windows用户可以使用批处理脚本
run_tests.bat
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

### 测试文件

- `main_test.go` - 核心功能测试
- `handlers_test.go` - HTTP处理程序测试
- `config_test.go` - 配置功能测试
- `database_test.go` - 数据库功能测试
- `temp/` - 真实传感器数据文件用于测试

### 基准测试结果

```
BenchmarkParseSensorMessage-16    181866    6088 ns/op    1705 B/op    37 allocs/op
```

- 每次解析操作耗时约 6 微秒
- 每次操作分配约 1.7KB 内存
- 内存分配次数为 37 次

## 许可证

本项目基于MIT许可证开源。 