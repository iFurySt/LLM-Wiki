package app

import (
	"context"

	"github.com/bytedance/docmesh/internal/config"
	"github.com/bytedance/docmesh/internal/db"
	"github.com/bytedance/docmesh/internal/httpserver"
	"github.com/bytedance/docmesh/internal/repository"
	"github.com/bytedance/docmesh/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	server *httpserver.Server
	db     *pgxpool.Pool
}

func New(cfg config.Config, logger *zap.Logger) (*App, error) {
	pool, err := db.Open(context.Background(), cfg.Postgres)
	if err != nil {
		return nil, err
	}
	if cfg.AutoMigrate {
		if err := db.ApplyMigrations(context.Background(), pool); err != nil {
			pool.Close()
			return nil, err
		}
	}
	repo := repository.New(pool)
	svc := service.New(repo)

	return &App{
		server: httpserver.New(cfg, logger, svc),
		db:     pool,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.db.Close()
	return a.server.Run(ctx)
}
