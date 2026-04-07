package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/bytedance/docmesh/internal/app"
	"github.com/bytedance/docmesh/internal/config"
	"github.com/bytedance/docmesh/internal/logging"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("initialize application", logging.Error(err))
	}

	if err := application.Run(ctx); err != nil {
		logger.Fatal("run application", logging.Error(err))
	}
}
