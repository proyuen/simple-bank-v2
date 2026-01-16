package response

import "time"

// AccountResponse 账户信息响应
type AccountResponse struct {
	ID        uint      `json:"id"`
	Owner     string    `json:"owner"`
	Balance   int64     `json:"balance"`   // 余额(单位:分)
	Currency  string    `json:"currency"`  // 货币类型
	CreatedAt time.Time `json:"created_at"`
}

// TransferResponse 转账记录响应
type TransferResponse struct {
	ID            uint      `json:"id"`
	FromAccountID uint      `json:"from_account_id"`
	ToAccountID   uint      `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}

// EntryResponse 账目记录响应
type EntryResponse struct {
	ID        uint      `json:"id"`
	AccountID uint      `json:"account_id"`
	Amount    int64     `json:"amount"` // 正数=入账, 负数=出账
	CreatedAt time.Time `json:"created_at"`
}

// TransferResultResponse 转账结果响应
// 包含完整的转账信息
type TransferResultResponse struct {
	Transfer    TransferResponse `json:"transfer"`
	FromAccount AccountResponse  `json:"from_account"`
	ToAccount   AccountResponse  `json:"to_account"`
	FromEntry   EntryResponse    `json:"from_entry"`
	ToEntry     EntryResponse    `json:"to_entry"`
}
