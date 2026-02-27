// Package errors 系统错误定义
//
// 系统错误用于应用启动和运行时的系统级错误
// 这些错误通常会导致应用无法正常运行，需要立即退出
package errors

var (
	// ==================== 配置错误 (10xx) ====================
	ErrConfigLoadFailed = &Error{
		Code:    CodeConfigLoadFailed,
		Message: GetMessage(CodeConfigLoadFailed),
	}

	ErrConfigParseFailed = &Error{
		Code:    CodeConfigParseFailed,
		Message: GetMessage(CodeConfigParseFailed),
	}

	ErrConfigValidateFailed = &Error{
		Code:    CodeConfigValidateFailed,
		Message: GetMessage(CodeConfigValidateFailed),
	}

	// ==================== 数据库错误 (11xx) ====================
	ErrDBConnectFailed = &Error{
		Code:    CodeDBConnectFailed,
		Message: GetMessage(CodeDBConnectFailed),
	}

	ErrDBPingFailed = &Error{
		Code:    CodeDBPingFailed,
		Message: GetMessage(CodeDBPingFailed),
	}

	ErrDBMigrateFailed = &Error{
		Code:    CodeDBMigrateFailed,
		Message: GetMessage(CodeDBMigrateFailed),
	}

	// ==================== Redis 错误 (12xx) ====================
	ErrRedisConnectFailed = &Error{
		Code:    CodeRedisConnectFailed,
		Message: GetMessage(CodeRedisConnectFailed),
	}

	ErrRedisPingFailed = &Error{
		Code:    CodeRedisPingFailed,
		Message: GetMessage(CodeRedisPingFailed),
	}

	// ==================== 依赖服务错误 (13xx) ====================
	ErrServiceUnavailable = &Error{
		Code:    CodeServiceUnavailable,
		Message: GetMessage(CodeServiceUnavailable),
	}

	ErrServiceTimeout = &Error{
		Code:    CodeServiceTimeout,
		Message: GetMessage(CodeServiceTimeout),
	}

	// ==================== 启动错误 (14xx) ====================
	ErrPortBindFailed = &Error{
		Code:    CodePortBindFailed,
		Message: GetMessage(CodePortBindFailed),
	}

	ErrServerStartFailed = &Error{
		Code:    CodeServerStartFailed,
		Message: GetMessage(CodeServerStartFailed),
	}
)
