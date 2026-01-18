# Error Handling Guidelines

## Custom Error Types

### Define in internal/errors/errors.go
```go
package errors

import "errors"

// User errors
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrDuplicateUsername = errors.New("username already exists")
    ErrDuplicateEmail    = errors.New("email already exists")
    ErrInvalidPassword   = errors.New("invalid password")
)

// Account errors
var (
    ErrAccountNotFound     = errors.New("account not found")
    ErrAccountNotBelongTo  = errors.New("account does not belong to user")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrCurrencyMismatch    = errors.New("currency mismatch")
)

// Auth errors
var (
    ErrInvalidToken  = errors.New("invalid token")
    ErrExpiredToken  = errors.New("token has expired")
    ErrBlockedToken  = errors.New("token has been blocked")
    ErrMissingAuth   = errors.New("missing authorization header")
)

// Session errors
var (
    ErrSessionNotFound = errors.New("session not found")
    ErrSessionBlocked  = errors.New("session has been blocked")
    ErrSessionExpired  = errors.New("session has expired")
)
```

## Error Checking Patterns

### Use errors.Is() for comparison
```go
// GOOD: works with wrapped errors
if errors.Is(err, apperrors.ErrUserNotFound) {
    // handle not found
}

// BAD: breaks with wrapped errors
if err == apperrors.ErrUserNotFound {
    // may not work
}
```

### Wrap errors with context
```go
import "fmt"

func (r *userRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
    var user model.User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user by id %d: %w", id, err)
    }
    return &user, nil
}
```

### Check specific GORM errors
```go
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    err := r.db.WithContext(ctx).Create(user).Error
    if err != nil {
        // Check for duplicate entry (MySQL error 1062)
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
```

## Layer-Specific Error Handling

### Repository Layer
```go
// Convert database errors to domain errors
func (r *accountRepository) GetByID(ctx context.Context, id int64) (*model.Account, error) {
    var account model.Account
    err := r.db.WithContext(ctx).First(&account, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperrors.ErrAccountNotFound
        }
        return nil, fmt.Errorf("db error: %w", err)
    }
    return &account, nil
}
```

### Service Layer
```go
// Add business logic validation errors
func (s *TransferService) CreateTransfer(ctx context.Context, arg *dto.CreateTransferRequest) (*dto.TransferResult, error) {
    // Validate from account
    fromAccount, err := s.accountRepo.GetByID(ctx, arg.FromAccountID)
    if err != nil {
        return nil, err // propagate ErrAccountNotFound
    }

    // Business validation
    if fromAccount.Owner != arg.Username {
        return nil, apperrors.ErrAccountNotBelongTo
    }

    if fromAccount.Balance < arg.Amount {
        return nil, apperrors.ErrInsufficientBalance
    }

    if fromAccount.Currency != arg.Currency {
        return nil, apperrors.ErrCurrencyMismatch
    }

    // ... continue with transfer
}
```

### Handler Layer
```go
// Convert domain errors to HTTP responses
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
    var req dto.CreateTransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    result, err := h.transferService.CreateTransfer(c.Request.Context(), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, result)
}

func (h *TransferHandler) handleError(c *gin.Context, err error) {
    switch {
    // 400 Bad Request
    case errors.Is(err, apperrors.ErrInsufficientBalance),
         errors.Is(err, apperrors.ErrCurrencyMismatch):
        c.JSON(http.StatusBadRequest, errorResponse(err))

    // 401 Unauthorized
    case errors.Is(err, apperrors.ErrInvalidToken),
         errors.Is(err, apperrors.ErrExpiredToken):
        c.JSON(http.StatusUnauthorized, errorResponse(err))

    // 403 Forbidden
    case errors.Is(err, apperrors.ErrAccountNotBelongTo),
         errors.Is(err, apperrors.ErrBlockedToken):
        c.JSON(http.StatusForbidden, errorResponse(err))

    // 404 Not Found
    case errors.Is(err, apperrors.ErrUserNotFound),
         errors.Is(err, apperrors.ErrAccountNotFound):
        c.JSON(http.StatusNotFound, errorResponse(err))

    // 409 Conflict
    case errors.Is(err, apperrors.ErrDuplicateUsername),
         errors.Is(err, apperrors.ErrDuplicateEmail):
        c.JSON(http.StatusConflict, errorResponse(err))

    // 500 Internal Server Error
    default:
        // Log the actual error for debugging
        log.Printf("Internal error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "internal server error",
        })
    }
}
```

## Error Response Format

### Standard Error Response
```go
// internal/dto/error.go
type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}

func errorResponse(err error) ErrorResponse {
    return ErrorResponse{Error: err.Error()}
}
```

### Validation Error Response
```go
func validationErrorResponse(err error) ErrorResponse {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        details := make(map[string]string)
        for _, fe := range ve {
            field := strings.ToLower(fe.Field())
            switch fe.Tag() {
            case "required":
                details[field] = fmt.Sprintf("%s is required", field)
            case "email":
                details[field] = "invalid email format"
            case "min":
                details[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
            case "max":
                details[field] = fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
            case "oneof":
                details[field] = fmt.Sprintf("%s must be one of: %s", field, fe.Param())
            default:
                details[field] = fmt.Sprintf("%s is invalid", field)
            }
        }
        return ErrorResponse{
            Error:   "validation failed",
            Details: details,
        }
    }
    return ErrorResponse{Error: err.Error()}
}
```

## Panic Recovery

### Use middleware for panic recovery
```go
// internal/middleware/recovery.go
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if r := recover(); r != nil {
                // Log the panic
                log.Printf("Panic recovered: %v\n%s", r, debug.Stack())

                c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                    "error": "internal server error",
                })
            }
        }()
        c.Next()
    }
}
```

## Logging Errors

### Log internal errors, not user errors
```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    result, err := h.userService.CreateUser(ctx, &req)
    if err != nil {
        // Log only unexpected errors
        if !isUserError(err) {
            log.Printf("Error creating user: %v", err)
        }
        h.handleError(c, err)
        return
    }
}

func isUserError(err error) bool {
    userErrors := []error{
        apperrors.ErrDuplicateUsername,
        apperrors.ErrDuplicateEmail,
        apperrors.ErrUserNotFound,
        apperrors.ErrInvalidPassword,
    }
    for _, ue := range userErrors {
        if errors.Is(err, ue) {
            return true
        }
    }
    return false
}
```
