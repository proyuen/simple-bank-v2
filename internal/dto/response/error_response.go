package response

import (
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
)

// ErrorResponse 错误响应
// 统一的错误响应格式
type ErrorResponse struct {
	Code    int    `json:"code"`    // 业务错误码
	Message string `json:"message"` // 错误消息
}

// NewErrorResponse 从 AppError 创建错误响应
func NewErrorResponse(err *apperrors.AppError) ErrorResponse {
	return ErrorResponse{
		Code:    err.Code,
		Message: err.Message,
	}
}

// SuccessResponse 成功响应
// 用于不需要返回数据的操作（如删除）
type SuccessResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(message string) SuccessResponse {
	return SuccessResponse{
		Code:    0,
		Message: message,
	}
}
