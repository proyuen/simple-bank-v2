package errors

import (
	"fmt"
	"net/http"
)

// AppError 是应用程序的统一错误类型
// 包含错误码、HTTP 状态码和错误消息
type AppError struct {
	Code       int    `json:"code"`    // 业务错误码
	Message    string `json:"message"` // 错误消息
	HTTPStatus int    `json:"-"`       // HTTP 状态码（不输出到 JSON）
}

// Error 实现 error 接口
// 这使得 AppError 可以像普通 error 一样使用
func (e *AppError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// ==================== 错误构造函数 ====================
// 这些函数用于快速创建常见的错误类型

// New 创建一个新的 AppError
// 使用错误码的默认消息
func New(code int) *AppError {
	return &AppError{
		Code:       code,
		Message:    GetMessage(code),
		HTTPStatus: codeToHTTPStatus(code),
	}
}

// NewWithMessage 创建一个带自定义消息的 AppError
// 当需要提供更具体的错误信息时使用
func NewWithMessage(code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: codeToHTTPStatus(code),
	}
}

// Wrap 包装一个已有的 error 为 AppError
// 常用于包装数据库错误等底层错误
func Wrap(code int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    err.Error(),
		HTTPStatus: codeToHTTPStatus(code),
	}
}

// ==================== 常用错误快捷函数 ====================

// ErrInvalidParams 返回参数验证错误
func ErrInvalidParams(message string) *AppError {
	return NewWithMessage(CodeInvalidParams, message)
}

// ErrUnauthorized 返回未授权错误
func ErrUnauthorized() *AppError {
	return New(CodeUnauthorized)
}

// ErrForbidden 返回禁止访问错误
func ErrForbidden() *AppError {
	return New(CodeForbidden)
}

// ErrNotFound 返回资源不存在错误
func ErrNotFound(resource string) *AppError {
	return NewWithMessage(CodeNotFound, resource+" not found")
}

// ErrUserNotFound 返回用户不存在错误
func ErrUserNotFound() *AppError {
	return New(CodeUserNotFound)
}

// ErrAccountNotFound 返回账户不存在错误
func ErrAccountNotFound() *AppError {
	return New(CodeAccountNotFound)
}

// ErrUsernameExists 返回用户名已存在错误
func ErrUsernameExists() *AppError {
	return New(CodeUsernameExists)
}

// ErrEmailExists 返回邮箱已存在错误
func ErrEmailExists() *AppError {
	return New(CodeEmailExists)
}

// ErrPasswordWrong 返回密码错误
func ErrPasswordWrong() *AppError {
	return New(CodePasswordWrong)
}

// ErrInsufficientBalance 返回余额不足错误
func ErrInsufficientBalance() *AppError {
	return New(CodeInsufficientBalance)
}

// ErrCurrencyMismatch 返回货币类型不匹配错误
func ErrCurrencyMismatch() *AppError {
	return New(CodeCurrencyMismatch)
}

// ErrInternalServer 返回服务器内部错误
func ErrInternalServer() *AppError {
	return New(CodeInternalError)
}

// ErrDatabase 包装数据库错误
func ErrDatabase(err error) *AppError {
	return Wrap(CodeDatabaseError, err)
}

// ==================== 辅助函数 ====================

// codeToHTTPStatus 根据业务错误码推断 HTTP 状态码
// 规则: 错误码前三位对应 HTTP 状态码
// 例如: 40101 → 401, 50001 → 500
func codeToHTTPStatus(code int) int {
	if code == CodeSuccess {
		return http.StatusOK
	}

	// 提取前三位作为 HTTP 状态码
	httpCode := code / 100

	// 验证是否为有效的 HTTP 状态码
	switch httpCode {
	case 400, 401, 403, 404, 409, 422:
		return httpCode
	case 500, 502, 503:
		return httpCode
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError 检查 error 是否为 AppError 类型
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError 将 error 转换为 AppError
// 如果不是 AppError，返回一个内部错误
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return ErrInternalServer()
}
