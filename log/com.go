package log

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const (
	LogDirName      = "logs"
	LogMaxAge       = 20 * 24 * time.Hour
	LogRotationTime = 24 * time.Hour
	LogRotationSize = 5 * 1024 * 1024 // 5MB
)

// LogConfig 日志配置结构
type LogConfig struct {
	LogDir        string
	MaxAge        time.Duration
	RotationTime  time.Duration
	RotationSize  int64
	EnableConsole bool
}

// 默认配置
var defaultLogConfig = LogConfig{
	LogDir:        filepath.Join(".", LogDirName),
	MaxAge:        LogMaxAge,
	RotationTime:  LogRotationTime,
	RotationSize:  LogRotationSize,
	EnableConsole: true, // 默认启用控制台输出
}

// SetConfig 设置日志配置
func SetConfig(config LogConfig) {
	defaultLogConfig = config
	// 重新初始化日志
	initLog()
	initAccessLog()
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() LogConfig {
	return defaultLogConfig
}

// ComLogFormatter 通用日志格式
type ComLogFormatter struct{}

// AccessFormatter 访问日志格式
type AccessFormatter struct{}

// ComLoggerClient 通用日志实例
var ComLoggerClient *logrus.Logger

// AccessLoggerClient 访问日志实例
var AccessLoggerClient *logrus.Logger

// ComFormat 实现 logrus.Formatter 接口
func (f *ComLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] [%s] %s", timestamp, entry.Level.String(), entry.Message)
	return []byte(msg + "\n"), nil
}

// Format 实现 logrus.Formatter 接口
func (f *AccessFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	// 安全地从 entry.Data 获取值
	ip := getStringValue(entry.Data, "ip", "0.0.0.0")
	method := getStringValue(entry.Data, "method", "GET")
	url := getStringValue(entry.Data, "url", "/")
	statusCode := getStringValue(entry.Data, "status_code", "200")
	responseSize := getStringValue(entry.Data, "response_size", "0")

	msg := fmt.Sprintf("[%s] [%s] [%s] [%s] [%s] [%s] %s",
		timestamp, ip, method, url, statusCode, responseSize, entry.Message)
	return []byte(msg + "\n"), nil
}

// getStringValue 从 logrus.Fields 安全获取字符串值
func getStringValue(data logrus.Fields, key, defaultValue string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
		// 如果不是字符串类型，尝试转换为字符串
		return fmt.Sprintf("%v", value)
	}
	return defaultValue
}

// initLog 初始化通用日志
func initLog() {
	// 日志实例化
	ComLoggerClient = logrus.New()

	// 确保日志目录存在
	if err := os.MkdirAll(defaultLogConfig.LogDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	// 设置日志级别
	ComLoggerClient.SetLevel(logrus.InfoLevel)

	// 设置Info日志切割
	logInfoPath := filepath.Join(defaultLogConfig.LogDir, "info.%Y-%m-%d.log")
	logInfoWriter, err := rotatelogs.New(
		logInfoPath,
		rotatelogs.WithMaxAge(defaultLogConfig.MaxAge),
		rotatelogs.WithRotationTime(defaultLogConfig.RotationTime),
		rotatelogs.WithRotationSize(defaultLogConfig.RotationSize),
	)
	if err != nil {
		fmt.Printf("Failed to create info log writer: %v\n", err)
		return
	}

	// 设置Error日志切割
	logErrorPath := filepath.Join(defaultLogConfig.LogDir, "error.%Y-%m-%d.log")
	logErrorWriter, err := rotatelogs.New(
		logErrorPath,
		rotatelogs.WithMaxAge(defaultLogConfig.MaxAge),
		rotatelogs.WithRotationTime(defaultLogConfig.RotationTime),
		rotatelogs.WithRotationSize(defaultLogConfig.RotationSize),
	)
	if err != nil {
		fmt.Printf("Failed to create error log writer: %v\n", err)
		return
	}

	writeMap := lfshook.WriterMap{
		logrus.InfoLevel:  logInfoWriter,
		logrus.ErrorLevel: logErrorWriter,
		logrus.WarnLevel:  logInfoWriter, // 警告信息也记录到 info 文件中
	}
	lfHook := lfshook.NewHook(writeMap, &ComLogFormatter{})
	ComLoggerClient.AddHook(lfHook)

	if isDEV() {
		// 开发环境下同时输出到控制台
		ComLoggerClient.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			PadLevelText:    true,
		})
		ComLoggerClient.SetOutput(os.Stdout)
	} else {
		// 生产环境不输出到控制台
		ComLoggerClient.SetOutput(os.NewFile(0, os.DevNull))
	}
}

