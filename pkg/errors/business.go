// Package errors 业务错误定义
//
// 业务错误包括客户端错误和服务器错误
// 这些错误会被转换为 HTTP 响应返回给客户端
package errors

var (
	// ==================== 参数错误 (400x) ====================
	ErrInvalidParams = &Error{
		Code:    CodeInvalidParams,
		Message: GetMessage(CodeInvalidParams),
	}

	ErrMissingParams = &Error{
		Code:    CodeMissingParams,
		Message: GetMessage(CodeMissingParams),
	}

	ErrInvalidFormat = &Error{
		Code:    CodeInvalidFormat,
		Message: GetMessage(CodeInvalidFormat),
	}

	// ==================== 认证错误 (401x) ====================
	ErrAuthError = &Error{
		Code:    CodeAuthError,
		Message: GetMessage(CodeAuthError),
	}

	ErrUnauthorized = &Error{
		Code:    CodeUnauthorized,
		Message: GetMessage(CodeUnauthorized),
	}

	ErrTokenExpired = &Error{
		Code:    CodeTokenExpired,
		Message: GetMessage(CodeTokenExpired),
	}

	ErrTokenInvalid = &Error{
		Code:    CodeTokenInvalid,
		Message: GetMessage(CodeTokenInvalid),
	}

	// ==================== 权限错误 (403x) ====================
	ErrForbidden = &Error{
		Code:    CodeForbidden,
		Message: GetMessage(CodeForbidden),
	}

	ErrAccessDenied = &Error{
		Code:    CodeAccessDenied,
		Message: GetMessage(CodeAccessDenied),
	}

	// ==================== 资源错误 (404x) ====================
	ErrNotFound = &Error{
		Code:    CodeNotFound,
		Message: GetMessage(CodeNotFound),
	}

	ErrUserNotFound = &Error{
		Code:    CodeUserNotFound,
		Message: GetMessage(CodeUserNotFound),
	}

	ErrOrderNotFound = &Error{
		Code:    CodeOrderNotFound,
		Message: GetMessage(CodeOrderNotFound),
	}

	// ==================== 冲突错误 (409x) ====================
	ErrConflict = &Error{
		Code:    CodeConflict,
		Message: GetMessage(CodeConflict),
	}

	ErrDuplicate = &Error{
		Code:    CodeDuplicate,
		Message: GetMessage(CodeDuplicate),
	}

	// ==================== 限流错误 (429x) ====================
	ErrTooManyRequests = &Error{
		Code:    CodeTooManyRequests,
		Message: GetMessage(CodeTooManyRequests),
	}

	ErrRateLimitExceeded = &Error{
		Code:    CodeRateLimitExceeded,
		Message: GetMessage(CodeRateLimitExceeded),
	}

	// ==================== 内部错误 (500x) ====================
	ErrInternalError = &Error{
		Code:    CodeInternalError,
		Message: GetMessage(CodeInternalError),
	}

	ErrPanic = &Error{
		Code:    CodePanic,
		Message: GetMessage(CodePanic),
	}

	// ==================== 数据库错误 (501x) ====================
	ErrDBError = &Error{
		Code:    CodeDBError,
		Message: GetMessage(CodeDBError),
	}

	ErrDBQueryError = &Error{
		Code:    CodeDBQueryError,
		Message: GetMessage(CodeDBQueryError),
	}

	ErrDBTxError = &Error{
		Code:    CodeDBTxError,
		Message: GetMessage(CodeDBTxError),
	}

	// ==================== 缓存错误 (502x) ====================
	ErrCacheError = &Error{
		Code:    CodeCacheError,
		Message: GetMessage(CodeCacheError),
	}

	ErrCacheGetError = &Error{
		Code:    CodeCacheGetError,
		Message: GetMessage(CodeCacheGetError),
	}

	ErrCacheSetError = &Error{
		Code:    CodeCacheSetError,
		Message: GetMessage(CodeCacheSetError),
	}

	// ==================== RPC 错误 (503x) ====================
	ErrRPCError = &Error{
		Code:    CodeRPCError,
		Message: GetMessage(CodeRPCError),
	}

	ErrRPCTimeout = &Error{
		Code:    CodeRPCTimeout,
		Message: GetMessage(CodeRPCTimeout),
	}

	// ==================== 第三方服务错误 (504x) ====================
	ErrThirdPartyError = &Error{
		Code:    CodeThirdPartyError,
		Message: GetMessage(CodeThirdPartyError),
	}

	ErrPaymentFailed = &Error{
		Code:    CodePaymentFailed,
		Message: GetMessage(CodePaymentFailed),
	}

	ErrSMSFailed = &Error{
		Code:    CodeSMSFailed,
		Message: GetMessage(CodeSMSFailed),
	}
)
