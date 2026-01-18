# Testing Guidelines

## Test File Organization

### File Naming
```
user_repository.go      → user_repository_test.go
account_service.go      → account_service_test.go
user_handler.go         → user_handler_test.go
```

### Package Naming
```go
// Same package for white-box testing (access private functions)
package repository

// Or use _test suffix for black-box testing (test public API only)
package repository_test
```

## Test Function Naming

### Pattern: Test<Function>_<Scenario>
```go
func TestCreateUser_Success(t *testing.T) {}
func TestCreateUser_DuplicateUsername(t *testing.T) {}
func TestCreateUser_DuplicateEmail(t *testing.T) {}
func TestCreateUser_InvalidEmail(t *testing.T) {}

func TestGetByID_Success(t *testing.T) {}
func TestGetByID_NotFound(t *testing.T) {}

func TestLogin_Success(t *testing.T) {}
func TestLogin_WrongPassword(t *testing.T) {}
func TestLogin_UserNotFound(t *testing.T) {}
```

## Table-Driven Tests

### Basic Structure
```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   dto.CreateUserRequest
        wantErr error
    }{
        {
            name: "success",
            input: dto.CreateUserRequest{
                Username: "testuser",
                Password: "password123",
                FullName: "Test User",
                Email:    "test@example.com",
            },
            wantErr: nil,
        },
        {
            name: "duplicate username",
            input: dto.CreateUserRequest{
                Username: "existinguser", // already exists
                Password: "password123",
                FullName: "Test User",
                Email:    "new@example.com",
            },
            wantErr: apperrors.ErrDuplicateUsername,
        },
        {
            name: "invalid email",
            input: dto.CreateUserRequest{
                Username: "newuser",
                Password: "password123",
                FullName: "Test User",
                Email:    "invalid-email",
            },
            wantErr: apperrors.ErrInvalidEmail,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            svc := setupUserService(t)

            // Execute
            _, err := svc.CreateUser(context.Background(), &tt.input)

            // Assert
            if tt.wantErr != nil {
                require.Error(t, err)
                require.ErrorIs(t, err, tt.wantErr)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## Mocking with Interfaces

### Define Mock
```go
// internal/repository/mock/user_repository.go
package mock

import (
    "context"

    "github.com/stretchr/testify/mock"

    "github.com/proyuen/simple-bank-v2/internal/model"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    args := m.Called(ctx, username)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}
```

### Use Mock in Tests
```go
func TestUserService_Login_Success(t *testing.T) {
    // Setup mock
    mockRepo := new(mock.MockUserRepository)
    mockTokenMaker := new(mock.MockTokenMaker)

    hashedPassword, _ := password.Hash("correct_password")
    expectedUser := &model.User{
        ID:             1,
        Username:       "testuser",
        HashedPassword: hashedPassword,
        Email:          "test@example.com",
    }

    // Setup expectations
    mockRepo.On("GetByUsername", mock.Anything, "testuser").
        Return(expectedUser, nil)
    mockTokenMaker.On("CreateToken", "testuser", mock.Anything).
        Return("access_token", &token.Payload{}, nil)

    // Create service with mocks
    svc := service.NewUserService(mockRepo, mockTokenMaker, nil)

    // Execute
    req := &dto.LoginRequest{
        Username: "testuser",
        Password: "correct_password",
    }
    resp, err := svc.Login(context.Background(), req)

    // Assert
    require.NoError(t, err)
    require.NotNil(t, resp)
    require.Equal(t, "testuser", resp.User.Username)
    require.Equal(t, "access_token", resp.AccessToken)

    // Verify mock expectations
    mockRepo.AssertExpectations(t)
    mockTokenMaker.AssertExpectations(t)
}
```

## HTTP Handler Tests

### Using httptest
```go
func TestCreateUserHandler_Success(t *testing.T) {
    // Setup
    mockService := new(mock.MockUserService)
    handler := handler.NewUserHandler(mockService)

    expectedResp := &dto.UserResponse{
        ID:       1,
        Username: "testuser",
        Email:    "test@example.com",
    }
    mockService.On("CreateUser", mock.Anything, mock.Anything).
        Return(expectedResp, nil)

    // Create request
    reqBody := `{
        "username": "testuser",
        "password": "password123",
        "full_name": "Test User",
        "email": "test@example.com"
    }`

    req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    // Setup router
    router := gin.New()
    router.POST("/api/v1/users", handler.CreateUser)

    // Execute
    router.ServeHTTP(w, req)

    // Assert
    require.Equal(t, http.StatusCreated, w.Code)

    var resp dto.UserResponse
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)
    require.Equal(t, "testuser", resp.Username)
}

