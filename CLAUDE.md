# Project Instructions for Simple Bank V2

Go 银行后端 MVP，使用三层架构 (Handler/Service/Repository)

## Frequently Used Commands

- **Build:** `go build -o bin/server ./cmd/server`
- **Run:** `make server` 或 `go run ./cmd/server`
- **Test:** `go test ./...`
- **Test with coverage:** `go test -cover ./...`
- **Lint:** `golangci-lint run`
- **Format:** `gofmt -w .`
- **Migration up:** `make migrateup`
- **Migration down:** `make migratedown`
- **Generate Swagger:** `swag init -g cmd/server/main.go`

## Code Style Preferences

- Use `gofmt` for all Go files (tabs for indentation)
- Package names: lowercase, single word (`repository`, `service`, `handler`)
- File names: lowercase with underscores (`user_repository.go`, `account_service.go`)
- Exported types/functions: PascalCase (`CreateUser`, `UserService`)
- Private types/functions: camelCase (`validatePassword`, `hashPassword`)
- Interface names: PascalCase, typically ending with "er" (`UserRepository`, `TokenMaker`)
- Error variables: `Err` prefix (`ErrUserNotFound`, `ErrInvalidToken`)
- Constants: PascalCase or ALL_CAPS (`DefaultPageSize`, `MAX_RETRY_COUNT`)
- Always pass `context.Context` as first parameter in functions that do I/O
- Use `binding` tags for Gin request validation
- Return `error` as the last return value

## Architectural Patterns

### Three-Layer Architecture
```
Handler  → HTTP request/response only, no business logic
Service  → Business logic only, no direct database access
Repository → Database CRUD only, no business logic
```

### Dependency Injection
- Use interfaces for dependencies, not concrete implementations
- Inject dependencies via constructor functions (`NewUserService(repo UserRepository)`)

### DTO and Model Separation
- `internal/model/`: GORM database models
- `internal/dto/`: Request/Response structures (never expose sensitive fields like passwords)

### Error Handling
- Define custom errors in `internal/errors/`
- Handler layer converts errors to HTTP status codes
- Use `errors.Is()` for error comparison

### Transaction Handling
- Handle transactions in Service layer using `execTx` pattern
- Never start transactions in Repository layer

## Project Structure

```
simple-bank-v2/
├── cmd/server/           # Application entry point (main.go)
├── internal/             # Private application code
│   ├── config/           # Configuration loading (Viper)
│   ├── handler/          # HTTP handlers (Gin)
│   ├── service/          # Business logic
│   ├── repository/       # Data access (GORM)
│   ├── model/            # GORM models
│   ├── dto/              # Request/Response DTOs
│   ├── errors/           # Custom error types
│   └── middleware/       # HTTP middleware (auth, logging)
├── pkg/                  # Reusable packages
│   ├── token/            # JWT token generation/validation
│   └── password/         # Password hashing (bcrypt)
├── db/migration/         # SQL migration files (MySQL)
└── docs/                 # API documentation (Swagger)
```

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.21+ | Backend development |
| Web Framework | Gin | REST API |
| ORM | GORM | Database operations |
| Database | MySQL 8.0+ | Data storage |
| Auth | JWT | User authentication |
| Config | Viper | Environment variable management |

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /api/v1/users | User registration | No |
| POST | /api/v1/users/login | User login | No |
| POST | /api/v1/tokens/renew | Refresh access token | No |
| POST | /api/v1/accounts | Create account | Yes |
| GET | /api/v1/accounts/:id | Get account by ID | Yes |
| GET | /api/v1/accounts | List accounts | Yes |
| POST | /api/v1/transfers | Create transfer | Yes |

## Git Workflow

- Use feature branches for all new development (`feature/add-user-auth`)
- Commit messages: `type: description` (e.g., `feat: add user registration endpoint`)
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
- Run `go test ./...` before committing
- Run `golangci-lint run` before creating PR

## Implementation Order

1. `db/migration/` - Database migrations
2. `internal/config/` - Configuration loading
3. `internal/model/` - GORM models
4. `internal/repository/` - Repository interfaces and implementations
5. `internal/dto/` - Request/Response structures
6. `internal/errors/` - Custom error types
7. `pkg/` - Utility packages (token, password)
8. `internal/service/` - Business logic
9. `internal/middleware/` - Auth middleware
10. `internal/handler/` - HTTP handlers
11. `cmd/server/main.go` - Application entry point

## Detailed Guidelines

For comprehensive implementation details, see the modular rules:

- Code Style @.claude/rules/code-style.md
- Error Handling @.claude/rules/error-handling.md
- API Design @.claude/rules/api-design.md
- Database @.claude/rules/database.md
- Testing @.claude/rules/testing.md
- Authentication @.claude/rules/authentication.md
