package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// SessionRepository 会话数据访问实现
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository 创建 SessionRepository 实例
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 创建会话
func (r *SessionRepository) Create(ctx context.Context, session *model.Session) error {
	result := r.db.WithContext(ctx).Create(session)
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// GetByID 根据ID查询会话
// id: UUID 字符串格式
func (r *SessionRepository) GetByID(ctx context.Context, id string) (*model.Session, error) {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.NewWithMessage(apperrors.CodeInvalidRequest, "invalid session id")
	}

	var session model.Session
	result := r.db.WithContext(ctx).Where("id = ?", sessionID).First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("session")
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &session, nil
}

// DeleteByUsername 删除用户的所有会话
// 用于"登出所有设备"功能
func (r *SessionRepository) DeleteByUsername(ctx context.Context, username string) error {
	result := r.db.WithContext(ctx).
		Where("username = ?", username).
		Delete(&model.Session{})
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// Block 封禁会话
// 用于检测到异常登录时手动封禁
func (r *SessionRepository) Block(ctx context.Context, id string) error {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.NewWithMessage(apperrors.CodeInvalidRequest, "invalid session id")
	}

	result := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Where("id = ?", sessionID).
		Update("is_blocked", true)
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound("session")
	}
	return nil
}
