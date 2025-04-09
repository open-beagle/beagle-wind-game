package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// 将LogLevel转换为zap日志级别
func (l LogLevel) zapLevel() zapcore.Level {
	switch l {
	case DEBUG:
		return zapcore.DebugLevel
	case INFO:
		return zapcore.InfoLevel
	case WARN:
		return zapcore.WarnLevel
	case ERROR:
		return zapcore.ErrorLevel
	case FATAL:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// 日志级别字符串映射
var levelStrings = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// 日志输出目标
type LogOutput int

const (
	CONSOLE LogOutput = iota
	FILE
	BOTH
)

// Logger 日志接口
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(map[string]interface{}) Logger
	WithRequestID(requestID string) Logger
	Sync() error
}

// ZapLogger 基于zap的日志实现
type ZapLogger struct {
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
	level         LogLevel
	module        string
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      LogLevel
	Output     LogOutput
	FilePath   string
	Module     string
	MaxSize    int  // 单位：MB
	MaxAge     int  // 单位：天
	MaxBackups int  // 最大备份数
	Compress   bool // 是否压缩
}

var (
	defaultLogger *ZapLogger
	once          sync.Once
)

// 初始化默认日志器
func initDefaultLogger() {
	// 修改为支持参数的初始化方法
	InitLogger("", INFO, false)
}

// InitLogger 根据指定参数初始化日志器
// logFile: 日志文件路径，为空则仅输出到控制台
// level: 日志级别
// enableBoth: 是否同时输出到控制台和文件，仅当logFile不为空时有效
func InitLogger(logFile string, level LogLevel, enableBoth bool) {
	// 确定日志输出模式
	logOutput := CONSOLE // 默认控制台输出
	if logFile != "" {
		if enableBoth {
			logOutput = BOTH // 同时输出到控制台和文件
		} else {
			logOutput = FILE // 仅输出到文件
		}
	}

	// 确保日志目录存在
	if logFile != "" {
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "无法创建日志目录: %v\n", err)
		}
	}

	// 创建日志器配置
	config := LoggerConfig{
		Level:      level,
		Output:     logOutput,
		FilePath:   logFile,
		Module:     "DEFAULT",
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 5,
		Compress:   true,
	}

	// 创建日志器
	logger, err := newZapLogger(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法创建日志器: %v\n", err)
		os.Exit(1)
	}

	defaultLogger = logger
}

// 创建新的zap日志器
func newZapLogger(config LoggerConfig) (*ZapLogger, error) {
	// 定义日志编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 设置日志级别
	zapLevel := config.Level.zapLevel()
	atomicLevel := zap.NewAtomicLevelAt(zapLevel)

	// 配置输出
	var cores []zapcore.Core

	// 控制台输出
	if config.Output == CONSOLE || config.Output == BOTH {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atomicLevel)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if config.Output == FILE || config.Output == BOTH {
		// 配置日志轮转
		lumberJackLogger := &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(lumberJackLogger), atomicLevel)
		cores = append(cores, fileCore)
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建日志器
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.Fields(zap.String("module", config.Module)))

	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		level:         config.Level,
		module:        config.Module,
	}, nil
}

// GetLogger 获取默认日志器
func GetLogger() Logger {
	once.Do(initDefaultLogger)
	return defaultLogger
}

// New 创建新的日志器
func New(module string) Logger {
	once.Do(initDefaultLogger)

	// 从默认日志器创建新的日志器
	logger := defaultLogger.logger.With(zap.String("module", module))

	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		level:         defaultLogger.level,
		module:        module,
	}
}

// NewWithConfig 使用配置创建新的日志器
func NewWithConfig(config LoggerConfig) (Logger, error) {
	return newZapLogger(config)
}

// WithField 添加字段
func (l *ZapLogger) WithField(key string, value interface{}) Logger {
	logger := l.logger.With(zap.Any(key, value))
	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		level:         l.level,
		module:        l.module,
	}
}

// WithFields 添加多个字段
func (l *ZapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	logger := l.logger.With(zapFields...)
	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		level:         l.level,
		module:        l.module,
	}
}

// WithRequestID 添加请求ID
func (l *ZapLogger) WithRequestID(requestID string) Logger {
	logger := l.logger.With(zap.String("request_id", requestID))
	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		level:         l.level,
		module:        l.module,
	}
}

// Debug 输出调试级别日志
func (l *ZapLogger) Debug(format string, args ...interface{}) {
	l.sugaredLogger.Debugf(format, args...)
}

// Info 输出信息级别日志
func (l *ZapLogger) Info(format string, args ...interface{}) {
	l.sugaredLogger.Infof(format, args...)
}

// Warn 输出警告级别日志
func (l *ZapLogger) Warn(format string, args ...interface{}) {
	l.sugaredLogger.Warnf(format, args...)
}

// Error 输出错误级别日志
func (l *ZapLogger) Error(format string, args ...interface{}) {
	l.sugaredLogger.Errorf(format, args...)
}

// Fatal 输出致命错误级别日志，并退出程序
func (l *ZapLogger) Fatal(format string, args ...interface{}) {
	l.sugaredLogger.Fatalf(format, args...)
}

// Sync 同步日志
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}

// 程序退出时确保日志同步
func init() {
	// 这里不需要使用 signal 库，因为 zap 库会在程序退出时自动同步
	// 但我们还是创建一个 defer 函数，确保在主函数结束时调用 Sync
	if defaultLogger != nil {
		defer defaultLogger.Sync()
	}
}
