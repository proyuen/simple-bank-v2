// Package router 配置 HTTP 路由
// 将 URL 路径映射到对应的 Handler 方法
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/proyuen/simple-bank-v2/internal/handler"
	"github.com/proyuen/simple-bank-v2/internal/middleware"
	"github.com/proyuen/simple-bank-v2/pkg/token"
)

// ==================== Handlers 容器 ====================

// Handlers 包含所有 HTTP Handler 的引用
// 用于在路由配置时注入依赖
type Handlers struct {
	// User Handler 处理用户相关路由
	User *handler.UserHandler

	// Account Handler 处理账户相关路由
	Account *handler.AccountHandler

	// Transfer Handler 处理转账和账目相关路由
	Transfer *handler.TransferHandler
}

// ==================== 路由配置 ====================

// SetupRouter 配置并返回 Gin 路由引擎
//
// 路由结构:
//
//	/api/v1
//	├── /users              (公开)
//	│   ├── POST /          → 用户注册
//	│   └── POST /login     → 用户登录
//	├── /tokens             (公开)
//	│   └── POST /renew     → 刷新 Token
//	├── /accounts           (需认证)
//	│   ├── POST /          → 创建账户
//	│   ├── GET /           → 获取账户列表
//	│   ├── GET /:id        → 获取账户详情
//	│   └── GET /:id/entries → 获取账目记录
//	└── /transfers          (需认证)
//	    ├── POST /          → 创建转账
//	    └── GET /           → 获取转账记录
//
// 参数:
//   - handlers: 包含所有 Handler 的容器
//   - tokenMaker: JWT 验证器，用于认证中间件
//
// 返回:
//   - *gin.Engine: 配置好的 Gin 路由引擎
func SetupRouter(handlers *Handlers, tokenMaker token.Maker) *gin.Engine {
	// 创建默认的 Gin 路由引擎
	// 默认包含 Logger 和 Recovery 中间件
	router := gin.Default()

	// ==================== API V1 路由组 ====================
	// 所有 API 路由都以 /api/v1 为前缀
	// 使用版本号便于 API 升级时保持向后兼容
	v1 := router.Group("/api/v1")

	// ==================== 公开路由 (无需认证) ====================
	// 这些路由任何人都可以访问

	// 用户路由组
	// /api/v1/users
	users := v1.Group("/users")
	{
		// POST /api/v1/users - 用户注册
		// 任何人都可以注册新账户
		users.POST("", handlers.User.CreateUser)

		// POST /api/v1/users/login - 用户登录
		// 返回 Access Token 和 Refresh Token
		users.POST("/login", handlers.User.LoginUser)
	}

	// Token 路由组
	// /api/v1/tokens
	tokens := v1.Group("/tokens")
	{
		// POST /api/v1/tokens/renew - 刷新 Token
		// 使用 Refresh Token 获取新的 Access Token
		// 注意: 这个路由不需要 Access Token，只需要 Refresh Token
		tokens.POST("/renew", handlers.User.RefreshToken)
	}

	// ==================== 受保护路由 (需要认证) ====================
	// 这些路由需要在请求头中携带有效的 Access Token
	// Authorization: Bearer <access_token>

	// 创建认证路由组
	// 应用 AuthMiddleware 中间件
	authRoutes := v1.Group("")
	authRoutes.Use(middleware.AuthMiddleware(tokenMaker))
	{
		// 账户路由组
		// /api/v1/accounts
		accounts := authRoutes.Group("/accounts")
		{
			// POST /api/v1/accounts - 创建账户
			// 为当前用户创建一个新的银行账户
			accounts.POST("", handlers.Account.CreateAccount)

			// GET /api/v1/accounts - 获取账户列表
			// 获取当前用户的所有账户 (支持分页)
			accounts.GET("", handlers.Account.ListAccounts)

			// GET /api/v1/accounts/:id - 获取账户详情
			// 获取指定账户的详细信息
			// 只能查看自己的账户
			accounts.GET("/:id", handlers.Account.GetAccount)

			// GET /api/v1/accounts/:id/entries - 获取账目记录
			// 获取指定账户的所有资金变动记录 (支持分页)
			accounts.GET("/:id/entries", handlers.Transfer.ListEntries)
		}

		// 转账路由组
		// /api/v1/transfers
		transfers := authRoutes.Group("/transfers")
		{
			// POST /api/v1/transfers - 创建转账
			// 从一个账户转账到另一个账户
			// 只能从自己的账户转出
			transfers.POST("", handlers.Transfer.CreateTransfer)

			// GET /api/v1/transfers - 获取转账记录
			// 获取指定账户的转账记录 (支持分页)
			// 需要指定 account_id 参数
			transfers.GET("", handlers.Transfer.ListTransfers)
		}
	}

	return router
}

// ==================== 健康检查路由 ====================

// SetupHealthRoutes 添加健康检查路由
//
// 这些路由用于监控和负载均衡器的健康检查
// 通常不需要认证
//
// 参数:
//   - router: Gin 路由引擎
func SetupHealthRoutes(router *gin.Engine) {
	// GET /health - 健康检查
	// 返回服务状态，用于负载均衡器/Kubernetes 探针
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Simple Bank V2 is running",
		})
	})

	// GET /ready - 就绪检查
	// 可以在这里检查数据库连接等依赖
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ready",
		})
	})
}
