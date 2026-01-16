// Package model 定义了与数据库表对应的 GORM 模型
// 这些结构体通过 gorm 标签与数据库表字段映射
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型 - 对应 users 表
//
// 重要字段说明:
//   - HashedPassword: 存储 bcrypt 加密后的密码，永远不要存储明文密码
//   - PasswordChangedAt: 用于强制用户在密码修改后重新登录
//
// 关联关系:
//   - User 1:N Accounts (一个用户可以有多个账户)
//   - User 1:N Sessions (一个用户可以有多个会话)
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Username          string         `gorm:"uniqueIndex;not null;size:255" json:"username"`
	HashedPassword    string         `gorm:"not null;size:255" json:"-"` // json:"-" 不输出到 JSON
	FullName          string         `gorm:"not null;size:255" json:"full_name"`
	Email             string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordChangedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"password_changed_at"`
	CreatedAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"` // 软删除

	// 关联关系 (不创建数据库字段，仅用于 GORM 预加载)
	Accounts []Account `gorm:"foreignKey:Owner;references:Username" json:"accounts,omitempty"`
	Sessions []Session `gorm:"foreignKey:Username;references:Username" json:"-"`
}

// TableName 指定表名 (GORM 默认会将 User 转为 users)
func (User) TableName() string {
	return "users"
}