func TestCreateUserHandler_ValidationError(t *testing.T) {
    handler := handler.NewUserHandler(nil) // service not needed for validation

    // Invalid request - missing required fields
    reqBody := `{"username": "ab"}` // too short

    req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router := gin.New()
    router.POST("/api/v1/users", handler.CreateUser)
    router.ServeHTTP(w, req)

    require.Equal(t, http.StatusBadRequest, w.Code)
}
```

## Integration Tests

### Database Setup with Test Container
```go
// internal/repository/integration_test.go
package repository_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
    ctx := context.Background()

    // Start MySQL container
    req := testcontainers.ContainerRequest{
        Image:        "mysql:8.0",
        ExposedPorts: []string{"3306/tcp"},
        Env: map[string]string{
            "MYSQL_ROOT_PASSWORD": "testpass",
            "MYSQL_DATABASE":      "testdb",
        },
        WaitingFor: wait.ForLog("ready for connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "3306")

    dsn := fmt.Sprintf("root:testpass@tcp(%s:%s)/testdb?parseTime=true", host, port.Port())

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    require.NoError(t, err)

    // Run migrations
    err = db.AutoMigrate(&model.User{}, &model.Account{}, &model.Entry{}, &model.Transfer{})
    require.NoError(t, err)

    return db
}

func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db := setupTestDB(t)
    repo := repository.NewUserRepository(db)

    t.Run("CreateAndGet", func(t *testing.T) {
        user := &model.User{
            Username:       "integrationuser",
            HashedPassword: "hashedpass",
            FullName:       "Integration User",
            Email:          "integration@example.com",
        }

        err := repo.Create(context.Background(), user)
        require.NoError(t, err)
        require.NotZero(t, user.ID)

        found, err := repo.GetByID(context.Background(), user.ID)
        require.NoError(t, err)
        require.Equal(t, user.Username, found.Username)
    })
}
```

## Test Utilities

### Random Data Generators
```go
// internal/testutil/random.go
package testutil

import (
    "math/rand"
    "strings"
    "time"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func RandomString(n int) string {
    var sb strings.Builder
    for i := 0; i < n; i++ {
        sb.WriteByte(alphabet[rand.Intn(len(alphabet))])
    }
    return sb.String()
}

func RandomUsername() string {
    return RandomString(8)
}

func RandomEmail() string {
    return RandomString(6) + "@example.com"
}

func RandomMoney() int64 {
    return rand.Int63n(1000000)
}

func RandomCurrency() string {
    currencies := []string{"USD", "EUR", "CNY"}
    return currencies[rand.Intn(len(currencies))]
}
```

### Test Fixtures
```go
// internal/testutil/fixtures.go
package testutil

import (
    "github.com/proyuen/simple-bank-v2/internal/model"
    "github.com/proyuen/simple-bank-v2/pkg/password"
)

func CreateTestUser(t *testing.T, db *gorm.DB) *model.User {
    hashedPass, _ := password.Hash("testpassword")

    user := &model.User{
        Username:       RandomUsername(),
        HashedPassword: hashedPass,
        FullName:       "Test User",
        Email:          RandomEmail(),
    }

    err := db.Create(user).Error
    require.NoError(t, err)

    return user
}

func CreateTestAccount(t *testing.T, db *gorm.DB, owner string) *model.Account {
    account := &model.Account{
        Owner:    owner,
        Balance:  RandomMoney(),
        Currency: RandomCurrency(),
    }

    err := db.Create(account).Error
    require.NoError(t, err)

    return account
}
```

## Test Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/service/...

# Run specific test function
go test -run TestCreateUser ./internal/service/

# Skip integration tests
go test -short ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run with race detection
go test -race ./...
```

## Required Test Packages

```go
// go.mod
require (
    github.com/stretchr/testify v1.8.4
    github.com/testcontainers/testcontainers-go v0.27.0
)
```

## Test Coverage Goals

| Layer | Minimum Coverage |
|-------|-----------------|
| Repository | 80% |
| Service | 85% |
| Handler | 75% |
| Utils/Helpers | 90% |
