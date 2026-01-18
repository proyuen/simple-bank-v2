package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/proyuen/simple-bank-v2/internal/dto/request"
	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// ==================== 接口定义 (由使用方定义) ====================

// TransferAccountRepository 转账服务需要的账户数据访问接口
type TransferAccountRepository interface {
	GetByID(ctx context.Context, id uint) (*model.Account, error)
	GetForUpdate(ctx context.Context, id uint) (*model.Account, error)
	UpdateBalance(ctx context.Context, id uint, amount int64) (*model.Account, error)
}

// TransferRepository 转账数据访问接口
type TransferRepository interface {
	Create(ctx context.Context, transfer *model.Transfer) error
	GetByID(ctx context.Context, id uint) (*model.Transfer, error)
	ListByAccountID(ctx context.Context, accountID uint, limit, offset int) ([]model.Transfer, int64, error)
}

// EntryRepository 账目数据访问接口
type EntryRepository interface {
	Create(ctx context.Context, entry *model.Entry) error
	GetByID(ctx context.Context, id uint) (*model.Entry, error)
	ListByAccountID(ctx context.Context, accountID uint, limit, offset int) ([]model.Entry, int64, error)
}

// TransactionManager 事务管理接口
type TransactionManager interface {
	Transaction(fc func(tx *gorm.DB) error) error
}

// ==================== Service 实现 ====================

// TransferService 转账业务逻辑
type TransferService struct {
	db           TransactionManager
	accountRepo  TransferAccountRepository
	transferRepo TransferRepository
	entryRepo    EntryRepository
}

// NewTransferService 创建 TransferService 实例
func NewTransferService(
	db TransactionManager,
	accountRepo TransferAccountRepository,
	transferRepo TransferRepository,
	entryRepo EntryRepository,
) *TransferService {
	return &TransferService{
		db:           db,
		accountRepo:  accountRepo,
		transferRepo: transferRepo,
		entryRepo:    entryRepo,
	}
}

// TransferResult 转账结果
type TransferResult struct {
	Transfer    *model.Transfer
	FromAccount *model.Account
	ToAccount   *model.Account
	FromEntry   *model.Entry
	ToEntry     *model.Entry
}

