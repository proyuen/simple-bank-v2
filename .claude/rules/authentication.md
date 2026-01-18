# JWT Authentication Guidelines

## Token Types

| Token | Purpose | Duration | Storage |
|-------|---------|----------|---------|
| Access Token | API authentication | 15 minutes | Client memory |
| Refresh Token | Get new access token | 24 hours | Client storage + DB |

## Token Maker Interface

### Interface Definition
```go
// pkg/token/maker.go
package token

import "time"

type Maker interface {
    // CreateToken creates a new token for a specific username and duration
    CreateToken(username string, duration time.Duration) (string, *Payload, error)

    // VerifyToken checks if the token is valid or not
    VerifyToken(token string) (*Payload, error)
}
```

### Payload Structure
```go
// pkg/token/payload.go
package token

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

var (
    ErrExpiredToken = errors.New("token has expired")
    ErrInvalidToken = errors.New("token is invalid")
)

type Payload struct {
    ID        uuid.UUID `json:"id"`
    Username  string    `json:"username"`
    IssuedAt  time.Time `json:"issued_at"`
    ExpiredAt time.Time `json:"expired_at"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
    tokenID, err := uuid.NewRandom()
    if err != nil {
        return nil, err
    }

    payload := &Payload{
        ID:        tokenID,
        Username:  username,
        IssuedAt:  time.Now(),
        ExpiredAt: time.Now().Add(duration),
    }

    return payload, nil
}

func (p *Payload) Valid() error {
    if time.Now().After(p.ExpiredAt) {
        return ErrExpiredToken
    }
    return nil
}
```

## JWT Implementation

### JWT Maker
```go
// pkg/token/jwt_maker.go
package token

import (
    "errors"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
    secretKey string
}

func NewJWTMaker(secretKey string) (Maker, error) {
    if len(secretKey) < minSecretKeySize {
        return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
    }
    return &JWTMaker{secretKey: secretKey}, nil
}

func (m *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
    payload, err := NewPayload(username, duration)
    if err != nil {
        return "", nil, err
    }

    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "id":         payload.ID.String(),
        "username":   payload.Username,
        "issued_at":  payload.IssuedAt.Unix(),
        "expired_at": payload.ExpiredAt.Unix(),
    })

    token, err := jwtToken.SignedString([]byte(m.secretKey))
    if err != nil {
        return "", nil, err
    }

    return token, payload, nil
}

