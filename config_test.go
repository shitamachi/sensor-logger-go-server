package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// 保存原始环境变量
	originalEnvs := make(map[string]string)
	envKeys := []string{"SERVER_PORT", "MONGO_URI", "MONGO_DATABASE", "MAX_DATA_STORE", "LOG_LEVEL"}

	for _, key := range envKeys {
		originalEnvs[key] = os.Getenv(key)
	}

	// 清理环境变量
	for _, key := range envKeys {
		os.Unsetenv(key)
	}

	// 测试默认配置加载
	err := LoadConfig()
	if err != nil {
		t.Fatalf("加载默认配置失败: %v", err)
	}

	// 验证默认值
	if AppConfig.ServerPort != "18000" {
		t.Errorf("期望服务器端口为 18000，实际为 %s", AppConfig.ServerPort)
	}

	if AppConfig.MongoDatabase != "sensor_logger" {
		t.Errorf("期望MongoDB数据库为 sensor_logger，实际为 %s", AppConfig.MongoDatabase)
	}

	if AppConfig.MaxDataStore != 100 {
		t.Errorf("期望最大数据存储为 100，实际为 %d", AppConfig.MaxDataStore)
	}

	if AppConfig.LogLevel != "info" {
		t.Errorf("期望日志级别为 info，实际为 %s", AppConfig.LogLevel)
	}

	// 恢复原始环境变量
	for key, value := range originalEnvs {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}

func TestLoadFromEnv(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("MONGO_DATABASE", "test_db")
	os.Setenv("MAX_DATA_STORE", "200")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENABLE_LOGGING", "false")

	// 重新加载配置
	AppConfig = defaultConfig
	loadFromEnv()

	// 验证环境变量覆盖
	if AppConfig.ServerPort != "8080" {
		t.Errorf("期望服务器端口为 8080，实际为 %s", AppConfig.ServerPort)
	}

	if AppConfig.MongoDatabase != "test_db" {
		t.Errorf("期望MongoDB数据库为 test_db，实际为 %s", AppConfig.MongoDatabase)
	}

	if AppConfig.MaxDataStore != 200 {
		t.Errorf("期望最大数据存储为 200，实际为 %d", AppConfig.MaxDataStore)
	}

	if AppConfig.LogLevel != "debug" {
		t.Errorf("期望日志级别为 debug，实际为 %s", AppConfig.LogLevel)
	}

	if AppConfig.EnableLogging != false {
		t.Errorf("期望启用日志为 false，实际为 %t", AppConfig.EnableLogging)
	}

	// 清理测试环境变量
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("MONGO_DATABASE")
	os.Unsetenv("MAX_DATA_STORE")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("ENABLE_LOGGING")
}

func TestValidateConfig(t *testing.T) {
	// 测试有效配置
	AppConfig = defaultConfig
	if err := validateConfig(); err != nil {
		t.Errorf("有效配置验证失败: %v", err)
	}

	// 测试无效端口
	AppConfig.ServerPort = "invalid"
	if err := validateConfig(); err == nil {
		t.Error("期望无效端口验证失败，但验证通过了")
	}

	// 测试端口超出范围
	AppConfig.ServerPort = "99999"
	if err := validateConfig(); err == nil {
		t.Error("期望端口超出范围验证失败，但验证通过了")
	}

	// 测试无效日志级别
	AppConfig = defaultConfig
	AppConfig.LogLevel = "invalid"
	if err := validateConfig(); err == nil {
		t.Error("期望无效日志级别验证失败，但验证通过了")
	}

	// 测试无效超时时间
	AppConfig = defaultConfig
	AppConfig.MongoTimeout = 0
	if err := validateConfig(); err == nil {
		t.Error("期望无效超时时间验证失败，但验证通过了")
	}
}

func TestGetServerAddr(t *testing.T) {
	// 测试只有端口
	AppConfig.ServerPort = "8080"
	AppConfig.ServerHost = ""
	addr := GetServerAddr()
	expected := ":8080"
	if addr != expected {
		t.Errorf("期望服务器地址为 %s，实际为 %s", expected, addr)
	}

	// 测试有主机和端口
	AppConfig.ServerHost = "localhost"
	addr = GetServerAddr()
	expected = "localhost:8080"
	if addr != expected {
		t.Errorf("期望服务器地址为 %s，实际为 %s", expected, addr)
	}
}

func TestLoadEnvFile(t *testing.T) {
	// 创建测试.env文件
	testEnvContent := `# 测试配置文件
SERVER_PORT=9000
MONGO_DATABASE=test_sensor
# 这是注释
MAX_DATA_STORE=500
LOG_LEVEL="debug"
ENABLE_LOGGING='true'

# 空行测试
INVALID_LINE_NO_EQUALS
DATA_DIR=/tmp/test
`

	// 写入测试文件
	testFile := "test.env"
	if err := os.WriteFile(testFile, []byte(testEnvContent), 0644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	defer os.Remove(testFile)

	// 加载测试文件
	if err := loadEnvFile(testFile); err != nil {
		t.Fatalf("加载测试文件失败: %v", err)
	}

	// 验证环境变量设置
	if os.Getenv("SERVER_PORT") != "9000" {
		t.Errorf("期望环境变量 SERVER_PORT 为 9000，实际为 %s", os.Getenv("SERVER_PORT"))
	}

	if os.Getenv("MONGO_DATABASE") != "test_sensor" {
		t.Errorf("期望环境变量 MONGO_DATABASE 为 test_sensor，实际为 %s", os.Getenv("MONGO_DATABASE"))
	}

	if os.Getenv("LOG_LEVEL") != "debug" {
		t.Errorf("期望环境变量 LOG_LEVEL 为 debug，实际为 %s", os.Getenv("LOG_LEVEL"))
	}

	if os.Getenv("ENABLE_LOGGING") != "true" {
		t.Errorf("期望环境变量 ENABLE_LOGGING 为 true，实际为 %s", os.Getenv("ENABLE_LOGGING"))
	}

	// 清理测试环境变量
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("MONGO_DATABASE")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("ENABLE_LOGGING")
	os.Unsetenv("DATA_DIR")
}
