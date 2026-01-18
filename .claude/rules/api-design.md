# API Design Guidelines

## RESTful Endpoint Naming

### Resource Naming
```
# GOOD: plural nouns, lowercase, hyphens for multi-word
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
GET    /api/v1/bank-accounts    # multi-word resource

# BAD: verbs, underscores, singular
GET    /api/v1/getUser          # verb in URL
POST   /api/v1/create_user      # underscore
GET    /api/v1/user/:id         # singular
```

### Nested Resources
```
# Account belongs to user
GET    /api/v1/users/:username/accounts
POST   /api/v1/users/:username/accounts

# Transfers between accounts
GET    /api/v1/accounts/:id/transfers
POST   /api/v1/transfers
```

### Action Endpoints (Non-CRUD)
```
# Use verb for actions that don't fit CRUD
POST   /api/v1/users/login
POST   /api/v1/tokens/renew
POST   /api/v1/accounts/:id/lock
POST   /api/v1/accounts/:id/unlock
```

## HTTP Methods

| Method | Usage | Idempotent | Request Body |
|--------|-------|------------|--------------|
| GET | Read resource(s) | Yes | No |
| POST | Create resource | No | Yes |
| PUT | Replace resource | Yes | Yes |
| PATCH | Partial update | Yes | Yes |
| DELETE | Delete resource | Yes | No |

## HTTP Status Codes

### Success Responses
```go
200 OK           // GET, PUT, PATCH success
201 Created      // POST success (resource created)
204 No Content   // DELETE success
```

### Client Error Responses
```go
400 Bad Request      // Invalid request body, validation failed
401 Unauthorized     // Missing or invalid authentication
403 Forbidden        // Authenticated but not authorized
404 Not Found        // Resource does not exist
409 Conflict         // Duplicate resource (username, email)
422 Unprocessable    // Semantic errors (insufficient balance)
```

### Server Error Responses
```go
500 Internal Server Error  // Unexpected server error
503 Service Unavailable    // Server overloaded or maintenance
```

## Request/Response Structure

### Request DTO Patterns
```go
// Create request - all required fields
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50,alphanum"`
    Password string `json:"password" binding:"required,min=6,max=100"`
    FullName string `json:"full_name" binding:"required,min=1,max=100"`
    Email    string `json:"email" binding:"required,email"`
}

// Update request - optional fields with pointers
type UpdateUserRequest struct {
    FullName *string `json:"full_name,omitempty" binding:"omitempty,min=1,max=100"`
    Email    *string `json:"email,omitempty" binding:"omitempty,email"`
}

