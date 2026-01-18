package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/proyuen/simple-bank-v2/internal/dto/request"
	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
	"github.com/proyuen/simple-bank-v2/pkg/password"
	"github.com/proyuen/simple-bank-v2/pkg/token"
)

// ==================== 接口定义 (由使用方定义) ====================

// UserRepository 用户数据访问接口
// 定义了 UserService 需要的数据访问能力
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

// SessionRepository 会话数据访问接口
// 定义了 UserService 需要的会话管理能力
type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	GetByID(ctx context.Context, id string) (*model.Session, error)
	DeleteByUsername(ctx context.Context, username string) error
	Block(ctx context.Context, id string) error
}

// ==================== Service 实现 ====================

// UserService 用户业务逻辑
type UserService struct {
	userRepo       UserRepository
	sessionRepo    SessionRepository
	tokenMaker     token.Maker
	accessDuration time.Duration
	refreshDuration time.Duration
}

// NewUserService 创建 UserService 实例
func NewUserService(
	userRepo UserRepository,
	sessionRepo SessionRepository,
	tokenMaker token.Maker,
	accessDuration, refreshDuration time.Duration,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		sessionRepo:     sessionRepo,
		tokenMaker:      tokenMaker,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
	}
}

// CreateUser 创建新用户
func (s *UserService) CreateUser(ctx context.Context, req *request.CreateUserRequest) (*response.UserResponse, error) {
	// 1. 密码加密
	hashedPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.ErrInternalServer()
	}

	// 2. 创建用户模型
	user := &model.User{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	// 3. 保存到数据库
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 4. 返回响应
	return s.toUserResponse(user), nil
}

// LoginUser 用户登录
func (s *UserService) LoginUser(ctx context.Context, req *request.LoginUserRequest, userAgent, clientIP string) (*response.LoginResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperrors.ErrPasswordWrong() // 不暴露用户是否存在
	}

	// 2. 验证密码
	if err := password.CheckPassword(req.Password, user.HashedPassword); err != nil {
		return nil, apperrors.ErrPasswordWrong()
	}

	// 3. 生成 Access Token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, s.accessDuration)
	if err != nil {
		return nil, apperrors.ErrInternalServer()
	}

	// 4. 生成 Refresh Token
	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, s.refreshDuration)
	if err != nil {
		return nil, apperrors.ErrInternalServer()
	}

	// 5. 保存会话
	session := &model.Session{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		ClientIP:     clientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	// 6. 返回响应
	return &response.LoginResponse{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		SessionID:             session.ID.String(),
		User:                  *s.toUserResponse(user),
	}, nil
}

// RefreshToken 刷新 Access Token
func (s *UserService) RefreshToken(ctx context.Context, req *request.RefreshTokenRequest) (*response.RefreshTokenResponse, error) {
	// 1. 验证 Refresh Token
	payload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidToken)
	}

	// 2. 查找会话
	session, err := s.sessionRepo.GetByID(ctx, payload.ID.String())
	if err != nil {
		return nil, err
	}

	// 3. 验证会话
	if session.IsBlocked {
		return nil, apperrors.New(apperrors.CodeAccountBlocked)
	}
	if session.Username != payload.Username {
		return nil, apperrors.New(apperrors.CodeUnauthorized)
	}
	if session.RefreshToken != req.RefreshToken {
		return nil, apperrors.New(apperrors.CodeInvalidToken)
	}

	// 4. 生成新的 Access Token
	accessToken, accessPayload, err := s.tokenMaker.CreateToken(payload.Username, s.accessDuration)
	if err != nil {
		return nil, apperrors.ErrInternalServer()
	}

	return &response.RefreshTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return s.toUserResponse(user), nil
}

// toUserResponse 转换为用户响应
func (s *UserService) toUserResponse(user *model.User) *response.UserResponse {
	return &response.UserResponse{
		ID:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// 确保 uuid 包被使用
var _ = uuid.New
