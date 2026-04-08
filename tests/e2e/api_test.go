package e2e

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpclient"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
)

func TestNamespaceAndDocumentCRUD(t *testing.T) {
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
	handler := httpserver.NewHandler(cfg, logger, svc)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := httpclient.New(server.URL, 10_000_000_000, "tenant-e2e")

	namespace, err := client.CreateNamespace(ctx, api.CreateNamespaceRequest{
		Key:         "org",
		DisplayName: "Org",
		Description: "tenant-wide knowledge",
		Visibility:  "tenant",
	})
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}

	document, err := client.CreateDocument(ctx, api.CreateDocumentRequest{
		NamespaceID:   namespace.ID,
		Slug:          "product-brief",
		Title:         "Product Brief",
		Content:       "# LLM-Wiki",
		AuthorType:    "agent",
		AuthorID:      "bootstrap-agent",
		ChangeSummary: "initial draft",
	})
	if err != nil {
		t.Fatalf("create document: %v", err)
	}
	if document.CurrentRevisionNo != 1 {
		t.Fatalf("expected revision 1, got %d", document.CurrentRevisionNo)
	}

	updated, err := client.UpdateDocument(ctx, document.ID, api.UpdateDocumentRequest{
		Title:         "Product Brief v2",
		Content:       "# LLM-Wiki\n\nUpdated.",
		AuthorType:    "user",
		AuthorID:      "tester",
		ChangeSummary: "clarify scope",
	})
	if err != nil {
		t.Fatalf("update document: %v", err)
	}
	if updated.CurrentRevisionNo != 2 {
		t.Fatalf("expected revision 2, got %d", updated.CurrentRevisionNo)
	}

	fetched, err := client.GetDocument(ctx, document.ID)
	if err != nil {
		t.Fatalf("get document: %v", err)
	}
	if len(fetched.Revisions) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(fetched.Revisions))
	}
	if fetched.Title != "Product Brief v2" {
		t.Fatalf("unexpected title: %q", fetched.Title)
	}

	bySlug, err := client.GetDocumentBySlug(ctx, namespace.ID, "product-brief")
	if err != nil {
		t.Fatalf("get document by slug: %v", err)
	}
	if bySlug.ID != document.ID {
		t.Fatalf("slug lookup returned wrong document: %d", bySlug.ID)
	}

	listed, err := client.ListDocuments(ctx, &namespace.ID, nil)
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if len(listed.Items) != 1 {
		t.Fatalf("expected 1 listed document, got %d", len(listed.Items))
	}

	archived, err := client.ArchiveDocument(ctx, document.ID, api.ArchiveDocumentRequest{
		AuthorType:    "agent",
		AuthorID:      "archiver",
		ChangeSummary: "archive after review",
	})
	if err != nil {
		t.Fatalf("archive document: %v", err)
	}
	if archived.Status != "archived" {
		t.Fatalf("expected archived status, got %q", archived.Status)
	}
	if archived.CurrentRevisionNo != 3 {
		t.Fatalf("expected revision 3 after archive, got %d", archived.CurrentRevisionNo)
	}
}