// CreateTransfer 创建转账
func (s *TransferService) CreateTransfer(ctx context.Context, owner string, req *request.CreateTransferRequest) (*response.TransferResponse, error) {
	// 1. 验证源账户存在且属于当前用户
	fromAccount, err := s.accountRepo.GetByID(ctx, req.FromAccountID)
	if err != nil {
		return nil, err
	}
	if fromAccount.Owner != owner {
		return nil, apperrors.ErrUnauthorized()
	}

	// 2. 验证目标账户存在
	toAccount, err := s.accountRepo.GetByID(ctx, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	// 3. 验证货币类型一致
	if fromAccount.Currency != toAccount.Currency {
		return nil, apperrors.NewWithMessage(apperrors.CodeInvalidRequest, "currency mismatch")
	}

	// 4. 验证余额充足
	if fromAccount.Balance < req.Amount {
		return nil, apperrors.NewWithMessage(apperrors.CodeInsufficientBalance, "insufficient balance")
	}

	// 5. 执行转账事务
	var result TransferResult
	err = s.db.Transaction(func(tx *gorm.DB) error {
		return s.execTransfer(ctx, req.FromAccountID, req.ToAccountID, req.Amount, &result)
	})
	if err != nil {
		return nil, err
	}

	// 6. 返回响应
	return s.toTransferResponse(result.Transfer), nil
}

// execTransfer 执行转账事务
func (s *TransferService) execTransfer(ctx context.Context, fromAccountID, toAccountID uint, amount int64, result *TransferResult) error {
	var err error

	// 1. 创建转账记录
	result.Transfer = &model.Transfer{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
	}
	if err = s.transferRepo.Create(ctx, result.Transfer); err != nil {
		return err
	}

	// 2. 创建源账户账目 (负数表示支出)
	result.FromEntry = &model.Entry{
		AccountID: fromAccountID,
		Amount:    -amount,
	}
	if err = s.entryRepo.Create(ctx, result.FromEntry); err != nil {
		return err
	}

	// 3. 创建目标账户账目 (正数表示收入)
	result.ToEntry = &model.Entry{
		AccountID: toAccountID,
		Amount:    amount,
	}
	if err = s.entryRepo.Create(ctx, result.ToEntry); err != nil {
		return err
	}

	// 4. 更新账户余额 (按 ID 顺序更新以避免死锁)
	if fromAccountID < toAccountID {
		result.FromAccount, result.ToAccount, err = s.addMoney(ctx, fromAccountID, -amount, toAccountID, amount)
	} else {
		result.ToAccount, result.FromAccount, err = s.addMoney(ctx, toAccountID, amount, fromAccountID, -amount)
	}

	return err
}

// addMoney 按顺序更新两个账户余额
func (s *TransferService) addMoney(ctx context.Context, accountID1 uint, amount1 int64, accountID2 uint, amount2 int64) (*model.Account, *model.Account, error) {
	account1, err := s.accountRepo.UpdateBalance(ctx, accountID1, amount1)
	if err != nil {
		return nil, nil, err
	}

	account2, err := s.accountRepo.UpdateBalance(ctx, accountID2, amount2)
	if err != nil {
		return nil, nil, err
	}

	return account1, account2, nil
}

// ListTransfers 获取账户的转账记录
func (s *TransferService) ListTransfers(ctx context.Context, owner string, accountID uint, req *request.PaginationRequest) (*response.ListResponse[response.TransferResponse], error) {
	// 1. 验证账户属于当前用户
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Owner != owner {
		return nil, apperrors.ErrUnauthorized()
	}

	// 2. 计算分页参数
	limit := req.Limit()
	offset := req.Offset()

	// 3. 查询转账记录
	transfers, total, err := s.transferRepo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 4. 转换为响应格式
	items := make([]response.TransferResponse, len(transfers))
	for i, transfer := range transfers {
		items[i] = *s.toTransferResponse(&transfer)
	}

	// 5. 返回分页响应
	result := response.NewListResponse(items, req.PageID, req.PageSize, total)
	return &result, nil
}

// ListEntries 获取账户的账目记录
func (s *TransferService) ListEntries(ctx context.Context, owner string, accountID uint, req *request.PaginationRequest) (*response.ListResponse[response.EntryResponse], error) {
	// 1. 验证账户属于当前用户
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account.Owner != owner {
		return nil, apperrors.ErrUnauthorized()
	}

	// 2. 计算分页参数
	limit := req.Limit()
	offset := req.Offset()

	// 3. 查询账目记录
	entries, total, err := s.entryRepo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	// 4. 转换为响应格式
	items := make([]response.EntryResponse, len(entries))
	for i, entry := range entries {
		items[i] = *s.toEntryResponse(&entry)
	}

	// 5. 返回分页响应
	result := response.NewListResponse(items, req.PageID, req.PageSize, total)
	return &result, nil
}

// toTransferResponse 转换为转账响应
func (s *TransferService) toTransferResponse(transfer *model.Transfer) *response.TransferResponse {
	return &response.TransferResponse{
		ID:            transfer.ID,
		FromAccountID: transfer.FromAccountID,
		ToAccountID:   transfer.ToAccountID,
		Amount:        transfer.Amount,
		CreatedAt:     transfer.CreatedAt,
	}
}

// toEntryResponse 转换为账目响应
func (s *TransferService) toEntryResponse(entry *model.Entry) *response.EntryResponse {
	return &response.EntryResponse{
		ID:        entry.ID,
		AccountID: entry.AccountID,
		Amount:    entry.Amount,
		CreatedAt: entry.CreatedAt,
	}
}
