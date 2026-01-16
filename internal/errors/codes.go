// Package errors 定义了应用程序的错误码和错误类型
// 使用错误码的好处：
// 1. 前端可以根据错误码做不同的处理（如 40101 跳转登录页）
// 2. 方便国际化（根据错误码显示不同语言的提示）
// 3. 便于日志排查问题
package errors

// ==================== 错误码规范 ====================
// 错误码格式: AABCC
// - AA: 模块代码 (10-99)
// - B: 错误类型 (0=成功, 1=客户端错误, 5=服务器错误)
// - CC: 具体错误 (01-99)
//
// 示例:
// - 40101: 4=HTTP状态码前缀, 01=认证模块, 01=未授权
// - 40001: 4=HTTP状态码前缀, 00=通用模块, 01=参数错误

// ==================== 成功码 ====================
const (
	// CodeSuccess 表示请求成功
	CodeSuccess = 0
)

// ==================== 通用错误码 (400xx) ====================
const (
	// CodeInvalidParams 参数验证失败
	// 例如: 邮箱格式不正确、必填字段为空
	CodeInvalidParams = 40001

	// CodeInvalidRequest 请求格式错误
	// 例如: JSON 解析失败
	CodeInvalidRequest = 40002
)

// ==================== 认证错误码 (401xx) ====================
const (
	// CodeUnauthorized 未授权（未登录或 token 无效）
	CodeUnauthorized = 40101

	// CodeTokenExpired Token 已过期
	CodeTokenExpired = 40102

	// CodeInvalidToken Token 格式错误或被篡改
	CodeInvalidToken = 40103
)

// ==================== 权限错误码 (403xx) ====================
const (
	// CodeForbidden 禁止访问（已登录但无权限）
	CodeForbidden = 40301

	// CodeAccountBlocked 账户被封禁
	CodeAccountBlocked = 40302
)

// ==================== 资源错误码 (404xx) ====================
const (
	// CodeNotFound 资源不存在
	CodeNotFound = 40401

	// CodeUserNotFound 用户不存在
	CodeUserNotFound = 40402

	// CodeAccountNotFound 账户不存在
	CodeAccountNotFound = 40403
)

// ==================== 冲突错误码 (409xx) ====================
const (
	// CodeAlreadyExists 资源已存在
	CodeAlreadyExists = 40901

	// CodeUsernameExists 用户名已被注册
	CodeUsernameExists = 40902

	// CodeEmailExists 邮箱已被注册
	CodeEmailExists = 40903
)

// ==================== 业务错误码 (422xx) ====================
const (
	// CodeInsufficientBalance 余额不足
	CodeInsufficientBalance = 42201

	// CodeCurrencyMismatch 货币类型不匹配
	CodeCurrencyMismatch = 42202

	// CodeSameAccount 不能转账给自己
	CodeSameAccount = 42203

	// CodePasswordWrong 密码错误
	CodePasswordWrong = 42204
)

// ==================== 服务器错误码 (500xx) ====================
const (
	// CodeInternalError 服务器内部错误
	CodeInternalError = 50001

	// CodeDatabaseError 数据库操作失败
	CodeDatabaseError = 50002
)

// codeMessages 存储错误码对应的默认消息
var codeMessages = map[int]string{
	CodeSuccess: "success",

	// 通用错误
	CodeInvalidParams:  "invalid parameters",
	CodeInvalidRequest: "invalid request format",

	// 认证错误
	CodeUnauthorized: "unauthorized",
	CodeTokenExpired: "token expired",
	CodeInvalidToken: "invalid token",

	// 权限错误
	CodeForbidden:      "access forbidden",
	CodeAccountBlocked: "account blocked",

	// 资源错误
	CodeNotFound:        "resource not found",
	CodeUserNotFound:    "user not found",
	CodeAccountNotFound: "account not found",

	// 冲突错误
	CodeAlreadyExists:  "resource already exists",
	CodeUsernameExists: "username already exists",
	CodeEmailExists:    "email already exists",

	// 业务错误
	CodeInsufficientBalance: "insufficient balance",
	CodeCurrencyMismatch:    "currency mismatch",
	CodeSameAccount:         "cannot transfer to same account",
	CodePasswordWrong:       "wrong password",

	// 服务器错误
	CodeInternalError: "internal server error",
	CodeDatabaseError: "database error",
}

// GetMessage 根据错误码获取默认错误消息
func GetMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "unknown error"
}
