package service

import (
	"context"

	"github.com/proyuen/simple-bank-v2/internal/dto/request"
	"github.com/proyuen/simple-bank-v2/internal/dto/response"
	apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
	"github.com/proyuen/simple-bank-v2/internal/model"
)

// ==================== 接口定义 (由使用方定义) ====================

// AccountRepository 账户数据访问接口
// 定义了 AccountService 需要的数据访问能力
type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id uint) (*model.Account, error)
	GetByOwnerAndCurrency(ctx context.Context, owner, currency string) (*model.Account, error)
	ListByOwner(ctx context.Context, owner string, limit, offset int) ([]model.Account, int64, error)
}

// ==================== Service 实现 ====================

// AccountService 账户业务逻辑
type AccountService struct {
	accountRepo AccountRepository
}

// NewAccountService 创建 AccountService 实例
func NewAccountService(accountRepo AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

// CreateAccount 创建新账户
func (s *AccountService) CreateAccount(ctx context.Context, owner string, req *request.CreateAccountRequest) (*response.AccountResponse, error) {
	// 1. 检查是否已存在相同货币类型的账户
	existingAccount, err := s.accountRepo.GetByOwnerAndCurrency(ctx, owner, req.Currency)
	if err == nil && existingAccount != nil {
		return nil, apperrors.NewWithMessage(apperrors.CodeAlreadyExists, "account with this currency already exists")
	}

	// 2. 创建账户模型
	account := &model.Account{
		Owner:    owner,
		Balance:  0,
		Currency: req.Currency,
	}

	// 3. 保存到数据库
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	// 4. 返回响应
	return s.toAccountResponse(account), nil
}

// GetAccount 获取账户详情
func (s *AccountService) GetAccount(ctx context.Context, owner string, accountID uint) (*response.AccountResponse, error) {
	// 1. 查询账户
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// 2. 验证账户所有权
	if account.Owner != owner {
		return nil, apperrors.ErrUnauthorized()
	}

	// 3. 返回响应
	return s.toAccountResponse(account), nil
}

// ListAccounts 获取用户的账户列表
func (s *AccountService) ListAccounts(ctx context.Context, owner string, req *request.PaginationRequest) (*response.ListResponse[response.AccountResponse], error) {
	// 1. 计算分页参数
	limit := req.Limit()
	offset := req.Offset()

	// 2. 查询账户列表
	accounts, total, err := s.accountRepo.ListByOwner(ctx, owner, limit, offset)
	if err != nil {
		return nil, err
	}

	// 3. 转换为响应格式
	items := make([]response.AccountResponse, len(accounts))
	for i, account := range accounts {
		items[i] = *s.toAccountResponse(&account)
	}

	// 4. 返回分页响应
	result := response.NewListResponse(items, req.PageID, req.PageSize, total)
	return &result, nil
}

// toAccountResponse 转换为账户响应
func (s *AccountService) toAccountResponse(account *model.Account) *response.AccountResponse {
	return &response.AccountResponse{
		ID:        account.ID,
		Owner:     account.Owner,
		Balance:   account.Balance,
		Currency:  account.Currency,
		CreatedAt: account.CreatedAt,
	}
}
