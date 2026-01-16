package model

import (
	"time"
)

// Transfer 转账记录模型 - 对应 transfers 表
//
// 用途: 记录账户间的转账操作
//
// 业务规则:
//   - Amount 必须为正数
//   - FromAccountID 和 ToAccountID 必须不同
//   - 两个账户的货币类型必须相同
//
// 转账流程:
//   1. 检查转出账户余额充足
//   2. 创建 Transfer 记录
//   3. 创建两条 Entry 记录 (一出一入)
//   4. 更新两个账户余额
//   以上操作在一个数据库事务中完成
type Transfer struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FromAccountID uint      `gorm:"not null;index" json:"from_account_id"` // 转出账户ID
	ToAccountID   uint      `gorm:"not null;index" json:"to_account_id"`   // 转入账户ID
	Amount        int64     `gorm:"not null" json:"amount"`                // 转账金额(必须>0)
	CreatedAt     time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	FromAccount Account `gorm:"foreignKey:FromAccountID" json:"from_account,omitempty"`
	ToAccount   Account `gorm:"foreignKey:ToAccountID" json:"to_account,omitempty"`
}

// TableName 指定表名
func (Transfer) TableName() string {
	return "transfers"
}
