// Package example 示例 Handler
//
// 核心功能：
// - 演示成功响应
// - 演示各种错误响应
// - 演示 Panic 恢复
// - 演示数据库错误处理
//
// 初级工程师学习要点：
// - Handler 层负责处理 HTTP 请求和响应
// - Handler 层负责记录日志
// - Handler 层调用 Service 层处理业务逻辑
// - Handler 层不直接访问数据库
package example

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/jingpc/awesome-be/internal/logger"
	"github.com/jingpc/awesome-be/internal/service/example"
	"github.com/jingpc/awesome-be/pkg/errors"
	"github.com/jingpc/awesome-be/pkg/response"
)

// Handler 示例处理器
type Handler struct {
	logger  *logger.Logger
	service *example.Service
}

// NewHandler 创建示例处理器
//
// 架构思路：
// - 通过构造函数注入依赖
// - 便于单元测试
// - 便于替换实现
func NewHandler(logger *logger.Logger, service *example.Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// Ping 成功响应示例
//
// 初级工程师学习要点：
// - 使用 response.Success 返回成功响应
// - HTTP 状态码自动设置为 200
// - 业务错误码自动设置为 0
func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"message": "pong",
	})
}

// Error 错误响应示例
//
// 初级工程师学习要点：
// - Handler 层负责记录日志
// - 使用 response.Error 返回错误响应
// - HTTP 状态码根据错误码自动设置
// - 业务错误码和消息自动填充
func (h *Handler) Error(c *gin.Context) {
	// Handler 层记录日志
	h.logger.Warn("user not found example",
		zap.String("user_id", "123"),
		zap.String("path", c.Request.URL.Path),
	)

	// 返回业务错误
	response.Error(c, errors.ErrUserNotFound.WithDetail("user id: 123"))
}

// Panic Panic 恢复示例
//
// 初级工程师学习要点：
// - Panic 会被 Recovery 中间件捕获
// - 自动记录堆栈信息
// - 返回统一的错误响应
// - 不会导致程序崩溃
func (h *Handler) Panic(c *gin.Context) {
	panic("test panic - this will be recovered by middleware")
}

// DBError 数据库错误示例
//
// 初级工程师学习要点：
// - Handler 调用 Service 处理业务逻辑
// - Service 返回错误，Handler 记录日志
// - GORM 错误会自动转换为业务错误
// - 使用 response.Error 返回错误响应
//
// 架构思路：
// - Handler 层不直接访问数据库
// - Service 层负责业务逻辑
// - 错误在 Handler 层统一处理和记录
func (h *Handler) DBError(c *gin.Context) {
	// 调用 Service 层
	err := h.service.GetUser(c.Request.Context(), 999)
	if err != nil {
		// Handler 层记录日志
		h.logger.Error("failed to get user",
			zap.Error(err),
			zap.Int("user_id", 999),
			zap.String("path", c.Request.URL.Path),
		)

		// 返回错误 (自动转换 GORM 错误)
		response.Error(c, err)
		return
	}

	// 正常情况下返回数据
	response.Success(c, gin.H{
		"message": "user found",
	})
}

// NotFound 404 错误示例
//
// 初级工程师学习要点：
// - 使用预定义的错误码
// - 可以添加详细信息
// - HTTP 状态码自动设置为 404
func (h *Handler) NotFound(c *gin.Context) {
	h.logger.Warn("resource not found example",
		zap.String("path", c.Request.URL.Path),
	)

	response.Error(c, errors.ErrNotFound.WithDetail("the requested resource does not exist"))
}
