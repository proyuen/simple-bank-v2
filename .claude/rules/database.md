# Database Guidelines

## GORM Model Definition

### Basic Model Structure
```go
// internal/model/user.go
package model

import "time"

type User struct {
    ID                int64     `gorm:"primaryKey;autoIncrement"`
    Username          string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    HashedPassword    string    `gorm:"type:varchar(255);not null"`
    FullName          string    `gorm:"type:varchar(255);not null"`
    Email             string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    PasswordChangedAt time.Time `gorm:"not null;default:'1970-01-01 00:00:01'"`
    CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName overrides default table name
func (User) TableName() string {
    return "users"
}
```

### Model with Foreign Key
```go
// internal/model/account.go
type Account struct {
    ID        int64     `gorm:"primaryKey;autoIncrement"`
    Owner     string    `gorm:"type:varchar(255);not null;index:idx_owner"`
    Balance   int64     `gorm:"not null;default:0"`
    Currency  string    `gorm:"type:varchar(3);not null"`
    CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
    UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

    // Foreign key relationship
    User User `gorm:"foreignKey:Owner;references:Username;constraint:OnDelete:CASCADE"`
}

// Composite unique index
func (Account) TableName() string {
    return "accounts"
}
```

### Model with UUID Primary Key
```go
// internal/model/session.go
type Session struct {
    ID           string    `gorm:"type:char(36);primaryKey"` // UUID stored as string
    Username     string    `gorm:"type:varchar(255);not null;index:idx_username"`
    RefreshToken string    `gorm:"type:varchar(512);not null"`
    UserAgent    string    `gorm:"type:varchar(255);not null;default:''"`
    ClientIP     string    `gorm:"type:varchar(45);not null;default:''"` // IPv6 support
    IsBlocked    bool      `gorm:"not null;default:false"`
    ExpiresAt    time.Time `gorm:"not null"`
    CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`

    // Foreign key
    User User `gorm:"foreignKey:Username;references:Username;constraint:OnDelete:CASCADE"`
}
```

## Repository Pattern

### Interface Definition
```go
// internal/repository/user_repository.go
package repository

import (
    "context"
    "github.com/proyuen/simple-bank-v2/internal/model"
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id int64) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
}
```

### Implementation
```go
// internal/repository/user_repository_impl.go
package repository

import (
    "context"
    "errors"
    "fmt"
    "strings"

    "gorm.io/gorm"

    apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
    "github.com/proyuen/simple-bank-v2/internal/model"
)

type userRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    err := r.db.WithContext(ctx).Create(user).Error
    if err != nil {
        if strings.Contains(err.Error(), "Duplicate entry") {
            if strings.Contains(err.Error(), "username") {
                return apperrors.ErrDuplicateUsername
            }
            if strings.Contains(err.Error(), "email") {
                return apperrors.ErrDuplicateEmail
            }
        }
        return fmt.Errorf("failed to create user: %w", err)
    }
    return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
    err := r.db.WithContext(ctx).Save(user).Error
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    return nil
}
```

## Transaction Handling

### Transaction Executor Pattern
```go
// internal/repository/tx.go
package repository

import (
    "context"
    "fmt"

    "gorm.io/gorm"
)

type TxExecutor struct {
    db *gorm.DB
}

func NewTxExecutor(db *gorm.DB) *TxExecutor {
    return &TxExecutor{db: db}
}

// ExecTx executes a function within a database transaction
func (t *TxExecutor) ExecTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
    tx := t.db.WithContext(ctx).Begin()
    if tx.Error != nil {
        return fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }

    err := fn(tx)
    if err != nil {
        if rbErr := tx.Rollback().Error; rbErr != nil {
            return fmt.Errorf("tx error: %v, rollback error: %w", err, rbErr)
        }
        return err
    }

    if err := tx.Commit().Error; err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

