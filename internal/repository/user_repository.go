package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// UserRepository 用户数据访问实现
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建 UserRepository 实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 创建新用户
// 如果用户名或邮箱已存在，返回相应错误
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		// 检查是否是唯一约束冲突
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return apperrors.New(apperrors.CodeUsernameExists)
		}
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// GetByUsername 根据用户名查询用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound()
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &user, nil
}

// GetByEmail 根据邮箱查询用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound()
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &user, nil
}

// GetByID 根据ID查询用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrUserNotFound()
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}
