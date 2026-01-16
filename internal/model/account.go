package model

import (
	"time"

	"gorm.io/gorm"
)

// Account 银行账户模型 - 对应 accounts 表
//
// 重要字段说明:
//   - Balance: 以"分"为单位存储，避免浮点数精度问题
//     例如: $100.50 存储为 10050
//   - Currency: 货币代码 (USD, EUR, CNY 等)
//   - Owner: 关联到 users.username
//
// 业务规则:
//   - 同一用户同一货币只能有一个账户 (由数据库唯一索引保证)
//   - 余额不能为负数 (由业务逻辑保证)
type Account struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Owner     string         `gorm:"not null;index;size:255" json:"owner"`           // 账户所有者(用户名)
	Balance   int64          `gorm:"not null;default:0" json:"balance"`              // 余额(单位:分)
	Currency  string         `gorm:"not null;size:3" json:"currency"`                // 货币类型
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联关系
	User     User    `gorm:"foreignKey:Owner;references:Username" json:"-"`
	Entries  []Entry `gorm:"foreignKey:AccountID" json:"entries,omitempty"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}

// BalanceInDollars 返回以美元为单位的余额 (仅用于显示)
// 例如: Balance = 10050 → 返回 100.50
func (a *Account) BalanceInDollars() float64 {
	return float64(a.Balance) / 100
}
