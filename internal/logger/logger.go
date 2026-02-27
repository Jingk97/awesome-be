// Package logger 提供高性能的结构化日志功能
//
// 日志模块基于 Uber 的 Zap 库封装，提供统一的日志接口。
//
// 初级工程师学习要点：
// - 理解结构化日志的优势（便于查询和分析）
// - 掌握日志级别的使用场景
// - 学习 Context 传递请求级别数据的方式
package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jingpc/awesome-be/internal/config"
)

// contextKey 是 context 中存储值的键类型
type contextKey string

const (
	// traceIDKey 是 TraceID 在 context 中的键
	traceIDKey contextKey = "trace_id"
)

// Logger 是日志记录器
//
// 初级工程师学习要点：
// - Logger 封装了 Zap，提供更简单的接口
// - 通过 Context 传递 TraceID，实现请求链路追踪
type Logger struct {
	zap    *zap.Logger
	config config.LoggerConfig
}

// New 创建新的日志记录器
//
// 初级工程师学习要点：
// - 这是工厂函数，用于创建 Logger 实例
// - 根据配置决定日志格式（JSON 或 Console）和输出位置
// - 支持同时输出到控制台和文件（云原生友好）
// - 如果输出到文件，会自动进行日志轮转（避免文件过大）
func New(cfg config.LoggerConfig) (*Logger, error) {
	// 1. 配置日志字段的编码方式
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 级别名小写（debug, info, warn, error）
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间格式：2006-01-02T15:04:05.000Z
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 调用位置：文件名:行号
	}

	// 2. 根据配置选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		// JSON 格式：适合生产环境，便于日志分析工具解析
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		// Console 格式：适合开发环境，人类可读
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 3. 解析日志级别
	level := parseLevel(cfg.Level)

	// 4. 配置输出目标（支持多个目标）
	var writers []zapcore.WriteSyncer

	// 4.1 控制台输出
	if cfg.Console.Enabled {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 4.2 文件输出
	if cfg.File.Enabled {
		// 使用 lumberjack 实现日志轮转
		// 日志轮转：当日志文件达到一定大小或时间后，自动创建新文件
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File.Filename,
			MaxSize:    cfg.File.MaxSize,    // 单个文件最大大小（MB）
			MaxBackups: cfg.File.MaxBackups, // 保留的旧文件数量
			MaxAge:     cfg.File.MaxAge,     // 保留的天数
			Compress:   cfg.File.Compress,   // 是否压缩旧文件
		})
		writers = append(writers, fileWriter)
	}

	// 4.3 如果没有启用任何输出，默认输出到 stdout
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 4.4 合并多个输出目标
	// 初级工程师学习要点：
	// - NewMultiWriteSyncer 可以将日志同时写入多个目标
	// - 这样可以同时输出到控制台（供 Kubernetes 收集）和文件（本地备份）
	writeSyncer := zapcore.NewMultiWriteSyncer(writers...)

	// 5. 创建 Core（Zap 的核心组件）
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 6. 配置可选项
	opts := []zap.Option{}

	// 是否显示调用位置（文件名和行号）
	if cfg.EnableCaller {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(1)) // 跳过封装层，显示真实调用位置
	}

	// 是否显示堆栈信息（仅 Error 级别以上）
	if cfg.EnableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// 7. 创建 Logger
	zapLogger := zap.New(core, opts...)

	return &Logger{
		zap:    zapLogger,
		config: cfg,
	}, nil
}

// parseLevel 解析日志级别字符串
//
// 初级工程师学习要点：
// - 日志级别从低到高：debug < info < warn < error < fatal
// - 设置某个级别后，只会输出该级别及以上的日志
// - 例如：设置 info，则会输出 info、warn、error、fatal
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// WithTraceID 将 TraceID 存储到 Context 中
//
// 初级工程师学习要点：
// - Context 用于在函数调用链中传递数据
// - TraceID 是请求的唯一标识，用于追踪一个请求的完整链路
// - 通常在中间件中从 HTTP Header 读取 TraceID 并存入 Context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID 从 Context 中获取 TraceID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// withContext 从 Context 中提取字段并添加到日志
//
// 初级工程师学习要点：
// - 这个方法会自动从 Context 中提取 TraceID
// - 这样每条日志都会包含 TraceID，方便追踪请求
func (l *Logger) withContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return l.zap
	}

	fields := []zap.Field{}

	// 添加 TraceID（如果存在）
	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	if len(fields) > 0 {
		return l.zap.With(fields...)
	}

	return l.zap
}

// Debug 记录调试级别日志
//
// 初级工程师学习要点：
// - Debug 用于开发调试，记录详细的执行信息
// - 生产环境通常不输出 Debug 日志
// - 参数采用 key-value 对的形式，便于结构化查询
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.zap.Sugar().Debugw(msg, keysAndValues...)
}

// DebugContext 记录调试级别日志（带 Context）
func (l *Logger) DebugContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withContext(ctx).Sugar().Debugw(msg, keysAndValues...)
}

// Info 记录信息级别日志
//
// 初级工程师学习要点：
// - Info 用于记录正常的业务流程
// - 例如：用户登录、订单创建、服务启动等
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.zap.Sugar().Infow(msg, keysAndValues...)
}

// InfoContext 记录信息级别日志（带 Context）
func (l *Logger) InfoContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withContext(ctx).Sugar().Infow(msg, keysAndValues...)
}

// Warn 记录警告级别日志
//
// 初级工程师学习要点：
// - Warn 用于记录警告信息，不影响主流程
// - 例如：配置缺失使用默认值、重试操作等
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.zap.Sugar().Warnw(msg, keysAndValues...)
}

// WarnContext 记录警告级别日志（带 Context）
func (l *Logger) WarnContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withContext(ctx).Sugar().Warnw(msg, keysAndValues...)
}

// Error 记录错误级别日志
//
// 初级工程师学习要点：
// - Error 用于记录错误信息，影响当前请求
// - 例如：数据库查询失败、API 调用失败等
// - 通常在 Handler 层记录错误
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.zap.Sugar().Errorw(msg, keysAndValues...)
}

// ErrorContext 记录错误级别日志（带 Context）
func (l *Logger) ErrorContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withContext(ctx).Sugar().Errorw(msg, keysAndValues...)
}

// Fatal 记录致命错误日志并退出程序
//
// 初级工程师学习要点：
// - Fatal 用于记录致命错误，服务无法继续运行
// - 调用后会自动退出程序（os.Exit(1)）
// - 例如：数据库连接失败、配置加载失败等
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.zap.Sugar().Fatalw(msg, keysAndValues...)
}

// FatalContext 记录致命错误日志并退出程序（带 Context）
func (l *Logger) FatalContext(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withContext(ctx).Sugar().Fatalw(msg, keysAndValues...)
}

// Sync 刷新日志缓冲区
//
// 初级工程师学习要点：
// - Zap 使用缓冲区提高性能，日志不是立即写入的
// - 程序退出前必须调用 Sync，确保所有日志都写入
// - 通常使用 defer logger.Sync() 确保一定会执行
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// GetZapLogger 获取底层的 Zap Logger
//
// 初级工程师学习要点：
// - 这个方法用于 Gin 中间件集成
// - 返回原始的 Zap Logger，用于替换 Gin 的默认日志
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zap
}
