# Simple Bank V2

一个使用 Go 语言构建的简单银行后端 API。

## 功能特性

- 👤 用户注册和登录
- 🏦 银行账户管理
- 💸 账户间转账
- 🔐 JWT 认证

## 技术栈

- **Go 1.21+** - 编程语言
- **Gin** - Web 框架
- **GORM** - ORM
- **PostgreSQL** - 数据库
- **JWT** - 认证

## 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Make

### 启动服务

```bash
# 1. 克隆项目
git clone https://github.com/yuanko/simple-bank-v2.git
cd simple-bank-v2

# 2. 启动 PostgreSQL
docker-compose up -d postgres

# 3. 执行数据库迁移
make migrateup

# 4. 启动服务
make server
```

### API 测试

```bash
# 注册用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123","full_name":"Test User","email":"test@example.com"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'
```

## API 文档

启动服务后访问: http://localhost:8080/swagger/index.html

## 项目结构

```
├── cmd/server/       # 应用入口
├── internal/         # 私有业务代码
│   ├── handler/      # HTTP 处理器
│   ├── service/      # 业务逻辑
│   ├── repository/   # 数据访问
│   └── model/        # 数据模型
├── pkg/              # 可复用包
└── db/migration/     # 数据库迁移
```

## 许可证

MIT License
