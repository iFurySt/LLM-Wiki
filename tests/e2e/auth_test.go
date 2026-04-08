package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
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

func TestBrowserOAuthAutoProvisionAndWhoAmI(t *testing.T) {
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
	if err := bootstrapTestToken(ctx, svc, "platform-admin", "platform-admin-token"); err != nil {
		t.Fatalf("bootstrap token: %v", err)
	}

	var appServer *httptest.Server
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/authorize":
			redirectURI := r.URL.Query().Get("redirect_uri")
			state := r.URL.Query().Get("state")
			http.Redirect(w, r, redirectURI+"&code=provider-code&state="+url.QueryEscape(state), http.StatusFound)
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "provider-access-token",
				"token_type":   "Bearer",
			})
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sub":                "oauth-user-1",
				"email":              "octocat@example.com",
				"name":               "The Octocat",
				"preferred_username": "octocat",
				"email_verified":     true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer providerServer.Close()

	if _, err := svc.UpsertOAuthProvider(ctx, api.UpsertOAuthProviderRequest{
		Name:              "github",
		DisplayName:       "GitHub",
		AuthURL:           providerServer.URL + "/authorize",
		TokenURL:          providerServer.URL + "/token",
		UserinfoURL:       providerServer.URL + "/userinfo",
		ClientID:          "client-id",
		ClientSecret:      "client-secret",
		Scopes:            []string{"openid", "email", "profile"},
		Enabled:           true,
		AutoCreateUsers:   true,
		AutoCreateTenants: true,
	}); err != nil {
		t.Fatalf("upsert oauth provider: %v", err)
	}

	appServer = httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer appServer.Close()

	callbackServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer callbackServer.Close()

	client := httpclient.New(appServer.URL, 10*time.Second, "")
	sum := sha256.Sum256([]byte("verifier"))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	startResp, err := client.StartBrowserLogin(ctx, api.StartBrowserLoginRequest{
		Provider:            "github",
		State:               "cli-state",
		RedirectURI:         callbackServer.URL + "/auth/callback",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	})
	if err != nil {
		t.Fatalf("start browser login: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appServer.URL+"/auth/oauth/callback?request_id="+url.QueryEscape(startResp.RequestID)+"&provider=github&code=provider-code", nil)
	if err != nil {
		t.Fatalf("build callback request: %v", err)
	}
	noRedirect := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		t.Fatalf("oauth callback request: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("unexpected oauth callback status: %d body=%s", resp.StatusCode, string(body))
	}
	location, err := resp.Location()
	if err != nil {
		t.Fatalf("oauth callback location: %v", err)
	}
	callback := location.Query()
	if callback.Get("state") != "cli-state" {
		t.Fatalf("unexpected callback state: %q", callback.Get("state"))
	}

	tokenResp, err := client.ExchangeToken(ctx, api.TokenExchangeRequest{
		GrantType:    "authorization_code",
		Code:         callback.Get("code"),
		CodeVerifier: "verifier",
	})
	if err != nil {
		t.Fatalf("exchange authorization code: %v", err)
	}

	client.SetAccessToken(tokenResp.AccessToken)
	whoami, err := client.WhoAmI(ctx)
	if err != nil {
		t.Fatalf("whoami: %v", err)
	}
	if whoami.TenantID != "octocat" {
		t.Fatalf("unexpected tenant id: %q", whoami.TenantID)
	}
	if whoami.DisplayName != "The Octocat" {
		t.Fatalf("unexpected display name: %q", whoami.DisplayName)
	}
	if whoami.PrincipalID == "" {
		t.Fatal("expected principal id to be set")
	}
}

func TestWorkspaceCreateInviteAcceptAndSwitch(t *testing.T) {
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
	providerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			_ = r.ParseForm()
			code := r.Form.Get("code")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": code,
				"token_type":   "Bearer",
			})
		case "/userinfo":
			switch strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")) {
			case "code-user1":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sub":                "oauth-user-1",
					"email":              "owner@example.com",
					"name":               "Owner User",
					"preferred_username": "owner",
					"email_verified":     true,
				})
			case "code-user2":
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sub":                "oauth-user-2",
					"email":              "member@example.com",
					"name":               "Member User",
					"preferred_username": "member",
					"email_verified":     true,
				})
			default:
				http.Error(w, "unknown token", http.StatusUnauthorized)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer providerServer.Close()

	if _, err := svc.UpsertOAuthProvider(ctx, api.UpsertOAuthProviderRequest{
		Name:              "github",
		DisplayName:       "GitHub",
		AuthURL:           providerServer.URL + "/authorize",
		TokenURL:          providerServer.URL + "/token",
		UserinfoURL:       providerServer.URL + "/userinfo",
		ClientID:          "client-id",
		ClientSecret:      "client-secret",
		Scopes:            []string{"openid", "email", "profile"},
		Enabled:           true,
		AutoCreateUsers:   true,
		AutoCreateTenants: true,
	}); err != nil {
		t.Fatalf("upsert oauth provider: %v", err)
	}

	appServer := httptest.NewServer(httpserver.NewHandler(cfg, logger, svc))
	defer appServer.Close()

	ownerClient := httpclient.New(appServer.URL, 10*time.Second, "")
	ownerToken := oauthBrowserLoginForTest(t, ctx, ownerClient, appServer.URL, "github", "code-user1")
	ownerClient.SetAccessToken(ownerToken.AccessToken)

	workspace, err := ownerClient.CreateWorkspace(ctx, api.CreateWorkspaceRequest{
		TenantID:    "team-space",
		DisplayName: "Team Space",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if workspace.TenantID != "team-space" {
		t.Fatalf("unexpected workspace tenant id: %q", workspace.TenantID)
	}
	switchedOwner, err := ownerClient.SwitchTenant(ctx, "team-space")
	if err != nil {
		t.Fatalf("owner switch tenant: %v", err)
	}
	ownerClient.SetAccessToken(switchedOwner.AccessToken)

	invite, err := ownerClient.CreateInvite(ctx, api.CreateInviteRequest{
		Email:          "member@example.com",
		Role:           "member",
		ExpiresInHours: 24,
	})
	if err != nil {
		t.Fatalf("create invite: %v", err)
	}
	if invite.Token == "" {
		t.Fatal("expected invite token")
	}

	memberClient := httpclient.New(appServer.URL, 10*time.Second, "")
	memberToken := oauthBrowserLoginForTest(t, ctx, memberClient, appServer.URL, "github", "code-user2")
	memberClient.SetAccessToken(memberToken.AccessToken)

	if _, err := memberClient.AcceptInvite(ctx, invite.Token); err != nil {
		t.Fatalf("accept invite: %v", err)
	}
	workspaces, err := memberClient.ListWorkspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces.Items) < 2 {
		t.Fatalf("expected at least 2 workspaces, got %d", len(workspaces.Items))
	}

	switched, err := memberClient.SwitchTenant(ctx, "team-space")
	if err != nil {
		t.Fatalf("switch tenant: %v", err)
	}
	memberClient.SetAccessToken(switched.AccessToken)
	whoami, err := memberClient.WhoAmI(ctx)
	if err != nil {
		t.Fatalf("whoami after switch: %v", err)
	}
	if whoami.TenantID != "team-space" {
		t.Fatalf("unexpected switched tenant: %q", whoami.TenantID)
	}
}

