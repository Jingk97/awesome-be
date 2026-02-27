// Package response 错误处理中间件
package response

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jingpc/awesome-be/internal/logger"
	"github.com/jingpc/awesome-be/pkg/errors"
)

// Recovery Panic 恢复中间件
//
// 初级工程师学习要点：
// - 使用 defer + recover 捕获 panic
// - 记录详细的堆栈信息
// - 返回统一的错误响应
// - 中断后续处理（c.Abort）
//
// 架构思路：
// - 这是最后一道防线，捕获所有未处理的 panic
// - 记录完整的堆栈信息，便于问题排查
// - 不暴露内部错误细节给客户端
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := string(debug.Stack())

				// 记录 Panic 日志
				log.Error("panic recovered",
					zap.Any("error", err),
					zap.String("stack", stack),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
				)

				// 返回统一错误响应
				Error(c, errors.ErrPanic)

				// 中断请求
				c.Abort()
			}
		}()

		c.Next()
	}
}