// Query parameters for list endpoints
type ListAccountsRequest struct {
    PageID   int32 `form:"page_id" binding:"required,min=1"`
    PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// URI parameters
type GetAccountRequest struct {
    ID int64 `uri:"id" binding:"required,min=1"`
}
```

### Response DTO Patterns
```go
// Single resource response
type UserResponse struct {
    ID                int64     `json:"id"`
    Username          string    `json:"username"`
    FullName          string    `json:"full_name"`
    Email             string    `json:"email"`
    PasswordChangedAt time.Time `json:"password_changed_at"`
    CreatedAt         time.Time `json:"created_at"`
}
// Note: never include hashed_password

// List response with pagination
type ListAccountsResponse struct {
    Accounts   []AccountResponse `json:"accounts"`
    TotalCount int64             `json:"total_count"`
    PageID     int32             `json:"page_id"`
    PageSize   int32             `json:"page_size"`
}

// Login response with tokens
type LoginResponse struct {
    User                  UserResponse `json:"user"`
    SessionID             string       `json:"session_id"`
    AccessToken           string       `json:"access_token"`
    AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
    RefreshToken          string       `json:"refresh_token"`
    RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
}

// Error response
type ErrorResponse struct {
    Error   string            `json:"error"`
    Details map[string]string `json:"details,omitempty"`
}
```

## API Versioning

### URL Path Versioning
```
/api/v1/users
/api/v2/users
```

### Router Setup
```go
func SetupRouter(h *Handlers) *gin.Engine {
    router := gin.Default()

    v1 := router.Group("/api/v1")
    {
        // Public routes
        v1.POST("/users", h.User.CreateUser)
        v1.POST("/users/login", h.User.Login)
        v1.POST("/tokens/renew", h.Token.RenewAccessToken)

        // Protected routes
        authRoutes := v1.Group("/").Use(authMiddleware)
        {
            authRoutes.POST("/accounts", h.Account.CreateAccount)
            authRoutes.GET("/accounts/:id", h.Account.GetAccount)
            authRoutes.GET("/accounts", h.Account.ListAccounts)
            authRoutes.POST("/transfers", h.Transfer.CreateTransfer)
        }
    }

    return router
}
```

## Authentication

### Authorization Header
```
Authorization: Bearer <access_token>
```

### Auth Middleware
```go
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrMissingAuth))
            return
        }

        fields := strings.Fields(authHeader)
        if len(fields) != 2 || strings.ToLower(fields[0]) != "bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(ErrInvalidAuthFormat))
            return
        }

        payload, err := tokenMaker.VerifyToken(fields[1])
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
            return
        }

        // Store payload in context for handlers
        c.Set("authorization_payload", payload)
        c.Next()
    }
}
```

### Accessing Auth Payload in Handler
```go
func (h *AccountHandler) CreateAccount(c *gin.Context) {
    payload := c.MustGet("authorization_payload").(*token.Payload)
    username := payload.Username

    // Use username for authorization
}
```

## Pagination

### Request Parameters
```go
type PaginationRequest struct {
    PageID   int32 `form:"page_id" binding:"required,min=1"`
    PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}
```

### Calculate Offset
```go
func (r *PaginationRequest) Offset() int {
    return int((r.PageID - 1) * r.PageSize)
}

func (r *PaginationRequest) Limit() int {
    return int(r.PageSize)
}
```

### Repository Implementation
```go
func (r *accountRepository) List(ctx context.Context, owner string, limit, offset int) ([]model.Account, int64, error) {
    var accounts []model.Account
    var total int64

    query := r.db.WithContext(ctx).Where("owner = ?", owner)

    // Get total count
    if err := query.Model(&model.Account{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated results
    if err := query.Limit(limit).Offset(offset).Find(&accounts).Error; err != nil {
        return nil, 0, err
    }

    return accounts, total, nil
}
```

## JSON Field Naming

### Use snake_case for JSON
```go
type TransferRequest struct {
    FromAccountID int64  `json:"from_account_id"`  // snake_case
    ToAccountID   int64  `json:"to_account_id"`
    Amount        int64  `json:"amount"`
    Currency      string `json:"currency"`
}
```

### Consistent Timestamp Format
```go
// Use RFC3339 format (Go default for time.Time)
{
    "created_at": "2024-01-15T10:30:00Z",
    "expires_at": "2024-01-15T11:30:00Z"
}
```

## Handler Template

```go
// internal/handler/account_handler.go
type AccountHandler struct {
    accountService service.AccountService
}

func NewAccountHandler(as service.AccountService) *AccountHandler {
    return &AccountHandler{accountService: as}
}

// CreateAccount godoc
// @Summary Create a new account
// @Tags accounts
// @Accept json
// @Produce json
// @Param request body dto.CreateAccountRequest true "Create account request"
// @Success 201 {object} dto.AccountResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /accounts [post]
// @Security BearerAuth
func (h *AccountHandler) CreateAccount(c *gin.Context) {
    // 1. Get auth payload
    payload := c.MustGet("authorization_payload").(*token.Payload)

    // 2. Bind request
    var req dto.CreateAccountRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, validationErrorResponse(err))
        return
    }

    // 3. Set owner from token
    req.Owner = payload.Username

    // 4. Call service
    account, err := h.accountService.CreateAccount(c.Request.Context(), &req)
    if err != nil {
        handleError(c, err)
        return
    }

    // 5. Return response
    c.JSON(http.StatusCreated, account)
}
```
