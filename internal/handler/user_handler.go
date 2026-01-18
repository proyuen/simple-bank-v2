// Package handler 提供 HTTP 请求处理器
// Handler 负责:
//  1. 解析和验证 HTTP 请求
//  2. 调用 Service 层处理业务逻辑
//  3. 返回 HTTP 响应
//
// Handler 不应包含任何业务逻辑，只做请求/响应处理
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/proyuen/simple-bank-v2/internal/dto/request"
	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/service"
)

// ==================== Handler 结构体 ====================

// UserHandler 处理用户相关的 HTTP 请求
type UserHandler struct {
	// userService 是用户服务层的引用
	// 使用接口而非具体类型，便于测试时注入 mock
	userService *service.UserService
}

// NewUserHandler 创建 UserHandler 实例
//
// 参数:
//   - userService: 用户服务层实例
//
// 返回:
//   - *UserHandler: Handler 实例
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ==================== Handler 方法 ====================

// CreateUser 处理用户注册请求
//
// 路由: POST /api/v1/users
// 请求体: CreateUserRequest (JSON)
// 响应: 201 Created + UserResponse
//
// 错误响应:
//   - 400 Bad Request: 参数验证失败
//   - 409 Conflict: 用户名或邮箱已存在
//   - 500 Internal Server Error: 服务器错误
//
// @Summary 用户注册
// @Description 创建一个新用户
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.CreateUserRequest true "用户注册信息"
// @Success 201 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Step 1: 绑定并验证请求体
	// ShouldBindJSON 会自动根据 binding 标签验证字段
	var req request.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 绑定失败，返回参数错误
		// handleValidationError 会解析具体的验证错误
		h.handleValidationError(c, err)
		return
	}

	// Step 2: 调用 Service 创建用户
	// 所有业务逻辑在 Service 层处理
	userResp, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		// 处理 Service 返回的错误
		h.handleError(c, err)
		return
	}

	// Step 3: 返回成功响应
	// 201 Created 表示资源创建成功
	c.JSON(http.StatusCreated, userResp)
}

// LoginUser 处理用户登录请求
//
// 路由: POST /api/v1/users/login
// 请求体: LoginUserRequest (JSON)
// 响应: 200 OK + LoginResponse
//
// 错误响应:
//   - 400 Bad Request: 参数验证失败
//   - 422 Unprocessable Entity: 密码错误
//
// @Summary 用户登录
// @Description 用户登录获取 Token
// @Tags users
// @Accept json
// @Produce json
// @Param request body request.LoginUserRequest true "登录信息"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Router /users/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	// Step 1: 绑定并验证请求体
	var req request.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 2: 获取客户端信息
	// 这些信息用于会话管理和安全审计
	userAgent := c.GetHeader("User-Agent") // 客户端标识 (浏览器/App等)
	clientIP := c.ClientIP()               // 客户端 IP 地址

	// Step 3: 调用 Service 处理登录
	loginResp, err := h.userService.LoginUser(c.Request.Context(), &req, userAgent, clientIP)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 4: 返回成功响应
	c.JSON(http.StatusOK, loginResp)
}

// RefreshToken 处理刷新 Token 请求
//
// 路由: POST /api/v1/tokens/renew
// 请求体: RefreshTokenRequest (JSON)
// 响应: 200 OK + RefreshTokenResponse
//
// 工作流程:
//  1. 验证 Refresh Token 有效性
//  2. 检查会话是否被封禁
//  3. 生成新的 Access Token
//
// @Summary 刷新 Token
// @Description 使用 Refresh Token 获取新的 Access Token
// @Tags tokens
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} response.RefreshTokenResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /tokens/renew [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	// Step 1: 绑定并验证请求体
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 2: 调用 Service 刷新 Token
	refreshResp, err := h.userService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 3: 返回成功响应
	c.JSON(http.StatusOK, refreshResp)
}

// ==================== 错误处理辅助方法 ====================

// handleError 统一处理 Service 层返回的错误
//
// 将 AppError 转换为 HTTP 响应
// 如果不是 AppError，返回 500 内部错误
func (h *UserHandler) handleError(c *gin.Context, err error) {
	// 尝试将错误转换为 AppError
	appErr := apperrors.AsAppError(err)

	// 使用 AppError 中的 HTTP 状态码
	c.JSON(appErr.HTTPStatus, response.NewErrorResponse(appErr))
}

// handleValidationError 处理请求参数验证错误
//
// Gin 的 binding 验证失败时调用此方法
// 返回 400 Bad Request 和详细的错误信息
func (h *UserHandler) handleValidationError(c *gin.Context, err error) {
	// 创建参数验证错误
	appErr := apperrors.ErrInvalidParams(err.Error())
	c.JSON(http.StatusBadRequest, response.NewErrorResponse(appErr))
}
