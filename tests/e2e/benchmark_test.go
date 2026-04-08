package e2e

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpclient"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
)

func BenchmarkCreateAndGetDocument(b *testing.B) {
	pg := testutil.StartPostgres(b)
	cfg := testutil.MustBaseConfig(pg.Config)

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.Postgres)
	if err != nil {
		b.Fatalf("open db: %v", err)
	}
	defer pool.Close()
	if err := db.ApplyMigrations(ctx, pool); err != nil {
		b.Fatalf("apply migrations: %v", err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		b.Fatalf("logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	svc := service.New(repository.New(pool))
	token := "test-bench-token"
	if err := bootstrapTestToken(ctx, svc, "tenant-bench", token); err != nil {
		b.Fatalf("bootstrap token: %v", err)
	}
	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer server.Close()

	client := httpclient.New(server.URL, 10*time.Second, token)
	namespace, err := client.CreateNamespace(ctx, api.CreateNamespaceRequest{
		Key:         "perf",
		DisplayName: "Perf",
	})
	if err != nil {
		b.Fatalf("create namespace: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, err := client.CreateDocument(ctx, api.CreateDocumentRequest{
			NamespaceID:   namespace.ID,
			Slug:          testutil.UniqueTenant("doc", i),
			Title:         "Benchmark",
			Content:       "payload",
			AuthorType:    "agent",
			AuthorID:      "bench",
			ChangeSummary: "create",
		})
		if err != nil {
			b.Fatalf("create document: %v", err)
		}
		if _, err := client.GetDocument(ctx, doc.ID); err != nil {
			b.Fatalf("get document: %v", err)
		}
	}
}
