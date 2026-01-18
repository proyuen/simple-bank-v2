package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// TransferRepository 转账数据访问实现
type TransferRepository struct {
	db *gorm.DB
}

// NewTransferRepository 创建 TransferRepository 实例
func NewTransferRepository(db *gorm.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

// Create 创建转账记录
func (r *TransferRepository) Create(ctx context.Context, transfer *model.Transfer) error {
	result := r.db.WithContext(ctx).Create(transfer)
	if result.Error != nil {
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// GetByID 根据ID查询转账
func (r *TransferRepository) GetByID(ctx context.Context, id uint) (*model.Transfer, error) {
	var transfer model.Transfer
	result := r.db.WithContext(ctx).First(&transfer, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("transfer")
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &transfer, nil
}

// ListByAccountID 获取与账户相关的所有转账
func (r *TransferRepository) ListByAccountID(ctx context.Context, accountID uint, limit, offset int) ([]model.Transfer, int64, error) {
	var transfers []model.Transfer
	var total int64

	condition := "from_account_id = ? OR to_account_id = ?"

	if err := r.db.WithContext(ctx).
		Model(&model.Transfer{}).
		Where(condition, accountID, accountID).
		Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := r.db.WithContext(ctx).
		Where(condition, accountID, accountID).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&transfers).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return transfers, total, nil
}
