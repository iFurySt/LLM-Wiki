package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
)

func TestReadyz(t *testing.T) {
	t.Parallel()

	pg := testutil.StartPostgres(t)
	cfg := testutil.MustBaseConfig(pg.Config)

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.Postgres)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer pool.Close()

	if err := db.ApplyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	svc := service.New(repository.New(pool))
	if _, err := svc.Initialize(ctx, api.InitializeRequest{
		TenantID:    "default",
		Username:    "admin",
		DisplayName: "Admin",
		Password:    "secret123",
	}); err != nil {
		t.Fatalf("initialize workspace: %v", err)
	}
	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer server.Close()

	resp, err := http.Get(server.URL + "/readyz")
	if err != nil {
		t.Fatalf("readyz request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestUIIndex(t *testing.T) {
	t.Parallel()

	pg := testutil.StartPostgres(t)
	cfg := testutil.MustBaseConfig(pg.Config)

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.Postgres)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer pool.Close()

	if err := db.ApplyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, service.New(repository.New(pool))))
	defer server.Close()

	resp, err := http.Get(server.URL + "/ui")
	if err != nil {
		t.Fatalf("ui request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
