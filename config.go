package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	ServerPort string
	ServerHost string

	// 数据库配置
	MongoURI      string
	MongoDatabase string
	MongoTimeout  int

	// 应用配置
	MaxDataStore  int
	EnableLogging bool
	LogLevel      string
	Environment   string

	// 文件存储配置
	DataDir       string
	EnableFileLog bool
}

// 默认配置
var defaultConfig = Config{
	ServerPort:    "18000",
	ServerHost:    "",
	MongoURI:      "mongodb://localhost:27017",
	MongoDatabase: "sensor_logger",
	MongoTimeout:  10,
	MaxDataStore:  100,
	EnableLogging: true,
	LogLevel:      "info",
	Environment:   "dev",
	DataDir:       "./data",
	EnableFileLog: true,
}

// 全局配置实例
var AppConfig Config

// LoadConfig 加载配置
func LoadConfig() error {
	// 从默认配置开始
	AppConfig = defaultConfig

	// 尝试加载.env文件
	if err := loadEnvFile(".env"); err != nil {
		// 在日志系统初始化前使用fmt.Printf
		fmt.Printf("警告: 无法加载.env文件: %v\n", err)
		fmt.Println("使用默认配置")
	}

	// 从系统环境变量覆盖配置
	loadFromEnv()

	// 验证配置
	if err := validateConfig(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}

	// 创建必要的目录
	if err := createDirectories(); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	return nil
}

// loadEnvFile 从.env文件加载配置
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析键值对
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// 在日志系统初始化前使用slog.Warn，但需要检查Logger是否已初始化
			if Logger != nil {
				Logger.Warn("env文件格式错误", 
					slog.Int("line", lineNum), 
					slog.String("content", line))
			} else {
				fmt.Printf("警告: .env文件第%d行格式错误: %s\n", lineNum, line)
			}
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 移除值两端的引号
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// 设置环境变量
		if err := os.Setenv(key, value); err != nil {
			if Logger != nil {
				Logger.Warn("设置环境变量失败", 
					slog.String("key", key), 
					slog.String("error", err.Error()))
			} else {
				fmt.Printf("警告: 无法设置环境变量 %s: %v\n", key, err)
			}
		}
	}

	return scanner.Err()
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv() {
	if val := os.Getenv("SERVER_PORT"); val != "" {
		AppConfig.ServerPort = val
	}
	if val := os.Getenv("SERVER_HOST"); val != "" {
		AppConfig.ServerHost = val
	}

	if val := os.Getenv("MONGO_URI"); val != "" {
		AppConfig.MongoURI = val
	}
	if val := os.Getenv("MONGO_DATABASE"); val != "" {
		AppConfig.MongoDatabase = val
	}
	if val := os.Getenv("MONGO_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			AppConfig.MongoTimeout = timeout
		}
	}

	if val := os.Getenv("MAX_DATA_STORE"); val != "" {
		if maxStore, err := strconv.Atoi(val); err == nil {
			AppConfig.MaxDataStore = maxStore
		}
	}
	if val := os.Getenv("ENABLE_LOGGING"); val != "" {
		AppConfig.EnableLogging = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		AppConfig.LogLevel = val
	}

	if val := os.Getenv("DATA_DIR"); val != "" {
		AppConfig.DataDir = val
	}
	if val := os.Getenv("ENABLE_FILE_LOG"); val != "" {
		AppConfig.EnableFileLog = strings.ToLower(val) == "true"
	}
	if val := os.Getenv("ENVIRONMENT"); val != "" {
		AppConfig.Environment = val
	}
}

// validateConfig 验证配置
func validateConfig() error {
	// 验证端口
	if port, err := strconv.Atoi(AppConfig.ServerPort); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("无效的服务器端口: %s", AppConfig.ServerPort)
	}

	// 验证MongoDB超时
	if AppConfig.MongoTimeout < 1 {
		return fmt.Errorf("MongoDB超时时间必须大于0: %d", AppConfig.MongoTimeout)
	}

	// 验证最大数据存储数量
	if AppConfig.MaxDataStore < 1 {
		return fmt.Errorf("最大数据存储数量必须大于0: %d", AppConfig.MaxDataStore)
	}

	// 验证日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if AppConfig.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("无效的日志级别: %s，支持的级别: %v", AppConfig.LogLevel, validLogLevels)
	}

	// 验证环境
	validEnvironments := []string{"dev", "development", "prod", "production"}
	isValidEnvironment := false
	for _, env := range validEnvironments {
		if AppConfig.Environment == env {
			isValidEnvironment = true
			break
		}
	}
	if !isValidEnvironment {
		return fmt.Errorf("无效的环境: %s，支持的环境: %v", AppConfig.Environment, validEnvironments)
	}

	return nil
}

// createDirectories 创建必要的目录
func createDirectories() error {
	if AppConfig.DataDir != "" {
		if err := os.MkdirAll(AppConfig.DataDir, 0755); err != nil {
			return fmt.Errorf("创建数据目录失败: %v", err)
		}
	}
	
	// 创建日志目录
	logDir := AppConfig.DataDir + "/logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	return nil
}

// GetServerAddr 获取服务器地址
func GetServerAddr() string {
	if AppConfig.ServerHost == "" {
		return ":" + AppConfig.ServerPort
	}
	return AppConfig.ServerHost + ":" + AppConfig.ServerPort
}

// PrintConfig 打印当前配置
func PrintConfig() {
	fmt.Println("=== 当前配置 ===")
	fmt.Printf("服务器端口: %s\n", AppConfig.ServerPort)
	fmt.Printf("服务器主机: %s\n", AppConfig.ServerHost)
	fmt.Printf("MongoDB URI: %s\n", AppConfig.MongoURI)
	fmt.Printf("MongoDB 数据库: %s\n", AppConfig.MongoDatabase)
	fmt.Printf("MongoDB 超时: %d秒\n", AppConfig.MongoTimeout)
	fmt.Printf("最大数据存储: %d条\n", AppConfig.MaxDataStore)
	fmt.Printf("启用日志: %t\n", AppConfig.EnableLogging)
	fmt.Printf("日志级别: %s\n", AppConfig.LogLevel)
	fmt.Printf("运行环境: %s\n", AppConfig.Environment)
	fmt.Printf("数据目录: %s\n", AppConfig.DataDir)
	fmt.Printf("启用文件日志: %t\n", AppConfig.EnableFileLog)
	fmt.Println("===============")
}


