// Package response 提供统一的 HTTP 响应格式
//
// 核心功能：
// - 统一的成功和错误响应格式
// - 自动添加 TraceID
// - 自动设置 HTTP 状态码
//
// 初级工程师学习要点：
// - 理解统一响应格式的重要性
// - 掌握如何从 Context 获取 TraceID
// - 学习如何设置 HTTP 状态码
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jingpc/awesome-be/pkg/errors"
)

// Response 统一响应结构
//
// 初级工程师学习要点：
// - Code: 业务错误码（0 表示成功）
// - Message: 错误消息或成功提示
// - Data: 响应数据（成功时返回）
// - TraceID: 链路追踪 ID（用于问题排查）
type Response struct {
	Code    int         `json:"code"`               // 业务错误码
	Message string      `json:"message"`            // 错误消息
	Data    interface{} `json:"data,omitempty"`     // 数据（成功时）
	TraceID string      `json:"trace_id,omitempty"` // 链路追踪 ID
}

// Success 成功响应
//
// 初级工程师学习要点：
// - HTTP 状态码固定为 200
// - 业务错误码为 0
// - 返回数据在 data 字段
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// SuccessWithMsg 带自定义消息的成功响应
func SuccessWithMsg(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
		TraceID: getTraceID(c),
	})
}

// Error 错误响应
//
// 初级工程师学习要点：
// - 自动转换标准错误为业务错误
// - HTTP 状态码根据错误码自动设置
// - 不返回 data 字段
func Error(c *gin.Context, err error) {
	// 转换为业务错误
	e := errors.FromError(err)
	if e == nil {
		e = errors.ErrInternalError
	}

	c.JSON(e.Code.HTTPStatus(), Response{
		Code:    int(e.Code),
		Message: e.Message,
		TraceID: getTraceID(c),
	})
}

// ErrorWithCode 使用错误码响应
func ErrorWithCode(c *gin.Context, code errors.Code) {
	c.JSON(code.HTTPStatus(), Response{
		Code:    int(code),
		Message: errors.GetMessage(code),
		TraceID: getTraceID(c),
	})
}

// ErrorWithMsg 使用自定义消息响应
func ErrorWithMsg(c *gin.Context, code errors.Code, message string) {
	c.JSON(code.HTTPStatus(), Response{
		Code:    int(code),
		Message: message,
		TraceID: getTraceID(c),
	})
}

// getTraceID 从 Context 获取 TraceID
//
// 初级工程师学习要点：
// - TraceID 由链路追踪中间件设置
// - 存储在 gin.Context 中
// - 用于关联日志和请求
func getTraceID(c *gin.Context) string {
	// 从 gin.Context 获取
	if traceID, exists := c.Get("trace_id"); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}

	// 从 Request.Context 获取
	if traceID := c.Request.Context().Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}

	// 从 Header 获取
	if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
		return traceID
	}

	return ""
}
