package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// 全局变量用于存储解析后的数据
var parsedDataStore = NewThreadSafeDataStore()

// handleRoot 处理根路径请求
func handleRoot(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>传感器日志服务器</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
            border-bottom: 3px solid #4CAF50;
            padding-bottom: 10px;
        }
        .status {
            background-color: #e8f5e8;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
            border-left: 4px solid #4CAF50;
        }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .info-card {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border: 1px solid #e9ecef;
        }
        .info-card h3 {
            margin-top: 0;
            color: #495057;
            font-size: 1.1em;
        }
        .info-card p {
            margin: 8px 0;
            color: #6c757d;
        }
        .links {
            text-align: center;
            margin-top: 30px;
        }
        .links a {
            display: inline-block;
            margin: 10px;
            padding: 12px 24px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background-color 0.3s;
        }
        .links a:hover {
            background-color: #45a049;
        }
        .api-link {
            background-color: #2196F3 !important;
        }
        .api-link:hover {
            background-color: #1976D2 !important;
        }
        .stats {
            text-align: center;
            margin-top: 20px;
            padding: 15px;
            background-color: #fff3cd;
            border-radius: 5px;
            border: 1px solid #ffeaa7;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 传感器日志服务器</h1>
        
        <div class="status">
            <strong>✅ 服务器运行正常</strong><br>
            当前时间: {{.CurrentTime}}<br>
            运行环境: {{.Environment}}<br>
            日志级别: {{.LogLevel}}
        </div>

        <div class="info-grid">
            <div class="info-card">
                <h3>📊 数据统计</h3>
                <p>内存中数据: {{.MemoryDataCount}} 条</p>
                <p>最大存储: {{.MaxDataStore}} 条</p>
                <p>文件日志: {{.FileLogStatus}}</p>
            </div>
            
            <div class="info-card">
                <h3>🔗 连接信息</h3>
                <p>服务器地址: {{.ServerAddr}}</p>
                <p>数据接收端点: /data</p>
                <p>MongoDB: {{.MongoStatus}}</p>
            </div>
        </div>

        <div class="links">
            <a href="/dashboard">📈 数据仪表板</a>
            <a href="/api/data" class="api-link">📋 内存数据API</a>
            <a href="/api/db/data" class="api-link">🗄️ 数据库API</a>
            <a href="/api/db/stats" class="api-link">📊 统计API</a>
        </div>

        <div class="stats">
            <strong>💡 使用提示</strong><br>
            在Sensor Logger应用中设置推送URL为: <code>http://[你的IP地址]:{{.ServerPort}}/data</code>
        </div>
    </div>
</body>
</html>
`

	// 准备模板数据
	data := struct {
		CurrentTime     string
		Environment     string
		LogLevel        string
		MemoryDataCount int
		MaxDataStore    int
		FileLogStatus   string
		ServerAddr      string
		MongoStatus     string
		ServerPort      string
	}{
		CurrentTime:     time.Now().Format("2006-01-02 15:04:05"),
		Environment:     AppConfig.Environment,
		LogLevel:        AppConfig.LogLevel,
		MemoryDataCount: parsedDataStore.Len(),
		MaxDataStore:    AppConfig.MaxDataStore,
		FileLogStatus:   map[bool]string{true: "启用", false: "禁用"}[AppConfig.EnableFileLog],
		ServerAddr:      GetServerAddr(),
		MongoStatus:     map[bool]string{true: "已连接", false: "未连接"}[mongoClient != nil],
		ServerPort:      AppConfig.ServerPort,
	}

	tmpl, err := template.New("root").Parse(html)
	if err != nil {
		http.Error(w, "模板解析失败", http.StatusInternalServerError)
		LogError("模板解析", err)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		LogError("模板执行", err)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// handleSensorData 处理传感器数据
func handleSensorData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusMethodNotAllowed, time.Since(startTime))
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusBadRequest)
		LogError("读取请求体", err, slog.String("remote_addr", r.RemoteAddr))
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusBadRequest, time.Since(startTime))
		return
	}

	// 解析传感器数据
	parsedData, err := parseSensorMessage(body)
	if err != nil {
		http.Error(w, "解析传感器数据失败", http.StatusBadRequest)
		LogError("解析传感器数据", err, slog.String("remote_addr", r.RemoteAddr))
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusBadRequest, time.Since(startTime))
		return
	}

	// 记录传感器数据接收日志
	LogSensorData(parsedData.MessageID, parsedData.DeviceID, parsedData.SessionID, parsedData.TotalReadings)

	// 保存到MongoDB
	if mongoClient != nil {
		dbStart := time.Now()
		if err := SaveSensorData(parsedData); err != nil {
			LogDatabaseOperation("save_sensor_messages", false, parsedData.TotalReadings, time.Since(dbStart))
			LogError("保存到MongoDB", err,
				slog.String("device_id", parsedData.DeviceID),
				slog.Int64("message_id", parsedData.MessageID))
		} else {
			LogDatabaseOperation("save_sensor_messages", true, parsedData.TotalReadings, time.Since(dbStart))
		}
	}

	// 存储解析后的数据到内存（用于快速访问）
	parsedDataStore.Add(*parsedData)

	// 只保留最近的配置数量条记录
	parsedDataStore.TrimToSize(AppConfig.MaxDataStore)

	// 保存原始数据到文件
	if AppConfig.EnableFileLog {
		if err := saveToFile(body, parsedData.ReceivedAt); err != nil {
			LogError("保存文件", err, slog.String("device_id", parsedData.DeviceID))
		}
	}

	// 显示解析结果
	if AppConfig.EnableLogging {
		displayParsedData(parsedData)
	}

	// 响应成功
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("数据接收成功"))

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// saveToFile 保存原始数据到文件
func saveToFile(data []byte, timestamp time.Time) error {
	// 确保数据目录存在
	if err := os.MkdirAll(AppConfig.DataDir, 0755); err != nil {
		return fmt.Errorf("创建数据目录失败: %v", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("sensor_messages_%s.json", timestamp.Format("20060102_150405"))
	filepath := fmt.Sprintf("%s/%s", AppConfig.DataDir, filename)

	// 写入文件
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	Logger.Debug("数据已保存到文件", slog.String("filepath", filepath))
	return nil
}

// displayParsedData 显示解析后的数据
func displayParsedData(data *ParsedSensorData) {
	fmt.Println("\n=== 收到传感器数据 ===")
	fmt.Printf("时间: %s\n", data.ReceivedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("消息ID: %d\n", data.MessageID)
	fmt.Printf("设备ID: %s\n", data.DeviceID)
	fmt.Printf("会话ID: %s\n", data.SessionID)
	fmt.Printf("总读数: %d\n", data.TotalReadings)

	// 显示传感器类型
	if len(data.SensorTypes) > 0 {
		fmt.Printf("传感器类型: %s\n", strings.Join(data.SensorTypes, ", "))
	}

	// 显示文件保存信息
	if AppConfig.EnableFileLog {
		filename := fmt.Sprintf("sensor_messages_%s.json", data.ReceivedAt.Format("20060102_150405"))
		fmt.Printf("数据已保存到文件: %s/%s\n", AppConfig.DataDir, filename)
	}

	fmt.Println("========================")
}

// handleDashboard 处理仪表板请求
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>传感器数据仪表板</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background-color: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            text-align: center;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #4CAF50;
        }
        .stat-label {
            color: #666;
            margin-top: 10px;
        }
        .data-container {
            background-color: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .sensor-data {
            border: 1px solid #ddd;
            margin-bottom: 20px;
            padding: 15px;
            border-radius: 5px;
            background-color: #f9f9f9;
        }
        .sensor-header {
            font-weight: bold;
            color: #333;
            margin-bottom: 10px;
            padding-bottom: 10px;
            border-bottom: 2px solid #4CAF50;
        }
        .sensor-values {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 10px;
            margin-top: 10px;
        }
        .sensor-value {
            background-color: white;
            padding: 10px;
            border-radius: 5px;
            border-left: 4px solid #2196F3;
        }
        .refresh-btn {
            position: fixed;
            top: 20px;
            right: 20px;
            background-color: #4CAF50;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        .refresh-btn:hover {
            background-color: #45a049;
        }
        .no-data {
            text-align: center;
            color: #666;
            font-style: italic;
            padding: 50px;
        }
        .back-link {
            display: inline-block;
            margin-bottom: 20px;
            color: #4CAF50;
            text-decoration: none;
            font-weight: bold;
        }
        .back-link:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <a href="/" class="back-link">← 返回首页</a>
        <h1>📊 传感器数据仪表板</h1>
        
        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">{{.TotalMessages}}</div>
                <div class="stat-label">总消息数</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.TotalReadings}}</div>
                <div class="stat-label">总读数</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.SensorTypeCount}}</div>
                <div class="stat-label">传感器类型</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{{.DeviceCount}}</div>
                <div class="stat-label">设备数量</div>
            </div>
        </div>

        <div class="data-container">
            <h2>最新传感器数据</h2>
            {{if .HasData}}
                {{range .LatestData}}
                <div class="sensor-data">
                    <div class="sensor-header">
                        {{.SensorType}} - {{.ReadableTime}} ({{.Accuracy}})
                    </div>
                    <div class="sensor-values">
                        {{range .Values}}
                        <div class="sensor-value">
                            <strong>{{.Name}}</strong><br>
                            {{.Value}} {{.Unit}}<br>
                            <small>{{.Description}}</small>
                        </div>
                        {{end}}
                    </div>
                </div>
                {{end}}
            {{else}}
                <div class="no-data">
                    暂无数据。请确保Sensor Logger应用正在发送数据。
                </div>
            {{end}}
        </div>
    </div>

    <button class="refresh-btn" onclick="location.reload()">🔄 刷新</button>

    <script>
        // 每30秒自动刷新
        setTimeout(function() {
            location.reload();
        }, 30000);
    </script>
</body>
</html>
`

	// 准备仪表板数据
	dashboardData := prepareDashboardData()

	tmpl, err := template.New("dashboard").Parse(html)
	if err != nil {
		http.Error(w, "模板解析失败", http.StatusInternalServerError)
		LogError("仪表板模板解析", err)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, dashboardData); err != nil {
		LogError("仪表板模板执行", err)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// prepareDashboardData 准备仪表板数据
func prepareDashboardData() DashboardData {
	data := DashboardData{
		TotalMessages:   parsedDataStore.Len(),
		TotalReadings:   0,
		SensorTypeCount: 0,
		DeviceCount:     0,
		HasData:         !parsedDataStore.IsEmpty(),
		LatestData:      []HumanReadableSensorData{},
	}

	if parsedDataStore.IsEmpty() {
		return data
	}

	// 获取所有数据用于计算统计信息
	allData := parsedDataStore.GetAllForRead()

	// 计算统计信息
	sensorTypes := make(map[string]bool)
	devices := make(map[string]bool)

	for _, parsedData := range allData {
		data.TotalReadings += parsedData.TotalReadings
		devices[parsedData.DeviceID] = true

		for _, sensorType := range parsedData.SensorTypes {
			sensorTypes[sensorType] = true
		}
	}

	data.SensorTypeCount = len(sensorTypes)
	data.DeviceCount = len(devices)

	// 获取最新数据的前20条读数
	if latestData, exists := parsedDataStore.GetLatestOne(); exists {
		maxReadings := 20
		if len(latestData.ParsedReadings) < maxReadings {
			maxReadings = len(latestData.ParsedReadings)
		}
		data.LatestData = latestData.ParsedReadings[:maxReadings]
	}

	return data
}

// handleAPIData 处理API数据请求
func handleAPIData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	data := parsedDataStore.Get()

	if err := json.NewEncoder(w).Encode(data); err != nil {
		LogError("API数据编码", err)
		http.Error(w, "数据编码失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// handleDBData 处理数据库数据请求
func handleDBData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取查询参数
	query := r.URL.Query()
	limit := 50 // 默认限制
	if l := query.Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	deviceID := query.Get("device")
	sensorType := query.Get("sensor")

	// 从数据库获取数据
	dbStart := time.Now()
	data, err := GetSensorDataFromDB(limit, deviceID, sensorType)
	if err != nil {
		LogDatabaseOperation("get_sensor_messages", false, 0, time.Since(dbStart))
		LogError("数据库查询", err,
			slog.String("device", deviceID),
			slog.String("sensor", sensorType),
			slog.Int("limit", limit))
		http.Error(w, "数据库查询失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogDatabaseOperation("get_sensor_messages", true, len(data), time.Since(dbStart))

	if err := json.NewEncoder(w).Encode(data); err != nil {
		LogError("数据库API编码", err)
		http.Error(w, "数据编码失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// handleDeviceInfo 处理设备信息请求
func handleDeviceInfo(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	dbStart := time.Now()
	devices, err := GetDeviceInfo()
	if err != nil {
		LogDatabaseOperation("get_device_info", false, 0, time.Since(dbStart))
		LogError("设备信息查询", err)
		http.Error(w, "设备信息查询失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogDatabaseOperation("get_device_info", true, len(devices), time.Since(dbStart))

	if err := json.NewEncoder(w).Encode(devices); err != nil {
		LogError("设备信息API编码", err)
		http.Error(w, "数据编码失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}

// handleDBStats 处理数据库统计信息请求
func handleDBStats(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	dbStart := time.Now()
	stats, err := GetDashboardStats()
	if err != nil {
		LogDatabaseOperation("get_dashboard_stats", false, 0, time.Since(dbStart))
		LogError("统计信息查询", err)
		http.Error(w, "统计信息查询失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogDatabaseOperation("get_dashboard_stats", true, 1, time.Since(dbStart))

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		LogError("统计信息API编码", err)
		http.Error(w, "数据编码失败", http.StatusInternalServerError)
		LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusInternalServerError, time.Since(startTime))
		return
	}

	LogAPIRequest(r.Method, r.URL.Path, r.RemoteAddr, http.StatusOK, time.Since(startTime))
}