func (m *JWTMaker) VerifyToken(token string) (*Payload, error) {
    jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(m.secretKey), nil
    })

    if err != nil {
        return nil, ErrInvalidToken
    }

    claims, ok := jwtToken.Claims.(jwt.MapClaims)
    if !ok || !jwtToken.Valid {
        return nil, ErrInvalidToken
    }

    id, err := uuid.Parse(claims["id"].(string))
    if err != nil {
        return nil, ErrInvalidToken
    }

    payload := &Payload{
        ID:        id,
        Username:  claims["username"].(string),
        IssuedAt:  time.Unix(int64(claims["issued_at"].(float64)), 0),
        ExpiredAt: time.Unix(int64(claims["expired_at"].(float64)), 0),
    }

    if err := payload.Valid(); err != nil {
        return nil, err
    }

    return payload, nil
}
```

## Session Management

### Session Model
```go
// internal/model/session.go
type Session struct {
    ID           string    `gorm:"type:char(36);primaryKey"`
    Username     string    `gorm:"type:varchar(255);not null;index"`
    RefreshToken string    `gorm:"type:varchar(512);not null"`
    UserAgent    string    `gorm:"type:varchar(255);not null;default:''"`
    ClientIP     string    `gorm:"type:varchar(45);not null;default:''"`
    IsBlocked    bool      `gorm:"not null;default:false"`
    ExpiresAt    time.Time `gorm:"not null"`
    CreatedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
```

### Session Repository
```go
// internal/repository/session_repository.go
type SessionRepository interface {
    Create(ctx context.Context, session *model.Session) error
    GetByID(ctx context.Context, id string) (*model.Session, error)
    BlockSession(ctx context.Context, id string) error
    BlockAllUserSessions(ctx context.Context, username string) error
    DeleteExpired(ctx context.Context) error
}
```

## Login Flow

```go
// internal/service/user_service.go
func (s *UserService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
    // 1. Get user by username
    user, err := s.userRepo.GetByUsername(ctx, req.Username)
    if err != nil {
        return nil, err
    }

    // 2. Verify password
    if err := password.Check(req.Password, user.HashedPassword); err != nil {
        return nil, apperrors.ErrInvalidPassword
    }

    // 3. Create access token
    accessToken, accessPayload, err := s.tokenMaker.CreateToken(
        user.Username,
        s.config.AccessTokenDuration,
    )
    if err != nil {
        return nil, err
    }

    // 4. Create refresh token
    refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(
        user.Username,
        s.config.RefreshTokenDuration,
    )
    if err != nil {
        return nil, err
    }

    // 5. Save session to database
    session := &model.Session{
        ID:           refreshPayload.ID.String(),
        Username:     user.Username,
        RefreshToken: refreshToken,
        UserAgent:    req.UserAgent,
        ClientIP:     req.ClientIP,
        IsBlocked:    false,
        ExpiresAt:    refreshPayload.ExpiredAt,
    }

    if err := s.sessionRepo.Create(ctx, session); err != nil {
        return nil, err
    }

    // 6. Return response
    return &dto.LoginResponse{
        SessionID:             session.ID,
        AccessToken:           accessToken,
        AccessTokenExpiresAt:  accessPayload.ExpiredAt,
        RefreshToken:          refreshToken,
        RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
        User:                  dto.NewUserResponse(user),
    }, nil
}
```

## Token Renewal Flow

```go
// internal/service/token_service.go
func (s *TokenService) RenewAccessToken(ctx context.Context, req *dto.RenewTokenRequest) (*dto.RenewTokenResponse, error) {
    // 1. Verify refresh token
    refreshPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
    if err != nil {
        return nil, err
    }

    // 2. Get session from database
    session, err := s.sessionRepo.GetByID(ctx, refreshPayload.ID.String())
    if err != nil {
        return nil, err
    }

    // 3. Check if session is blocked
    if session.IsBlocked {
        return nil, apperrors.ErrSessionBlocked
    }

    // 4. Check if session is expired
    if time.Now().After(session.ExpiresAt) {
        return nil, apperrors.ErrSessionExpired
    }

    // 5. Verify session belongs to correct user
    if session.Username != refreshPayload.Username {
        return nil, apperrors.ErrInvalidToken
    }

    // 6. Verify refresh token matches
    if session.RefreshToken != req.RefreshToken {
        return nil, apperrors.ErrInvalidToken
    }

    // 7. Create new access token
    accessToken, accessPayload, err := s.tokenMaker.CreateToken(
        refreshPayload.Username,
        s.config.AccessTokenDuration,
    )
    if err != nil {
        return nil, err
    }

    return &dto.RenewTokenResponse{
        AccessToken:          accessToken,
        AccessTokenExpiresAt: accessPayload.ExpiredAt,
    }, nil
}
```

## Auth Middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    apperrors "github.com/proyuen/simple-bank-v2/internal/errors"
    "github.com/proyuen/simple-bank-v2/pkg/token"
)

const (
    AuthorizationHeaderKey  = "Authorization"
    AuthorizationTypeBearer = "bearer"
    AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader(AuthorizationHeaderKey)
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": apperrors.ErrMissingAuth.Error(),
            })
            return
        }

        fields := strings.Fields(authHeader)
        if len(fields) != 2 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "invalid authorization header format",
            })
            return
        }

        authType := strings.ToLower(fields[0])
        if authType != AuthorizationTypeBearer {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "unsupported authorization type",
            })
            return
        }

        accessToken := fields[1]
        payload, err := tokenMaker.VerifyToken(accessToken)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": err.Error(),
            })
            return
        }

        c.Set(AuthorizationPayloadKey, payload)
        c.Next()
    }
}
```

## Using Auth Payload in Handlers

```go
// internal/handler/account_handler.go
func (h *AccountHandler) CreateAccount(c *gin.Context) {
    // Get authenticated user from context
    authPayload := c.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)

    var req dto.CreateAccountRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, errorResponse(err))
        return
    }

    // Set owner from authenticated user
    req.Owner = authPayload.Username

    // ... continue with account creation
}
```

## Password Change - Invalidate Sessions

```go
// internal/service/user_service.go
func (s *UserService) ChangePassword(ctx context.Context, username string, req *dto.ChangePasswordRequest) error {
    // 1. Get user
    user, err := s.userRepo.GetByUsername(ctx, username)
    if err != nil {
        return err
    }

    // 2. Verify old password
    if err := password.Check(req.OldPassword, user.HashedPassword); err != nil {
        return apperrors.ErrInvalidPassword
    }

    // 3. Hash new password
    hashedPassword, err := password.Hash(req.NewPassword)
    if err != nil {
        return err
    }

    // 4. Update user
    user.HashedPassword = hashedPassword
    user.PasswordChangedAt = time.Now()

    if err := s.userRepo.Update(ctx, user); err != nil {
        return err
    }

    // 5. Block all existing sessions (force re-login)
    if err := s.sessionRepo.BlockAllUserSessions(ctx, username); err != nil {
        return err
    }

    return nil
}
```

## Security Best Practices

1. **Secret Key**: Minimum 32 characters, stored in environment variable
2. **Token Duration**: Access token short (15min), refresh token longer (24h)
3. **HTTPS Only**: Always use HTTPS in production
4. **Session Storage**: Store refresh tokens in database for revocation
5. **Password Change**: Invalidate all sessions when password changes
6. **Token Rotation**: Consider rotating refresh tokens on each use
