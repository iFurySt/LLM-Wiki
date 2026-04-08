package e2e

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPStreamableHTTP(t *testing.T) {
	t.Parallel()

	serverURL := newMCPTestServer(t)
	session := connectMCPClient(t, &mcp.StreamableClientTransport{
		Endpoint:   serverURL + "/mcp",
		HTTPClient: tenantHTTPClient("tenant-mcp"),
	})
	defer session.Close()

	tools, err := session.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools.Tools) == 0 {
		t.Fatalf("expected MCP tools to be available")
	}

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "llm_wiki_create_namespace",
		Arguments: map[string]any{
			"key":          "projects",
			"display_name": "Projects",
			"description":  "shared project docs",
			"visibility":   "tenant",
		},
	})
	if err != nil {
		t.Fatalf("create namespace tool: %v", err)
	}
	if result.IsError {
		t.Fatalf("create namespace tool returned error result")
	}

	listed, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "llm_wiki_list_namespaces",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("list namespaces tool: %v", err)
	}
	if len(listed.Content) == 0 {
		t.Fatalf("expected namespace tool content")
	}

	resource, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "llm-wiki://namespaces"})
	if err != nil {
		t.Fatalf("read resource: %v", err)
	}
	if len(resource.Contents) != 1 || !strings.Contains(resource.Contents[0].Text, "Projects") {
		t.Fatalf("unexpected resource contents: %#v", resource.Contents)
	}
}

func TestMCPSSE(t *testing.T) {
	t.Parallel()

	serverURL := newMCPTestServer(t)
	session := connectMCPClient(t, &mcp.SSEClientTransport{
		Endpoint:   serverURL + "/sse",
		HTTPClient: tenantHTTPClient("tenant-mcp-sse"),
	})
	defer session.Close()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "llm_wiki_list_spaces",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("list spaces over SSE: %v", err)
	}
	if result.IsError {
		t.Fatalf("list spaces over SSE returned error")
	}
}

func newMCPTestServer(t *testing.T) string {
	t.Helper()

	pg := testutil.StartPostgres(t)
	cfg := testutil.MustBaseConfig(pg.Config)

	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.Postgres)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := db.ApplyMigrations(ctx, pool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	logger, err := logging.New(cfg.Environment, cfg.LogLevel)
	if err != nil {
		t.Fatalf("logger: %v", err)
	}
	t.Cleanup(func() { _ = logger.Sync() })

	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, service.New(repository.New(pool))))
	t.Cleanup(server.Close)
	return server.URL
}

func connectMCPClient(t *testing.T, transport mcp.Transport) *mcp.ClientSession {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "llm-wiki-e2e-client",
		Version: "0.0.1",
	}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	return session
}

func tenantHTTPClient(tenantID string) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &tenantRoundTripper{
			base:     http.DefaultTransport,
			tenantID: tenantID,
		},
	}
}

type tenantRoundTripper struct {
	base     http.RoundTripper
	tenantID string
}

func (r *tenantRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = cloned.Header.Clone()
	cloned.Header.Set("X-LLM-Wiki-Tenant-ID", r.tenantID)
	return r.base.RoundTrip(cloned)
}

func readBody(t *testing.T, body io.ReadCloser) string {
	t.Helper()
	defer body.Close()
	payload, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(payload)
}
