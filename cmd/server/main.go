package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/proyuen/simple-bank-v2/internal/config"
	"github.com/proyuen/simple-bank-v2/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 加载配置
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 设置日志
	setupLogger(cfg.IsProduction())

	// 创建应用
	app, err := server.NewApp(cfg)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}
	defer func() {
		if err := app.Close(); err != nil {
			slog.Error("close app", "error", err)
		}
	}()

	// 设置信号处理
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 启动服务器
	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	slog.Info("server exited gracefully")
	return nil
}

func setupLogger(production bool) {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if production {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
