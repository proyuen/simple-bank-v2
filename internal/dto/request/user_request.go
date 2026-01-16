// Package request 定义了 API 请求的数据结构
// 使用 validator 标签进行请求参数验证
package request

// CreateUserRequest 创建用户请求
// 用于: POST /api/v1/users
type CreateUserRequest struct {
	// Username 用户名
	// 规则: 必填, 3-50字符, 只允许字母数字下划线
	Username string `json:"username" binding:"required,min=3,max=50,alphanum"`

	// Password 密码
	// 规则: 必填, 至少6字符
	Password string `json:"password" binding:"required,min=6"`

	// FullName 真实姓名
	// 规则: 必填
	FullName string `json:"full_name" binding:"required"`

	// Email 邮箱
	// 规则: 必填, 有效的邮箱格式
	Email string `json:"email" binding:"required,email"`
}

// LoginUserRequest 用户登录请求
// 用于: POST /api/v1/users/login
type LoginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新 Token 请求
// 用于: POST /api/v1/token/refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
