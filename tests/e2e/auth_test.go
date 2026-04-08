package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestDeviceLoginAndWhoAmI(t *testing.T) {
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
		TenantID:    "tenant-device",
		Username:    "admin",
		DisplayName: "Device User",
		Password:    "secret123",
	}); err != nil {
		t.Fatalf("initialize device tenant: %v", err)
	}
	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer server.Close()

	client := httpclient.New(server.URL, 10*time.Second, "")
	startResp, err := client.StartDeviceLogin(ctx, api.StartDeviceLoginRequest{
		TenantID:    "tenant-device",
		DisplayName: "Device User",
	})
	if err != nil {
		t.Fatalf("start device login: %v", err)
	}

	form := url.Values{}
	form.Set("user_code", startResp.UserCode)
	form.Set("username", "admin")
	form.Set("password", "secret123")
	resp, err := http.Post(server.URL+"/auth/device", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("approve device login: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("approve device login status: %d", resp.StatusCode)
	}

	tokenResp, err := client.ExchangeToken(ctx, api.TokenExchangeRequest{
		GrantType:  "urn:ietf:params:oauth:grant-type:device_code",
		DeviceCode: startResp.DeviceCode,
	})
	if err != nil {
		t.Fatalf("exchange device code: %v", err)
	}
	client.SetAccessToken(tokenResp.AccessToken)

	whoami, err := client.WhoAmI(ctx)
	if err != nil {
		t.Fatalf("whoami: %v", err)
	}
	if whoami.TenantID != "tenant-device" {
		t.Fatalf("unexpected tenant: %q", whoami.TenantID)
	}
	if whoami.PrincipalType != "user" {
		t.Fatalf("unexpected principal type: %q", whoami.PrincipalType)
	}
}

func TestServicePrincipalAndTokenIssue(t *testing.T) {
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
	adminToken := "tenant-admin-token"
	if err := bootstrapTestToken(ctx, svc, "tenant-auth", adminToken); err != nil {
		t.Fatalf("bootstrap token: %v", err)
	}

	server := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer server.Close()

	adminClient := httpclient.New(server.URL, 10*time.Second, adminToken)
	principal, err := adminClient.CreateServicePrincipal(ctx, api.CreateServicePrincipalRequest{
		DisplayName: "ci-runner",
	})
	if err != nil {
		t.Fatalf("create service principal: %v", err)
	}

	tokenResp, err := adminClient.IssueToken(ctx, api.IssueTokenRequest{
		PrincipalID:      principal.ID,
		DisplayName:      "ci-runner-token",
		Scopes:           []string{"documents.read", "documents.write"},
		ExpiresInSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if tokenResp.PlaintextToken == "" {
		t.Fatalf("expected plaintext token to be returned on issue")
	}

	serviceClient := httpclient.New(server.URL, 10*time.Second, tokenResp.PlaintextToken)
	whoami, err := serviceClient.WhoAmI(ctx)
	if err != nil {
		t.Fatalf("service whoami: %v", err)
	}
	if whoami.PrincipalType != "service" && whoami.PrincipalType != "admin" {
		t.Fatalf("unexpected principal type: %q", whoami.PrincipalType)
	}
}
