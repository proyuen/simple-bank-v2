package model

import (
	"time"

	"github.com/google/uuid"
)

// Session 会话模型 - 对应 sessions 表
//
// 用途: 存储 JWT Refresh Token，实现 token 轮换和会话管理
//
// 工作原理:
//   1. 用户登录成功后，创建一个 Session 记录
//   2. Session.ID 作为 Refresh Token 的 payload
//   3. 刷新 token 时，验证 Session 是否存在且未被封禁
//   4. 用户登出时，删除或封禁对应的 Session
//
// 安全特性:
//   - IsBlocked: 可以手动封禁某个会话(如检测到异常登录)
//   - UserAgent/ClientIP: 用于审计和异常检测
//   - ExpiresAt: 自动过期，需要定期清理过期记录
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"not null;index;size:255" json:"username"`       // 关联的用户名
	RefreshToken string    `gorm:"not null;size:512" json:"-"`                    // Refresh Token (不输出到JSON)
	UserAgent    string    `gorm:"not null;size:255;default:''" json:"user_agent"` // 客户端标识
	ClientIP     string    `gorm:"not null;size:45;default:''" json:"client_ip"`   // 客户端IP
	IsBlocked    bool      `gorm:"not null;default:false" json:"is_blocked"`       // 是否被封禁
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`                     // 过期时间
	CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	User User `gorm:"foreignKey:Username;references:Username" json:"-"`
}

// TableName 指定表名
func (Session) TableName() string {
	return "sessions"
}

// IsExpired 检查会话是否已过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsValid 检查会话是否有效 (未过期且未封禁)
func (s *Session) IsValid() bool {
	return !s.IsExpired() && !s.IsBlocked
}
