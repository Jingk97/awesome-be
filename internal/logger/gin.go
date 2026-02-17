// Package logger 提供 Gin 中间件集成
package logger

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLogger 返回 Gin 日志中间件
//
// 初级工程师学习要点：
// - 中间件是在请求处理前后执行的函数
// - 这个中间件会记录每个 HTTP 请求的信息
// - 替换 Gin 默认的 Logger 中间件
func GinLogger(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 从 Header 中读取 TraceID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID != "" {
			// 将 TraceID 存入 Context
			ctx := WithTraceID(c.Request.Context(), traceID)
			c.Request = c.Request.WithContext(ctx)
		}

		// 处理请求
		c.Next()

		// 计算请求耗时
		latency := time.Since(start)

		// 记录请求日志
		logger.InfoContext(
			c.Request.Context(),
			"HTTP Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// GinRecovery 返回 Gin Recovery 中间件
//
// 初级工程师学习要点：
// - Recovery 中间件用于捕获 panic，防止程序崩溃
// - 当发生 panic 时，会记录错误日志并返回 500 错误
// - 替换 Gin 默认的 Recovery 中间件
func GinRecovery(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 日志
				logger.ErrorContext(
					c.Request.Context(),
					"Panic recovered",
					"error", fmt.Sprintf("%v", err),
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
				)

				// 返回 500 错误
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()

		c.Next()
	}
}
