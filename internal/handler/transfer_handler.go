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

// TransferHandler 处理转账相关的 HTTP 请求
type TransferHandler struct {
	transferService *service.TransferService
}

// NewTransferHandler 创建 TransferHandler 实例
func NewTransferHandler(transferService *service.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

// ==================== Handler 方法 ====================

// CreateTransfer 处理转账请求
//
// 路由: POST /api/v1/transfers (需要认证)
// 请求体: CreateTransferRequest (JSON)
// 响应: 201 Created + TransferResponse
//
// 业务规则:
//   - 只能从自己的账户转出
//   - 两个账户的货币类型必须相同
//   - 转出账户余额必须充足
//   - 转账在数据库事务中完成
//
// 事务中的操作:
//  1. 创建 Transfer 记录
//  2. 创建两条 Entry 记录 (一出一入)
//  3. 更新两个账户的余额
//
// @Summary 创建转账
// @Description 从一个账户转账到另一个账户
// @Tags transfers
// @Accept json
// @Produce json
// @Param request body request.CreateTransferRequest true "转账信息"
// @Success 201 {object} response.TransferResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 422 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /transfers [post]
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
	// Step 1: 获取当前登录用户
	// 用于验证转出账户的所有权
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证请求体
	var req request.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 额外验证 - 不能转账给自己
	if req.FromAccountID == req.ToAccountID {
		appErr := apperrors.NewWithMessage(apperrors.CodeSameAccount, "cannot transfer to same account")
		c.JSON(http.StatusUnprocessableEntity, response.NewErrorResponse(appErr))
		return
	}

	// Step 4: 调用 Service 执行转账
	// Service 会处理:
	//   - 验证账户所有权
	//   - 验证货币类型
	//   - 验证余额
	//   - 在事务中执行转账
	transferResp, err := h.transferService.CreateTransfer(c.Request.Context(), payload.Username, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 5: 返回成功响应
	c.JSON(http.StatusCreated, transferResp)
}

// ListTransfers 处理获取转账记录请求
//
// 路由: GET /api/v1/transfers (需要认证)
// 参数: account_id, page_id, page_size (Query 参数)
// 响应: 200 OK + ListResponse[TransferResponse]
//
// 业务规则:
//   - 只能查看自己账户的转账记录
//   - 包括转入和转出的记录
//
// @Summary 获取转账记录
// @Description 获取指定账户的转账记录（分页）
// @Tags transfers
// @Produce json
// @Param account_id query int true "账户ID"
// @Param page_id query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(5) maximum(100)
// @Success 200 {object} response.ListResponse[response.TransferResponse]
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /transfers [get]
func (h *TransferHandler) ListTransfers(c *gin.Context) {
	// Step 1: 获取当前登录用户
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证 Query 参数
	var req request.ListTransfersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 构造分页参数
	paginationReq := &request.PaginationRequest{
		PageID:   req.PageID,
		PageSize: req.PageSize,
	}

	// Step 4: 调用 Service 获取转账记录
	// Service 会验证账户所有权
	listResp, err := h.transferService.ListTransfers(c.Request.Context(), payload.Username, req.AccountID, paginationReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 5: 返回成功响应
	c.JSON(http.StatusOK, listResp)
}

// ListEntries 处理获取账目记录请求
//
// 路由: GET /api/v1/accounts/:id/entries (需要认证)
// 参数: id (URL 路径参数), page_id, page_size (Query 参数)
// 响应: 200 OK + ListResponse[EntryResponse]
//
// 业务规则:
//   - 只能查看自己账户的账目记录
//   - Entry 记录每一笔资金变动 (入账/出账)
//
// @Summary 获取账目记录
// @Description 获取指定账户的账目记录（分页）
// @Tags entries
// @Produce json
// @Param id path int true "账户ID"
// @Param page_id query int true "页码" minimum(1)
// @Param page_size query int true "每页条数" minimum(5) maximum(100)
// @Success 200 {object} response.ListResponse[response.EntryResponse]
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /accounts/{id}/entries [get]
func (h *TransferHandler) ListEntries(c *gin.Context) {
	// Step 1: 获取当前登录用户
	payload := middleware.MustGetAuthPayload(c)

	// Step 2: 绑定并验证 URL 参数和 Query 参数
	var uriReq request.GetAccountRequest
	if err := c.ShouldBindUri(&uriReq); err != nil {
		h.handleValidationError(c, err)
		return
	}

	var queryReq request.PaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Step 3: 调用 Service 获取账目记录
	listResp, err := h.transferService.ListEntries(c.Request.Context(), payload.Username, uriReq.ID, &queryReq)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// Step 4: 返回成功响应
	c.JSON(http.StatusOK, listResp)
}

// ==================== 错误处理辅助方法 ====================

// handleError 统一处理 Service 层返回的错误
func (h *TransferHandler) handleError(c *gin.Context, err error) {
	appErr := apperrors.AsAppError(err)
	c.JSON(appErr.HTTPStatus, response.NewErrorResponse(appErr))
}

// handleValidationError 处理请求参数验证错误
func (h *TransferHandler) handleValidationError(c *gin.Context, err error) {
	appErr := apperrors.ErrInvalidParams(err.Error())
	c.JSON(http.StatusBadRequest, response.NewErrorResponse(appErr))
}