func oauthBrowserLoginForTest(t *testing.T, ctx context.Context, client *httpclient.Client, appServerURL string, provider string, providerCode string) api.TokenExchangeResponse {
	t.Helper()
	sum := sha256.Sum256([]byte("verifier"))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	startResp, err := client.StartBrowserLogin(ctx, api.StartBrowserLoginRequest{
		Provider:            provider,
		State:               "cli-state",
		RedirectURI:         "http://127.0.0.1/callback",
		CodeChallenge:       challenge,
		CodeChallengeMethod: "S256",
	})
	if err != nil {
		t.Fatalf("start browser login: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appServerURL+"/auth/oauth/callback?request_id="+url.QueryEscape(startResp.RequestID)+"&provider="+url.QueryEscape(provider)+"&code="+url.QueryEscape(providerCode), nil)
	if err != nil {
		t.Fatalf("build callback request: %v", err)
	}
	noRedirect := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		t.Fatalf("oauth callback request: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("unexpected oauth callback status: %d body=%s", resp.StatusCode, string(body))
	}
	location, err := resp.Location()
	if err != nil {
		t.Fatalf("oauth callback location: %v", err)
	}
	tokenResp, err := client.ExchangeToken(ctx, api.TokenExchangeRequest{
		GrantType:    "authorization_code",
		Code:         location.Query().Get("code"),
		CodeVerifier: "verifier",
	})
	if err != nil {
		t.Fatalf("exchange authorization code: %v", err)
	}
	return tokenResp
}
