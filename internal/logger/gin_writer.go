// Package logger Gin 输出适配器
package logger

import (
	"strings"

	"go.uber.org/zap"
)

// GinWriter Gin 框架输出适配器
//
// 架构思路：
// - 实现 io.Writer 接口，将 Gin 的输出重定向到我们的日志系统
// - Gin 的 debug 日志（路由注册等）也会被记录到日志文件
//
// 初级工程师学习要点：
// - 理解 io.Writer 接口的作用
// - 掌握如何适配第三方库的日志输出
type GinWriter struct {
	logger *Logger
}

// NewGinWriter 创建 Gin 输出适配器
func NewGinWriter(logger *Logger) *GinWriter {
	return &GinWriter{
		logger: logger,
	}
}

// Write 实现 io.Writer 接口
//
// 初级工程师学习要点：
// - io.Writer 接口只有一个方法：Write(p []byte) (n int, err error)
// - 返回写入的字节数和错误
// - Gin 会调用这个方法输出日志
func (w *GinWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	msg = strings.TrimSpace(msg)

	// 过滤空消息
	if msg == "" {
		return len(p), nil
	}

	// 根据消息内容判断日志级别
	if strings.Contains(msg, "[WARNING]") || strings.Contains(msg, "[WARN]") {
		w.logger.Warn("gin framework", zap.String("message", msg))
	} else if strings.Contains(msg, "[ERROR]") {
		w.logger.Error("gin framework", zap.String("message", msg))
	} else {
		// [GIN-debug] 和其他信息级别的日志
		w.logger.Debug("gin framework", zap.String("message", msg))
	}

	return len(p), nil
}
