package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// 定义 Token 相关错误
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

// Payload 包含 JWT Token 的载荷数据
type Payload struct {
	ID        uuid.UUID `json:"id"`         // Token 唯一标识
	Username  string    `json:"username"`   // 用户名
	IssuedAt  time.Time `json:"issued_at"`  // 签发时间
	ExpiredAt time.Time `json:"expired_at"` // 过期时间
}

// NewPayload 创建一个新的 Token 载荷
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  now,
		ExpiredAt: now.Add(duration),
	}

	return payload, nil
}

// Valid 检查 Token 载荷是否有效
// 实现 jwt.Claims 接口
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
