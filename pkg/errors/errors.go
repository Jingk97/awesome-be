// Package errors 提供统一的错误处理机制
//
// 核心设计原则：
// 1. 错误和日志解耦 - 错误传递不依赖日志系统
// 2. 分层处理 - Repository 返回原始错误，Service 包装错误，Handler 记录日志
// 3. 统一响应 - 所有错误最终转换为统一的 HTTP 响应格式
//
// 初级工程师学习要点：
// - 理解错误码的分类体系
// - 掌握错误的包装和传递
// - 学习如何在不同层级处理错误
package errors

import (
	"fmt"
	"net/http"
)

// Code 错误码类型
type Code int

// Error 业务错误
//
// 初级工程师学习要点：
// - Error 实现了 error 接口
// - 包含错误码、消息、详细信息和原始错误
// - 支持错误链（通过 Unwrap 方法）
type Error struct {
	Code    Code   // 错误码
	Message string // 用户可见的错误消息
	Detail  string // 详细错误信息（可选，用于日志）
	Err     error  // 原始错误（用于错误链）
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误，支持 errors.Is 和 errors.As
func (e *Error) Unwrap() error {
	return e.Err
}

// WithError 添加原始错误
//
// 初级工程师学习要点：
// - 返回新的 Error 实例，不修改原实例（不可变性）
// - 保留错误链，便于追踪错误来源
func (e *Error) WithError(err error) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Detail:  e.Detail,
		Err:     err,
	}
}

// WithDetail 添加详细信息
func (e *Error) WithDetail(detail string) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Detail:  detail,
		Err:     e.Err,
	}
}

// WithDetailf 添加格式化的详细信息
func (e *Error) WithDetailf(format string, args ...interface{}) *Error {
	return &Error{
		Code:    e.Code,
		Message: e.Message,
		Detail:  fmt.Sprintf(format, args...),
		Err:     e.Err,
	}
}

// HTTPStatus 返回对应的 HTTP 状态码
//
// 初级工程师学习要点：
// - 错误码和 HTTP 状态码的映射关系
// - 1xxx -> 500 (系统错误)
// - 400x -> 400 (参数错误)
// - 401x -> 401 (认证错误)
// - 403x -> 403 (权限错误)
// - 404x -> 404 (资源不存在)
// - 409x -> 409 (冲突错误)
// - 429x -> 429 (限流错误)
// - 5xxx -> 500 (服务器错误)
func (c Code) HTTPStatus() int {
	switch {
	case c >= 1000 && c < 2000:
		return http.StatusInternalServerError // 500
	case c >= 4000 && c < 4010:
		return http.StatusBadRequest // 400
	case c >= 4010 && c < 4020:
		return http.StatusUnauthorized // 401
	case c >= 4030 && c < 4040:
		return http.StatusForbidden // 403
	case c >= 4040 && c < 4050:
		return http.StatusNotFound // 404
	case c >= 4090 && c < 4100:
		return http.StatusConflict // 409
	case c >= 4290 && c < 4300:
		return http.StatusTooManyRequests // 429
	case c >= 5000 && c < 6000:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusOK // 200
	}
}

// New 创建新的错误
func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf 创建格式化的错误
func Newf(code Code, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
