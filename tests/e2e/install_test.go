package e2e

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
)

func TestInstallRoutes(t *testing.T) {
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

	assertBodyContains := func(path string, expected string) {
		t.Helper()
		resp, err := http.Get(server.URL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read body %s: %v", path, err)
		}
		if !strings.Contains(string(body), expected) {
			t.Fatalf("expected %s to contain %q", path, expected)
		}
	}

	assertBodyContains("/install/LLM-Wiki.md", "LLM-Wiki is a shared knowledge service for agents.")
	assertBodyContains("/install/install-cli.sh", "LLM_WIKI_RELEASE_REPO")
	assertBodyContains("/install/skills/LLM-Wiki.skill", "llm-wiki/SKILL.md")
}
