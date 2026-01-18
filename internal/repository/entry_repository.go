package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// EntryRepository 账目数据访问实现
type EntryRepository struct {
	db *gorm.DB
}

// NewEntryRepository 创建 EntryRepository 实例
func NewEntryRepository(db *gorm.DB) *EntryRepository {
	return &EntryRepository{db: db}
}

// Create 创建账目记录
func (r *EntryRepository) Create(ctx context.Context, entry *model.Entry) error {
	result := r.db.WithContext(ctx).Create(entry)
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// GetByID 根据ID查询账目
func (r *EntryRepository) GetByID(ctx context.Context, id uint) (*model.Entry, error) {
	var entry model.Entry
	result := r.db.WithContext(ctx).First(&entry, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("entry")
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &entry, nil
}

// ListByAccountID 获取账户的所有账目 (带分页)
func (r *EntryRepository) ListByAccountID(ctx context.Context, accountID uint, limit, offset int) ([]model.Entry, int64, error) {
	var entries []model.Entry
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&model.Entry{}).
		Where("account_id = ?", accountID).
		Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return entries, total, nil
}
