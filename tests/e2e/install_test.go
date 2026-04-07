package e2e

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ifuryst/docmesh/internal/db"
	"github.com/ifuryst/docmesh/internal/httpserver"
	"github.com/ifuryst/docmesh/internal/logging"
	"github.com/ifuryst/docmesh/internal/repository"
	"github.com/ifuryst/docmesh/internal/service"
	"github.com/ifuryst/docmesh/internal/testutil"
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

	assertBodyContains("/install/DocMesh.md", "DocMesh CLI Install")
	assertBodyContains("/install/install-cli.sh", "DOCMESH_RELEASE_REPO")
	assertBodyContains("/install/skills/DocMesh.skill", "docmesh/SKILL.md")
}
