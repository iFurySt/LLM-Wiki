package e2e

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ifuryst/llm-wiki/internal/db"
	"github.com/ifuryst/llm-wiki/internal/httpserver"
	"github.com/ifuryst/llm-wiki/internal/logging"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/testutil"
)

func TestSetupAndAdminUserManagement(t *testing.T) {
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
	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer server.Close()

	resp, err := http.Get(server.URL + "/setup")
	if err != nil {
		t.Fatalf("get setup: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 setup page, got %d", resp.StatusCode)
	}

	form := url.Values{
		"ns":           {"default"},
		"username":     {"admin"},
		"display_name": {"Workspace Admin"},
		"password":     {"secret123"},
	}
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	setupReq, _ := http.NewRequest(http.MethodPost, server.URL+"/setup", strings.NewReader(form.Encode()))
	setupReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setupResp, err := noRedirectClient.Do(setupReq)
	if err != nil {
		t.Fatalf("post setup: %v", err)
	}
	defer setupResp.Body.Close()
	if setupResp.StatusCode != http.StatusSeeOther {
		body, _ := io.ReadAll(setupResp.Body)
		t.Fatalf("expected redirect after setup, got %d: %s", setupResp.StatusCode, string(body))
	}

	var sessionCookie *http.Cookie
	for _, item := range setupResp.Cookies() {
		if item.Name == "llm_wiki_session" {
			sessionCookie = item
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected web session cookie after setup")
	}

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/admin/users", nil)
	req.AddCookie(sessionCookie)
	usersResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get admin users: %v", err)
	}
	defer usersResp.Body.Close()
	if usersResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 admin users, got %d", usersResp.StatusCode)
	}

	createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/admin/users", strings.NewReader(url.Values{
		"username":     {"editor"},
		"display_name": {"Editor User"},
		"password":     {"secret123"},
	}.Encode()))
	createReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createReq.AddCookie(sessionCookie)
	createResp, err := noRedirectClient.Do(createReq)
	if err != nil {
		t.Fatalf("create user from admin ui: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect after create user, got %d", createResp.StatusCode)
	}
}
