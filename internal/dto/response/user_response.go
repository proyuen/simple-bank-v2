// Package response 定义了 API 响应的数据结构
package response

import "time"

// UserResponse 用户信息响应
// 注意: 不包含密码等敏感信息
type UserResponse struct {
	ID                uint      `json:"id"`
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken           string       `json:"access_token"`            // Access Token
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"` // Access Token 过期时间
	RefreshToken          string       `json:"refresh_token"`           // Refresh Token
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"` // Refresh Token 过期时间
	SessionID             string       `json:"session_id"`              // 会话ID
	User                  UserResponse `json:"user"`                    // 用户信息
}

// RefreshTokenResponse 刷新 Token 响应
type RefreshTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}