### Using Transaction in Service
```go
// internal/service/transfer_service.go
func (s *TransferService) CreateTransfer(ctx context.Context, arg *dto.CreateTransferRequest) (*dto.TransferResult, error) {
    var result dto.TransferResult

    err := s.txExecutor.ExecTx(ctx, func(tx *gorm.DB) error {
        // Create transfer record
        transfer := &model.Transfer{
            FromAccountID: arg.FromAccountID,
            ToAccountID:   arg.ToAccountID,
            Amount:        arg.Amount,
        }
        if err := tx.Create(transfer).Error; err != nil {
            return err
        }
        result.Transfer = transfer

        // Update account balances with row locking
        // Lock accounts in consistent order to prevent deadlock
        account1ID, account2ID := arg.FromAccountID, arg.ToAccountID
        if account1ID > account2ID {
            account1ID, account2ID = account2ID, account1ID
        }

        // Lock first account
        var account1 model.Account
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&account1, account1ID).Error; err != nil {
            return err
        }

        // Lock second account
        var account2 model.Account
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            First(&account2, account2ID).Error; err != nil {
            return err
        }

        // Update balances
        if err := tx.Model(&model.Account{}).
            Where("id = ?", arg.FromAccountID).
            Update("balance", gorm.Expr("balance - ?", arg.Amount)).Error; err != nil {
            return err
        }

        if err := tx.Model(&model.Account{}).
            Where("id = ?", arg.ToAccountID).
            Update("balance", gorm.Expr("balance + ?", arg.Amount)).Error; err != nil {
            return err
        }

        // Create entry records
        fromEntry := &model.Entry{AccountID: arg.FromAccountID, Amount: -arg.Amount}
        if err := tx.Create(fromEntry).Error; err != nil {
            return err
        }
        result.FromEntry = fromEntry

        toEntry := &model.Entry{AccountID: arg.ToAccountID, Amount: arg.Amount}
        if err := tx.Create(toEntry).Error; err != nil {
            return err
        }
        result.ToEntry = toEntry

        return nil
    })

    if err != nil {
        return nil, err
    }

    return &result, nil
}
```

## Query Patterns

### Pagination
```go
func (r *accountRepository) List(ctx context.Context, owner string, limit, offset int) ([]model.Account, int64, error) {
    var accounts []model.Account
    var total int64

    query := r.db.WithContext(ctx).Model(&model.Account{}).Where("owner = ?", owner)

    // Count total
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Fetch with pagination
    if err := query.Order("id DESC").Limit(limit).Offset(offset).Find(&accounts).Error; err != nil {
        return nil, 0, err
    }

    return accounts, total, nil
}
```

### Select Specific Columns
```go
func (r *userRepository) GetBasicInfo(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).
        Select("id", "username", "email", "created_at").
        First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### Preload Relationships
```go
func (r *accountRepository) GetWithUser(ctx context.Context, id int64) (*model.Account, error) {
    var account model.Account
    err := r.db.WithContext(ctx).
        Preload("User").
        First(&account, id).Error
    if err != nil {
        return nil, err
    }
    return &account, nil
}
```

### Row Locking for Updates
```go
import "gorm.io/gorm/clause"

func (r *accountRepository) GetForUpdate(ctx context.Context, id int64) (*model.Account, error) {
    var account model.Account
    err := r.db.WithContext(ctx).
        Clauses(clause.Locking{Strength: "UPDATE"}).
        First(&account, id).Error
    if err != nil {
        return nil, err
    }
    return &account, nil
}
```

### Update Specific Fields
```go
func (r *accountRepository) UpdateBalance(ctx context.Context, id int64, amount int64) error {
    return r.db.WithContext(ctx).
        Model(&model.Account{}).
        Where("id = ?", id).
        Update("balance", gorm.Expr("balance + ?", amount)).Error
}
```

## Database Connection

### Connection Setup
```go
// cmd/server/main.go
package main

import (
    "log"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/proyuen/simple-bank-v2/internal/config"
)

func main() {
    cfg, err := config.LoadConfig(".")
    if err != nil {
        log.Fatal("cannot load config:", err)
    }

    db, err := gorm.Open(mysql.Open(cfg.DBSource()), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatal("cannot connect to database:", err)
    }

    // Get underlying SQL DB for connection pool settings
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatal("cannot get database connection:", err)
    }

    // Connection pool settings
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    // ... continue with app setup
}
```

## MySQL-Specific Notes

### UUID Generation
```go
import "github.com/google/uuid"

// Generate UUID before insert (MySQL doesn't have native UUID type)
session := &model.Session{
    ID:           uuid.New().String(),
    Username:     username,
    RefreshToken: refreshToken,
    // ...
}
```

### Timestamp Handling
```go
// MySQL TIMESTAMP range: '1970-01-01 00:00:01' to '2038-01-19 03:14:07'
// For password_changed_at default, use '1970-01-01 00:00:01' not zero time
```

### Character Set
```go
// DSN should include charset for proper Unicode support
dsn := "user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=true&loc=Local"
```
