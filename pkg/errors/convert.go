// Package errors 错误转换工具
//
// 提供标准错误到业务错误的转换功能
package errors

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// FromError 从标准错误转换为业务错误
//
// 初级工程师学习要点：
// - 使用 errors.As 检查错误类型
// - 使用 errors.Is 检查特定错误
// - 自动转换常见的第三方库错误（GORM、Redis）
func FromError(err error) *Error {
	if err == nil {
		return nil
	}

	// 如果已经是 Error 类型，直接返回
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	// GORM 错误转换
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound.WithError(err)
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicate.WithError(err)
	}
	if errors.Is(err, gorm.ErrInvalidTransaction) {
		return ErrDBTxError.WithError(err)
	}

	// Redis 错误转换
	if isRedisError(err) {
		return ErrCacheError.WithError(err)
	}

	// 默认返回内部错误
	return ErrInternalError.WithError(err)
}

// isRedisError 检查是否是 Redis 错误
func isRedisError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "redis") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "i/o timeout")
}

// Is 检查错误是否匹配
//
// 初级工程师学习要点：
// - 包装标准库的 errors.Is
// - 支持错误链检查
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 提取错误类型
//
// 初级工程师学习要点：
// - 包装标准库的 errors.As
// - 支持错误类型断言
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
