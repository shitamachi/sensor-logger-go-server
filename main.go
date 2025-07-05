package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// 版本信息变量（通过构建时注入）
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	GitBranch = "unknown"
)

func main() {
	// 加载配置
	if err := LoadConfig(); err != nil {
		fmt.Printf("配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	logConfig := GetLogConfig()
	if err := InitLogger(logConfig); err != nil {
		fmt.Printf("日志系统初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化MongoDB连接
	if err := InitMongoDB(); err != nil {
		Logger.Error("MongoDB初始化失败", slog.String("error", err.Error()))
		Logger.Info("将继续运行，但不会保存数据到数据库")
	}

	// 设置优雅关闭
	setupGracefulShutdown()

	// 设置路由
	http.HandleFunc("/data", handleSensorData)
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/dashboard", handleDashboard)
	http.HandleFunc("/api/data", handleAPIData)
	http.HandleFunc("/api/db/data", handleDBData)
	http.HandleFunc("/api/db/devices", handleDeviceInfo)
	http.HandleFunc("/api/db/stats", handleDBStats)

	// 显示启动信息
	fmt.Println("=== 传感器日志服务器 ===")
	fmt.Printf("版本: %s\n", Version)
	fmt.Printf("构建时间: %s\n", BuildTime)
	fmt.Printf("Git提交: %s\n", GitCommit)
	fmt.Printf("Git分支: %s\n", GitBranch)
	fmt.Printf("服务器启动在端口 %s\n", AppConfig.ServerPort)

	// 显示配置信息
	if AppConfig.EnableLogging {
		PrintConfig()
	}

	fmt.Println("\n本机IP地址:")
	getLocalIPs()
	fmt.Printf("\n请在Sensor Logger应用中设置推送URL为: http://[你的IP地址]:%s/data\n", AppConfig.ServerPort)
	fmt.Println("使用 'Tap to Test Pushing' 按钮测试连接")
	fmt.Printf("访问 http://[你的IP地址]:%s/dashboard 查看数据仪表板\n", AppConfig.ServerPort)

	// 显示API端点
	fmt.Println("\n=== API端点 ===")
	fmt.Printf("内存数据API: http://[你的IP地址]:%s/api/data\n", AppConfig.ServerPort)
	fmt.Printf("数据库数据API: http://[你的IP地址]:%s/api/db/data\n", AppConfig.ServerPort)
	fmt.Printf("设备信息API: http://[你的IP地址]:%s/api/db/devices\n", AppConfig.ServerPort)
	fmt.Printf("统计信息API: http://[你的IP地址]:%s/api/db/stats\n", AppConfig.ServerPort)
	fmt.Println("===============")

	// 启动服务器
	serverAddr := GetServerAddr()

	// 记录启动日志
	configMap := map[string]interface{}{
		"environment":    AppConfig.Environment,
		"log_level":      AppConfig.LogLevel,
		"mongo_enabled":  mongoClient != nil,
		"file_log":       AppConfig.EnableFileLog,
		"max_data_store": AppConfig.MaxDataStore,
	}
	LogStartup(serverAddr, configMap)

	Logger.Info("服务器启动完成", slog.String("address", serverAddr))

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		Logger.Error("服务器启动失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		LogShutdown("收到关闭信号")

		// 关闭MongoDB连接
		if err := CloseMongoDB(); err != nil {
			Logger.Error("关闭MongoDB连接失败", slog.String("error", err.Error()))
		}

		Logger.Info("服务器已关闭")
		os.Exit(0)
	}()
}
