package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// AccountRepository 账户数据访问实现
type AccountRepository struct {
	db *gorm.DB
}

// NewAccountRepository 创建 AccountRepository 实例
func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create 创建新账户
func (r *AccountRepository) Create(ctx context.Context, account *model.Account) error {
	result := r.db.WithContext(ctx).Create(account)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return apperrors.NewWithMessage(apperrors.CodeAlreadyExists, "account with this currency already exists")
		}
		return apperrors.ErrDatabase(result.Error)
	}
	return nil
}

// GetByID 根据ID查询账户
func (r *AccountRepository) GetByID(ctx context.Context, id uint) (*model.Account, error) {
	var account model.Account
	result := r.db.WithContext(ctx).First(&account, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAccountNotFound()
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &account, nil
}

// GetByOwnerAndCurrency 根据所有者和货币类型查询账户
func (r *AccountRepository) GetByOwnerAndCurrency(ctx context.Context, owner, currency string) (*model.Account, error) {
	var account model.Account
	result := r.db.WithContext(ctx).
		Where("owner = ? AND currency = ?", owner, currency).
		First(&account)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrAccountNotFound()
		}
		return nil, apperrors.ErrDatabase(result.Error)
	}
	return &account, nil
}

// ListByOwner 获取用户的所有账户 (带分页)
func (r *AccountRepository) ListByOwner(ctx context.Context, owner string, limit, offset int) ([]model.Account, int64, error) {
	var accounts []model.Account
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("owner = ?", owner).
		Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := r.db.WithContext(ctx).
		Where("owner = ?", owner).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&accounts).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return accounts, total, nil
}

// GetForUpdate 获取账户并锁定 (FOR UPDATE)
func (r *AccountRepository) GetForUpdate(ctx context.Context, id uint) (*model.Account, error) {
	var account model.Account
	result := r.db.WithContext(ctx).
		Raw("SELECT * FROM accounts WHERE id = ? FOR UPDATE", id).
		Scan(&account)
	if result.Error != nil {
		return nil, apperrors.ErrDatabase(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.ErrAccountNotFound()
	}
	return &account, nil
}

// UpdateBalance 更新账户余额
func (r *AccountRepository) UpdateBalance(ctx context.Context, id uint, amount int64) (*model.Account, error) {
	var account model.Account

	result := r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount))

	if result.Error != nil {
		return nil, apperrors.ErrDatabase(result.Error)
	}

	if err := r.db.WithContext(ctx).First(&account, id).Error; err != nil {
		return nil, apperrors.ErrDatabase(err)
	}

	return &account, nil
}
