// Package server 提供 HTTP 服务器的初始化和生命周期管理
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/proyuen/simple-bank-v2/internal/config"
	"github.com/proyuen/simple-bank-v2/internal/handler"
	"github.com/proyuen/simple-bank-v2/internal/repository"
	"github.com/proyuen/simple-bank-v2/internal/router"
	"github.com/proyuen/simple-bank-v2/internal/service"
	"github.com/proyuen/simple-bank-v2/pkg/token"
)

// App 封装应用程序的所有依赖
type App struct {
	config     config.Config
	db         *gorm.DB
	tokenMaker token.Maker
	httpServer *http.Server
}

// NewApp 创建并初始化应用程序
func NewApp(cfg config.Config) (*App, error) {
	app := &App{config: cfg}

	if err := app.setupDatabase(); err != nil {
		return nil, fmt.Errorf("setup database: %w", err)
	}

	if err := app.setupTokenMaker(); err != nil {
		return nil, fmt.Errorf("setup token maker: %w", err)
	}

	app.setupHTTPServer()

	return app, nil
}

// setupDatabase 初始化数据库连接
func (a *App) setupDatabase() error {
	var gormLogMode logger.LogLevel
	if a.config.IsProduction() {
		gormLogMode = logger.Error
	} else {
		gormLogMode = logger.Info
	}

	db, err := gorm.Open(mysql.Open(a.config.DBSource()), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogMode),
	})
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(a.config.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(a.config.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(a.config.DBConnMaxLifetime)

	a.db = db
	slog.Info("database connected",
		"host", a.config.DBHost,
		"database", a.config.DBName,
	)
	return nil
}

// setupTokenMaker 初始化 JWT Token 生成器
func (a *App) setupTokenMaker() error {
	tokenMaker, err := token.NewJWTMaker(a.config.TokenSecretKey)
	if err != nil {
		return fmt.Errorf("create token maker: %w", err)
	}
	a.tokenMaker = tokenMaker
	return nil
}

// setupHTTPServer 初始化 HTTP 服务器
func (a *App) setupHTTPServer() {
	if a.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Repositories
	userRepo := repository.NewUserRepository(a.db)
	accountRepo := repository.NewAccountRepository(a.db)
	sessionRepo := repository.NewSessionRepository(a.db)
	transferRepo := repository.NewTransferRepository(a.db)
	entryRepo := repository.NewEntryRepository(a.db)
	txManager := repository.NewTxManager(a.db)

	// 创建 Services
	userService := service.NewUserService(
		userRepo,
		sessionRepo,
		a.tokenMaker,
		a.config.AccessTokenDuration,
		a.config.RefreshTokenDuration,
	)
	accountService := service.NewAccountService(accountRepo)
	transferService := service.NewTransferService(
		txManager,
		accountRepo,
		transferRepo,
		entryRepo,
	)

	// 创建 Handlers
	handlers := &router.Handlers{
		User:     handler.NewUserHandler(userService),
		Account:  handler.NewAccountHandler(accountService),
		Transfer: handler.NewTransferHandler(transferService),
	}

	// 设置路由
	r := router.SetupRouter(handlers, a.tokenMaker)
	router.SetupHealthRoutes(r)

	a.httpServer = &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: r,
	}
}

// Run 启动 HTTP 服务器并等待关闭信号
func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("server starting", "address", a.config.ServerAddress)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")
		return a.shutdown()
	}
}

// shutdown 优雅关闭服务器
func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.ServerShutdownTimeout)
	defer cancel()

	slog.Info("shutting down server", "timeout", a.config.ServerShutdownTimeout)
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

// Close 清理应用程序资源
func (a *App) Close() error {
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err != nil {
			return fmt.Errorf("get sql.DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("close database: %w", err)
		}
		slog.Info("database connection closed")
	}
	return nil
}
