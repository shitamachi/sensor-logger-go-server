package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Logger 全局日志实例
var Logger *slog.Logger

// LogConfig 日志配置
type LogConfig struct {
	Level       slog.Level
	Environment string // "dev" 或 "production"
	EnableFile  bool
	FilePath    string
	AddSource   bool
}

// InitLogger 初始化日志系统
func InitLogger(config LogConfig) error {
	var handler slog.Handler

	// 根据环境选择不同的处理器
	switch config.Environment {
	case "production":
		handler = createProductionHandler(config)
	case "dev", "development":
		handler = createDevelopmentHandler(config)
	default:
		// 默认使用开发环境配置
		handler = createDevelopmentHandler(config)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	// 记录初始化信息
	Logger.Info("日志系统初始化完成",
		slog.String("environment", config.Environment),
		slog.String("level", config.Level.String()),
		slog.Bool("file_logging", config.EnableFile),
		slog.Bool("source_info", config.AddSource))

	return nil
}

// createDevelopmentHandler 创建开发环境日志处理器
func createDevelopmentHandler(config LogConfig) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 自定义时间格式
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05.000")),
				}
			}
			// 简化源码路径
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					source.File = filepath.Base(source.File)
				}
			}
			return a
		},
	}

	if config.EnableFile {
		return createMultiHandler(opts, config.FilePath)
	}

	return slog.NewTextHandler(os.Stdout, opts)
}

// createProductionHandler 创建生产环境日志处理器
func createProductionHandler(config LogConfig) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// 生产环境使用标准时间格式
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	if config.EnableFile {
		return createMultiHandler(opts, config.FilePath)
	}

	return slog.NewJSONHandler(os.Stdout, opts)
}

// createMultiHandler 创建多输出处理器（同时输出到控制台和文件）
func createMultiHandler(opts *slog.HandlerOptions, filePath string) slog.Handler {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		slog.Error("创建日志目录失败", "error", err)
		return slog.NewTextHandler(os.Stdout, opts)
	}

	// 打开日志文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("打开日志文件失败", "error", err)
		return slog.NewTextHandler(os.Stdout, opts)
	}

	// 创建多输出处理器
	return &MultiHandler{
		handlers: []slog.Handler{
			slog.NewTextHandler(os.Stdout, opts), // 控制台输出
			slog.NewJSONHandler(file, opts),      // 文件输出（JSON格式）
		},
	}
}

// MultiHandler 多输出处理器
type MultiHandler struct {
	handlers []slog.Handler
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// GetLogConfig 从应用配置获取日志配置
func GetLogConfig() LogConfig {
	// 确定环境
	environment := "dev"
	if AppConfig.LogLevel == "warn" || AppConfig.LogLevel == "error" {
		environment = "production"
	}

	// 设置日志文件路径
	logFile := filepath.Join(AppConfig.DataDir, "logs", "sensor-logger.log")

	return LogConfig{
		Level:       parseLogLevel(AppConfig.LogLevel),
		Environment: environment,
		EnableFile:  AppConfig.EnableFileLog,
		FilePath:    logFile,
		AddSource:   environment == "dev",
	}
}

// LogSensorData 记录传感器数据接收日志
func LogSensorData(messageID int64, deviceID, sessionID string, sensorCount int) {
	Logger.Debug("传感器数据接收",
		slog.Int64("message_id", messageID),
		slog.String("device_id", deviceID),
		slog.String("session_id", sessionID),
		slog.Int("sensor_count", sensorCount))
}

// LogDatabaseOperation 记录数据库操作日志
func LogDatabaseOperation(operation string, success bool, recordCount int, duration time.Duration) {
	if success {
		Logger.Debug("数据库操作成功",
			slog.String("operation", operation),
			slog.Int("records", recordCount),
			slog.Duration("duration", duration))
	} else {
		Logger.Error("数据库操作失败",
			slog.String("operation", operation),
			slog.Int("records", recordCount),
			slog.Duration("duration", duration))
	}
}

// LogAPIRequest 记录API请求日志
func LogAPIRequest(method, path, remoteAddr string, statusCode int, duration time.Duration) {
	Logger.Debug("API请求",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("remote_addr", remoteAddr),
		slog.Int("status_code", statusCode),
		slog.Duration("duration", duration))
}

// LogError 记录错误日志
func LogError(operation string, err error, attrs ...slog.Attr) {
	args := []any{
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	}

	for _, attr := range attrs {
		args = append(args, attr)
	}

	Logger.Error("操作失败", args...)
}

// LogStartup 记录启动日志
func LogStartup(serverAddr string, config map[string]interface{}) {
	Logger.Info("服务器启动",
		slog.String("server_addr", serverAddr),
		slog.Any("config", config))
}

// LogShutdown 记录关闭日志
func LogShutdown(reason string) {
	Logger.Info("服务器关闭", slog.String("reason", reason))
}
