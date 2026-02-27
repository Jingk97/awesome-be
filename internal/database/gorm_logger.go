// Package database 提供数据库访问功能
package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/jingpc/awesome-be/internal/logger"
)

// GormLogger GORM 日志适配器
//
// 初级工程师学习要点：
// - 实现 gorm/logger.Interface 接口
// - 将 GORM 的日志转发到我们的 Zap 日志系统
// - 这样所有日志（应用日志 + GORM 日志）都通过统一的日志系统输出
type GormLogger struct {
	logger        *logger.Logger
	logLevel      gormlogger.LogLevel
	slowThreshold time.Duration
}

// NewGormLogger 创建 GORM 日志适配器
//
// 初级工程师学习要点：
// - 接收我们的 Logger 实例和配置参数
// - 返回实现了 gorm/logger.Interface 的适配器
func NewGormLogger(log *logger.Logger, level string, slowThreshold time.Duration) *GormLogger {
	var logLevel gormlogger.LogLevel

	switch level {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	case "info":
		logLevel = gormlogger.Info
	default:
		logLevel = gormlogger.Info
	}

	return &GormLogger{
		logger:        log,
		logLevel:      logLevel,
		slowThreshold: slowThreshold,
	}
}

// LogMode 设置日志级别
//
// 初级工程师学习要点：
// - GORM 会调用这个方法来设置日志级别
// - 返回一个新的 logger 实例（不修改原实例）
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录 Info 级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录 Warn 级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录 Error 级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 记录 SQL 执行日志
//
// 初级工程师学习要点：
// - GORM 在执行 SQL 时会调用这个方法
// - fc 是调用位置信息
// - begin 是 SQL 开始执行时间
// - sql 是执行的 SQL 语句
// - rows 是影响的行数
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		// SQL 执行错误
		l.logger.Error("SQL execution error",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlogger.Warn:
		// 慢查询
		l.logger.Warn("Slow SQL query",
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.slowThreshold),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	case l.logLevel >= gormlogger.Info:
		// 正常 SQL 执行
		l.logger.Debug("SQL execution",
			zap.Duration("elapsed", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
		)
	}
}
