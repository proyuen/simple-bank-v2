package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/proyuen/simple-bank-v2/internal/dto/request"
	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/middleware"
	"github.com/proyuen/simple-bank-v2/internal/service"
)

// ==================== Handler 结构体 ====================

// AccountHandler 处理账户相关的 HTTP 请求
type AccountHandler struct {
	accountService *service.AccountService
}

// NewAccountHandler 创建 AccountHandler 实例
func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// ==================== Handler 方法 ====================

// CreateAccount 处理创建账户请求
//
// 路由: POST /api/v1/accounts (需要认证)
// 请求体: CreateAccountRequest (JSON)
// 响应: 201 Created + AccountResponse
//
// 业务规则:
//   - 只有已登录用户可以创建账户
//   - 账户所有者自动设置为当前登录用户
//   - 同一用户同一货币只能有一个账户
//
// @Summary 创建账户
// @Description 为当前用户创建一个新的银行账户
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body request.CreateAccountRequest true "账户信息"
// @Success 201 {object} response.AccountResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	// Step 1: 获取当前登录用户
	// MustGetAuthPayload 从 Context 获取 AuthMiddleware 存入的 payload
	// 由于这个路由受 AuthMiddleware 保护，payload 一定存在
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证请求体
	var req request.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 调用 Service 创建账户
	// 第一个参数是账户所有者 (当前登录用户的用户名)
	accountResp, err := h.accountService.CreateAccount(c.Request.Context(), payload.Username, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 4: 返回成功响应
	c.JSON(http.StatusCreated, accountResp)
}

// GetAccount 处理获取账户详情请求
//
// 路由: GET /api/v1/accounts/:id (需要认证)
// 参数: id (URL 路径参数)
// 响应: 200 OK + AccountResponse
//
// 业务规则:
//   - 只能查看自己的账户
//   - 查看他人账户返回 403 Forbidden
//
// @Summary 获取账户
// @Description 获取指定账户的详细信息
// @Tags accounts
// @Produce json
// @Param id path int true "账户ID"
// @Success 200 {object} response.AccountResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id} [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	// Step 1: 获取当前登录用户
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证 URL 参数
	// ShouldBindUri 解析 URL 中的参数 (如 :id)
	var req request.GetAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 调用 Service 获取账户
	// Service 会验证账户是否属于当前用户
	accountResp, err := h.accountService.GetAccount(c.Request.Context(), payload.Username, req.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 4: 返回成功响应
	c.JSON(http.StatusOK, accountResp)
}

// ListAccounts 处理获取账户列表请求
//
// 路由: GET /api/v1/accounts (需要认证)
// 参数: page_id, page_size (Query 参数)
// 响应: 200 OK + ListResponse[AccountResponse]
//
// 业务规则:
//   - 只返回当前用户的账户
//   - 支持分页
//
// @Summary 获取账户列表
// @Description 获取当前用户的所有账户（分页）
// @Tags accounts
// @Produce json
// @Param page_id query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(5) maximum(100)
// @Success 200 {object} response.ListResponse[response.AccountResponse]
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts [get]
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	// Step 1: 获取当前登录用户
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证 Query 参数
	// ShouldBindQuery 解析 URL 中的查询参数 (如 ?page_id=1&page_size=10)
	var req request.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 调用 Service 获取账户列表
	listResp, err := h.accountService.ListAccounts(c.Request.Context(), payload.Username, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 4: 返回成功响应
	c.JSON(http.StatusOK, listResp)
}

// ==================== 错误处理辅助方法 ====================

// handleError 统一处理 Service 层返回的错误
func (h *AccountHandler) handleError(c *gin.Context, err error) {
	appErr := apperrors.AsAppError(err)
	c.JSON(appErr.HTTPStatus, response.NewErrorResponse(appErr))
}

// handleValidationError 处理请求参数验证错误
func (h *AccountHandler) handleValidationError(c *gin.Context, err error) {
	appErr := apperrors.ErrInvalidParams(err.Error())
	c.JSON(http.StatusBadRequest, response.NewErrorResponse(appErr))
}
