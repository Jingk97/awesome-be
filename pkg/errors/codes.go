// Package errors 错误码定义
package errors

// 错误码分类：
// - 0: 成功
// - 1xxx: 系统错误（启动、配置、依赖）
// - 4xxx: 客户端错误（参数、认证、权限）
// - 5xxx: 服务器错误（内部、数据库、第三方）

const (
	// ==================== 成功 ====================
	CodeSuccess Code = 0

	// ==================== 系统错误 (1xxx) ====================
	// 配置错误 (10xx)
	CodeConfigLoadFailed     Code = 1001 // 配置加载失败
	CodeConfigParseFailed    Code = 1002 // 配置解析失败
	CodeConfigValidateFailed Code = 1003 // 配置验证失败

	// 数据库错误 (11xx)
	CodeDBConnectFailed Code = 1101 // 数据库连接失败
	CodeDBPingFailed    Code = 1102 // 数据库 PING 失败
	CodeDBMigrateFailed Code = 1103 // 数据库迁移失败

	// Redis 错误 (12xx)
	CodeRedisConnectFailed Code = 1201 // Redis 连接失败
	CodeRedisPingFailed    Code = 1202 // Redis PING 失败

	// 依赖服务错误 (13xx)
	CodeServiceUnavailable Code = 1301 // 依赖服务不可用
	CodeServiceTimeout     Code = 1302 // 依赖服务超时

	// 启动错误 (14xx)
	CodePortBindFailed    Code = 1401 // 端口绑定失败
	CodeServerStartFailed Code = 1402 // 服务启动失败

	// ==================== 客户端错误 (4xxx) ====================
	// 参数错误 (400x)
	CodeInvalidParams Code = 4001 // 参数错误
	CodeMissingParams Code = 4002 // 缺少参数
	CodeInvalidFormat Code = 4003 // 格式错误

	// 认证错误 (401x)
	CodeAuthError    Code = 4011 // 认证失败
	CodeUnauthorized Code = 4012 // 未认证
	CodeTokenExpired Code = 4013 // Token 过期
	CodeTokenInvalid Code = 4014 // Token 无效

	// 权限错误 (403x)
	CodeForbidden    Code = 4031 // 无权限
	CodeAccessDenied Code = 4032 // 访问被拒绝

	// 资源错误 (404x)
	CodeNotFound      Code = 4041 // 资源不存在
	CodeUserNotFound  Code = 4042 // 用户不存在
	CodeOrderNotFound Code = 4043 // 订单不存在

	// 冲突错误 (409x)
	CodeConflict  Code = 4091 // 资源冲突
	CodeDuplicate Code = 4092 // 资源重复

	// 限流错误 (429x)
	CodeTooManyRequests   Code = 4291 // 请求过多
	CodeRateLimitExceeded Code = 4292 // 超过限流

	// ==================== 服务器错误 (5xxx) ====================
	// 内部错误 (500x)
	CodeInternalError Code = 5001 // 内部错误
	CodePanic         Code = 5002 // Panic 错误

	// 数据库错误 (501x)
	CodeDBError      Code = 5011 // 数据库错误
	CodeDBQueryError Code = 5012 // 查询失败
	CodeDBTxError    Code = 5013 // 事务失败

	// 缓存错误 (502x)
	CodeCacheError    Code = 5021 // 缓存错误
	CodeCacheGetError Code = 5022 // 缓存获取失败
	CodeCacheSetError Code = 5023 // 缓存设置失败

	// RPC 错误 (503x)
	CodeRPCError   Code = 5031 // RPC 调用错误
	CodeRPCTimeout Code = 5032 // RPC 超时

	// 第三方服务错误 (504x)
	CodeThirdPartyError Code = 5041 // 第三方服务错误
	CodePaymentFailed   Code = 5042 // 支付失败
	CodeSMSFailed       Code = 5043 // 短信发送失败
)

// 错误消息映射
var messages = map[Code]string{
	CodeSuccess: "success",

	// 系统错误
	CodeConfigLoadFailed:     "配置加载失败",
	CodeConfigParseFailed:    "配置解析失败",
	CodeConfigValidateFailed: "配置验证失败",
	CodeDBConnectFailed:      "数据库连接失败",
	CodeDBPingFailed:         "数据库 PING 失败",
	CodeDBMigrateFailed:      "数据库迁移失败",
	CodeRedisConnectFailed:   "Redis 连接失败",
	CodeRedisPingFailed:      "Redis PING 失败",
	CodeServiceUnavailable:   "依赖服务不可用",
	CodeServiceTimeout:       "依赖服务超时",
	CodePortBindFailed:       "端口绑定失败",
	CodeServerStartFailed:    "服务启动失败",

	// 客户端错误
	CodeInvalidParams:     "参数错误",
	CodeMissingParams:     "缺少参数",
	CodeInvalidFormat:     "格式错误",
	CodeAuthError:         "认证失败",
	CodeUnauthorized:      "未认证",
	CodeTokenExpired:      "Token 过期",
	CodeTokenInvalid:      "Token 无效",
	CodeForbidden:         "无权限",
	CodeAccessDenied:      "访问被拒绝",
	CodeNotFound:          "资源不存在",
	CodeUserNotFound:      "用户不存在",
	CodeOrderNotFound:     "订单不存在",
	CodeConflict:          "资源冲突",
	CodeDuplicate:         "资源重复",
	CodeTooManyRequests:   "请求过多",
	CodeRateLimitExceeded: "超过限流",

	// 服务器错误
	CodeInternalError:   "内部错误",
	CodePanic:           "系统异常",
	CodeDBError:         "数据库错误",
	CodeDBQueryError:    "查询失败",
	CodeDBTxError:       "事务失败",
	CodeCacheError:      "缓存错误",
	CodeCacheGetError:   "缓存获取失败",
	CodeCacheSetError:   "缓存设置失败",
	CodeRPCError:        "RPC 调用错误",
	CodeRPCTimeout:      "RPC 超时",
	CodeThirdPartyError: "第三方服务错误",
	CodePaymentFailed:   "支付失败",
	CodeSMSFailed:       "短信发送失败",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code Code) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return "未知错误"
}