// initAccessLog 初始化访问日志
func initAccessLog() {
	// 日志实例化
	AccessLoggerClient = logrus.New()

	// 确保日志目录存在
	if err := os.MkdirAll(defaultLogConfig.LogDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}

	AccessLoggerClient.SetLevel(logrus.InfoLevel) // 设置日志级别

	// 设置日志切割
	logAccessPath := filepath.Join(defaultLogConfig.LogDir, "access.%Y-%m-%d.log")
	logInfoWriter, err := rotatelogs.New(
		logAccessPath,
		rotatelogs.WithMaxAge(defaultLogConfig.MaxAge),
		rotatelogs.WithRotationTime(defaultLogConfig.RotationTime),
		rotatelogs.WithRotationSize(defaultLogConfig.RotationSize),
	)
	if err != nil {
		fmt.Printf("Failed to create access log writer: %v\n", err)
		return
	}

	// 设置日志输出
	writeMap := lfshook.WriterMap{
		logrus.InfoLevel: logInfoWriter,
	}
	lfHook := lfshook.NewHook(writeMap, &AccessFormatter{}) // 自定义日志 hook
	AccessLoggerClient.AddHook(lfHook)

	if isDEV() {
		// 开发环境下同时输出到控制台
		AccessLoggerClient.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			PadLevelText:    true,
		})
		AccessLoggerClient.SetOutput(os.Stdout)
	} else {
		// 生产环境不输出到控制台
		AccessLoggerClient.SetOutput(os.NewFile(0, os.DevNull))
	}
}

// ComLoggerFmt 控制台日志输出函数
func ComLoggerFmt(s ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	timestamp = fmt.Sprintf("[%s] ", timestamp)
	msg := append([]any{timestamp}, s...)
	msg = append(msg, "\n")
	fmt.Print(msg...)
}

// Info 记录信息日志
func Info(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Info(args...)
	}
}

// Infof 记录格式化信息日志
func Infof(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Infof(format, args...)
	}
}

// Error 记录错误日志
func Error(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Error(args...)
	}
}

// Errorf 记录格式化错误日志
func Errorf(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Errorf(format, args...)
	}
}

// Warn 记录警告日志
func Warn(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Warn(args...)
	}
}

// Warnf 记录格式化警告日志
func Warnf(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Warnf(format, args...)
	}
}

// Debug 记录调试日志
func Debug(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Debug(args...)
	}
}

// Debugf 记录格式化调试日志
func Debugf(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Debugf(format, args...)
	}
}

// Fatal 记录致命错误日志
func Fatal(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Fatal(args...)
	}
}

// Fatalf 记录格式化致命错误日志
func Fatalf(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Fatalf(format, args...)
	}
}

// Panic 记录恐慌日志
func Panic(args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Panic(args...)
	}
}

// Panicf 记录格式化恐慌日志
func Panicf(format string, args ...any) {
	if ComLoggerClient != nil {
		ComLoggerClient.Panicf(format, args...)
	}
}

// AccessInfo 记录访问日志
func AccessInfo(message string, fields map[string]any) {
	if AccessLoggerClient != nil {
		AccessLoggerClient.WithFields(logrus.Fields(fields)).Info(message)
	}
}

// AccessInfof 格式化记录访问日志
func AccessInfof(fields map[string]any, format string, args ...any) {
	if AccessLoggerClient != nil {
		AccessLoggerClient.WithFields(logrus.Fields(fields)).Infof(format, args...)
	}
}
