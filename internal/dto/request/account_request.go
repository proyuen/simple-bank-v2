package request

// CreateAccountRequest 创建账户请求
// 用于: POST /api/v1/accounts
type CreateAccountRequest struct {
	// Currency 货币类型
	// 规则: 必填, 必须是支持的货币代码
	Currency string `json:"currency" binding:"required,oneof=USD EUR CNY"`
}

// GetAccountRequest 获取账户请求
// 用于: GET /api/v1/accounts/:id
type GetAccountRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// ListAccountsRequest 获取账户列表请求
// 用于: GET /api/v1/accounts
type ListAccountsRequest struct {
	PageID   int `form:"page_id" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=5,max=100"`
}
