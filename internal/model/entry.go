package model

import (
	"time"
)

// Entry 账目记录模型 - 对应 entries 表
//
// 用途: 记录每一笔资金变动
//
// Amount 字段说明:
//   - 正数: 入账 (例如: +100 表示收到 $1.00)
//   - 负数: 出账 (例如: -100 表示支出 $1.00)
//
// 示例:
//   转账 $10 从账户A到账户B:
//   - Entry 1: AccountID=A, Amount=-1000 (出账)
//   - Entry 2: AccountID=B, Amount=+1000 (入账)
type Entry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AccountID uint      `gorm:"not null;index" json:"account_id"` // 关联的账户ID
	Amount    int64     `gorm:"not null" json:"amount"`           // 金额(正=入账, 负=出账)
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	Account Account `gorm:"foreignKey:AccountID" json:"-"`
}

// TableName 指定表名
func (Entry) TableName() string {
	return "entries"
}

// IsCredit 判断是否为入账
func (e *Entry) IsCredit() bool {
	return e.Amount > 0
}

// IsDebit 判断是否为出账
func (e *Entry) IsDebit() bool {
	return e.Amount < 0
}
