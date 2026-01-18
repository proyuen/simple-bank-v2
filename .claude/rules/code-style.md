# Go Code Style Guidelines

## Naming Conventions

### Packages
```go
// GOOD: lowercase, single word
package repository
package service
package handler

// BAD: underscores, mixed case
package user_repository  // wrong
package UserService      // wrong
```

### Files
```go
// Pattern: lowercase_with_underscores.go
user_repository.go
account_service.go
auth_middleware.go
transfer_handler.go

// Test files: add _test suffix
user_repository_test.go
account_service_test.go
```

### Variables
```go
// Local variables: camelCase
userID := 123
accessToken := "xxx"
pageSize := 10

// Package-level exported: PascalCase
var DefaultPageSize = 10
var MaxRetryCount = 3

// Package-level unexported: camelCase
var defaultTimeout = 30 * time.Second
```

### Constants
```go
// Grouped constants with iota
const (
    StatusPending = iota
    StatusActive
    StatusBlocked
)

// String constants
const (
    CurrencyUSD = "USD"
    CurrencyEUR = "EUR"
    CurrencyCNY = "CNY"
)

// Error constants: Err prefix
var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidPassword = errors.New("invalid password")
    ErrInvalidToken    = errors.New("invalid token")
)
```

### Functions and Methods
```go
// Exported: PascalCase
func CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
func (s *UserService) GetByID(ctx context.Context, id int64) (*User, error)

// Unexported: camelCase
func hashPassword(password string) (string, error)
func (s *UserService) validateEmail(email string) error
```

### Interfaces
```go
// Pattern: PascalCase, typically verb + "er"
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
}

type TokenMaker interface {
    CreateToken(username string, duration time.Duration) (string, error)
    VerifyToken(token string) (*Payload, error)
}

// Single method interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

### Structs
```go
// PascalCase for exported structs
type User struct {
    ID        int64     `gorm:"primaryKey"`
    Username  string    `gorm:"uniqueIndex"`
    Email     string    `gorm:"uniqueIndex"`
    CreatedAt time.Time
}

// Group related fields with blank lines
type TransferRequest struct {
    // Source account
    FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`

    // Destination account
    ToAccountID int64 `json:"to_account_id" binding:"required,min=1"`

    // Transfer details
    Amount   int64  `json:"amount" binding:"required,gt=0"`
    Currency string `json:"currency" binding:"required,oneof=USD EUR CNY"`
}
```

## Struct Tags

### JSON Tags
```go
type UserResponse struct {
    ID        int64     `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// Omit empty values
type UpdateUserRequest struct {
    FullName *string `json:"full_name,omitempty"`
    Email    *string `json:"email,omitempty"`
}

// Hide sensitive fields
type User struct {
    HashedPassword string `json:"-"` // never serialize
}
```

### GORM Tags
```go
type Account struct {
    ID        int64  `gorm:"primaryKey;autoIncrement"`
    Owner     string `gorm:"type:varchar(255);not null;index"`
    Balance   int64  `gorm:"not null;default:0"`
    Currency  string `gorm:"type:varchar(3);not null"`
    CreatedAt time.Time
    UpdatedAt time.Time

    // Foreign key
    User User `gorm:"foreignKey:Owner;references:Username"`
}

// Unique constraint
type User struct {
    Username string `gorm:"uniqueIndex"`
    Email    string `gorm:"uniqueIndex"`
}

// Composite unique index
type Account struct {
    Owner    string `gorm:"uniqueIndex:idx_owner_currency"`
    Currency string `gorm:"uniqueIndex:idx_owner_currency"`
}
```

### Gin Binding Tags
```go
type CreateUserRequest struct {
    // Required field
    Username string `json:"username" binding:"required"`

    // Required with length validation
    Password string `json:"password" binding:"required,min=6,max=100"`

    // Email validation
    Email string `json:"email" binding:"required,email"`

    // Enum validation
    Currency string `json:"currency" binding:"required,oneof=USD EUR CNY"`

    // Numeric range
    Amount int64 `json:"amount" binding:"required,gt=0,lte=1000000"`
}

// Query parameters use `form` tag
type ListAccountsRequest struct {
    PageID   int32 `form:"page_id" binding:"required,min=1"`
    PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// URI parameters use `uri` tag
type GetAccountRequest struct {
    ID int64 `uri:"id" binding:"required,min=1"`
}
```

## Function Signatures

### Parameter Order
```go
// 1. context first
// 2. main input parameters
// 3. optional/config parameters
// 4. return: result, then error
func CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
func GetAccount(ctx context.Context, id int64) (*Account, error)
func ListAccounts(ctx context.Context, owner string, limit, offset int32) ([]Account, error)
```

### Return Values
```go
// Single value + error
func GetByID(ctx context.Context, id int64) (*User, error)

// Multiple values + error
func CreateTransfer(ctx context.Context, arg CreateTransferParams) (*Transfer, *Entry, *Entry, error)

// Use named struct for multiple returns
type TransferResult struct {
    Transfer    *Transfer
    FromAccount *Account
    ToAccount   *Account
    FromEntry   *Entry
    ToEntry     *Entry
}
func CreateTransfer(ctx context.Context, arg CreateTransferParams) (*TransferResult, error)
```

## Comments

### Package Comments
```go
// Package repository provides data access layer implementations
// for the Simple Bank application. It contains interfaces and
// implementations for all database operations.
package repository
```

### Function Comments
```go
// CreateUser creates a new user in the database.
// It hashes the password before storing and returns the created user.
//
// Returns ErrDuplicateUsername if username already exists.
// Returns ErrDuplicateEmail if email already exists.
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
```

### Inline Comments
```go
// Use sparingly, only when logic is not obvious
func (s *TransferService) CreateTransfer(ctx context.Context, arg CreateTransferParams) (*TransferResult, error) {
    // Lock accounts in consistent order to prevent deadlock
    if arg.FromAccountID < arg.ToAccountID {
        // Lock from_account first
    } else {
        // Lock to_account first
    }
}
```

## Import Organization

```go
import (
    // Standard library (group 1)
    "context"
    "errors"
    "time"

    // Third-party packages (group 2)
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    // Internal packages (group 3)
    "github.com/proyuen/simple-bank-v2/internal/dto"
    "github.com/proyuen/simple-bank-v2/internal/model"
    "github.com/proyuen/simple-bank-v2/internal/repository"
)
```

## Code Formatting

- Always run `gofmt` or `goimports` before committing
- Use tabs for indentation (Go standard)
- Maximum line length: 120 characters (soft limit)
- No trailing whitespace
- Single blank line between functions
- No blank line at start/end of function body
