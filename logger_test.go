package main

import (
	"log/slog"
	"testing"
	"time"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"INFO", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"WARN", slog.LevelWarn},
		{"error", slog.LevelError},
		{"ERROR", slog.LevelError},
		{"invalid", slog.LevelInfo}, // 默认值
		{"", slog.LevelInfo},        // 默认值
	}

	for _, test := range tests {
		result := parseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLogLevel(%s) = %v, 期望 %v", test.input, result, test.expected)
		}
	}
}

func TestGetLogConfig(t *testing.T) {
	// 保存原始配置
	originalConfig := AppConfig

	// 测试开发环境配置
	AppConfig = Config{
		LogLevel:      "debug",
		EnableFileLog: true,
		DataDir:       "./test_data",
	}

	config := GetLogConfig()

	if config.Level != slog.LevelDebug {
		t.Errorf("期望日志级别为DEBUG，实际为%v", config.Level)
	}

	if config.Environment != "dev" {
		t.Errorf("期望环境为dev，实际为%s", config.Environment)
	}

	if !config.EnableFile {
		t.Error("期望启用文件日志")
	}

	if !config.AddSource {
		t.Error("期望在开发环境中添加源码信息")
	}

	// 测试生产环境配置
	AppConfig.LogLevel = "warn"
	config = GetLogConfig()

	if config.Environment != "production" {
		t.Errorf("期望环境为production，实际为%s", config.Environment)
	}

	if config.AddSource {
		t.Error("期望在生产环境中不添加源码信息")
	}

	// 恢复原始配置
	AppConfig = originalConfig
}

func TestLogStructures(t *testing.T) {
	// 测试日志配置结构
	config := LogConfig{
		Level:       slog.LevelInfo,
		Environment: "test",
		EnableFile:  false,
		FilePath:    "/tmp/test.log",
		AddSource:   true,
	}

	if config.Level != slog.LevelInfo {
		t.Errorf("期望日志级别为INFO，实际为%v", config.Level)
	}

	if config.Environment != "test" {
		t.Errorf("期望环境为test，实际为%s", config.Environment)
	}
}

func TestLogFunctions(t *testing.T) {
	// 初始化一个测试用的Logger
	originalLogger := Logger
	defer func() {
		Logger = originalLogger
	}()

	// 设置测试配置
	testConfig := LogConfig{
		Level:       slog.LevelInfo,
		Environment: "test",
		EnableFile:  false,
		AddSource:   false,
	}

	// 初始化Logger
	if err := InitLogger(testConfig); err != nil {
		t.Fatalf("初始化Logger失败: %v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("日志函数引发panic: %v", r)
		}
	}()

	// 测试各种日志函数
	LogSensorData(123, "test-device", "test-session", 5)
	LogDatabaseOperation("test-op", true, 10, time.Millisecond*100)
	LogAPIRequest("GET", "/test", "127.0.0.1", 200, time.Millisecond*50)
	LogStartup(":8080", map[string]interface{}{"test": "value"})
	LogShutdown("测试关闭")
}

func TestMultiHandlerInterface(t *testing.T) {
	// 测试MultiHandler实现了slog.Handler接口
	handler := &MultiHandler{
		handlers: []slog.Handler{},
	}

	// 测试接口方法存在
	_ = handler.Enabled(nil, slog.LevelInfo)
	_ = handler.WithAttrs([]slog.Attr{})
	_ = handler.WithGroup("test")
}
