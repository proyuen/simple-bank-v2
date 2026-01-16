package request

// CreateTransferRequest 创建转账请求
// 用于: POST /api/v1/transfers
type CreateTransferRequest struct {
	// FromAccountID 转出账户ID
	FromAccountID uint `json:"from_account_id" binding:"required,min=1"`

	// ToAccountID 转入账户ID
	ToAccountID uint `json:"to_account_id" binding:"required,min=1"`

	// Amount 转账金额 (单位: 分)
	// 例如: 1000 = $10.00
	Amount int64 `json:"amount" binding:"required,gt=0"`

	// Currency 货币类型
	// 必须与两个账户的货币类型匹配
	Currency string `json:"currency" binding:"required,oneof=USD EUR CNY"`
}

// ListTransfersRequest 获取转账记录请求
// 用于: GET /api/v1/transfers
type ListTransfersRequest struct {
	AccountID uint `form:"account_id" binding:"required,min=1"`
	PageID    int  `form:"page_id" binding:"required,min=1"`
	PageSize  int  `form:"page_size" binding:"required,min=5,max=100"`
}

// ListEntriesRequest 获取账目记录请求
// 用于: GET /api/v1/accounts/:id/entries
type ListEntriesRequest struct {
	AccountID uint `uri:"id" binding:"required,min=1"`
	PageID    int  `form:"page_id" binding:"required,min=1"`
	PageSize  int  `form:"page_size" binding:"required,min=5,max=100"`
}
