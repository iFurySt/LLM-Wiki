package httpserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/config"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
)

const webSessionCookie = "llm_wiki_session"
const webSessionCookieTTL = 24 * time.Hour

type authPageData struct {
	Title        string
	Heading      string
	Description  string
	Action       string
	TenantID     string
	RequestID    string
	UserCode     string
	ErrorMessage string
	ProviderHTML string
}

type adminPageData struct {
	TenantID       string
	DisplayName    string
	Users          []api.UserResponse
	Providers      []api.OAuthProviderResponse
	ErrorMessage   string
	SuccessMessage string
}

func registerAuthBrowserRoutes(engine *gin.Engine, svc *service.Service, cfg config.Config) {
	engine.GET("/setup", func(c *gin.Context) {
		status, err := svc.SetupStatus(c.Request.Context(), cfg.Auth.BootstrapTenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		if status.Initialized {
			c.Redirect(http.StatusSeeOther, "/admin/login")
			return
		}
		renderAuthShell(c, authPageData{
			Title:       "Initialize LLM-Wiki",
			Heading:     "Initialize LLM-Wiki",
			Description: "Create the first admin account and default tenant.",
			Action:      "/setup",
			TenantID:    status.DefaultTenant,
		}, setupFormHTML)
	})
	engine.POST("/setup", func(c *gin.Context) {
		user, err := svc.Initialize(c.Request.Context(), api.InitializeRequest{
			TenantID:    c.PostForm("tenant_id"),
			Username:    c.PostForm("username"),
			DisplayName: c.PostForm("display_name"),
			Password:    c.PostForm("password"),
		})
		if err != nil {
			renderAuthShell(c, authPageData{
				Title:        "Initialize LLM-Wiki",
				Heading:      "Initialize LLM-Wiki",
				Description:  "Create the first admin account and default tenant.",
				Action:       "/setup",
				TenantID:     c.PostForm("tenant_id"),
				ErrorMessage: err.Error(),
			}, setupFormHTML)
			return
		}
		session, err := svc.CreateWebSession(c.Request.Context(), user.PrincipalID, user.TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		setWebSessionCookie(c, session.ID)
		c.Redirect(http.StatusSeeOther, "/admin/users")
	})

	engine.GET("/admin/login", func(c *gin.Context) {
		status, err := svc.SetupStatus(c.Request.Context(), cfg.Auth.BootstrapTenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		if !status.Initialized {
			c.Redirect(http.StatusSeeOther, "/setup")
			return
		}
		renderAuthShell(c, authPageData{
			Title:       "Admin Login",
			Heading:     "Admin Login",
			Description: "Sign in with a tenant, username, and password.",
			Action:      "/admin/login",
			TenantID:    status.DefaultTenant,
		}, loginFormHTML)
	})
	engine.POST("/admin/login", func(c *gin.Context) {
		user, _, err := svc.LoginUser(c.Request.Context(), c.PostForm("tenant_id"), c.PostForm("username"), c.PostForm("password"))
		if err != nil {
			renderAuthShell(c, authPageData{
				Title:        "Admin Login",
				Heading:      "Admin Login",
				Description:  "Sign in with a tenant, username, and password.",
				Action:       "/admin/login",
				TenantID:     c.PostForm("tenant_id"),
				ErrorMessage: "Invalid tenant, username, or password.",
			}, loginFormHTML)
			return
		}
		session, err := svc.CreateWebSession(c.Request.Context(), user.PrincipalID, user.TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		setWebSessionCookie(c, session.ID)
		c.Redirect(http.StatusSeeOther, "/admin/users")
	})
	engine.POST("/admin/logout", func(c *gin.Context) {
		if sessionID, err := c.Cookie(webSessionCookie); err == nil {
			_ = svc.DeleteWebSession(c.Request.Context(), sessionID)
		}
		clearWebSessionCookie(c)
		c.Redirect(http.StatusSeeOther, "/admin/login")
	})

	engine.GET("/admin/users", requireAdminWebSession(svc), func(c *gin.Context) {
		principal, _ := auth.PrincipalFromContext(c.Request.Context())
		users, err := svc.ListUsers(c.Request.Context(), principal.TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		providers, err := svc.ListAdminOAuthProviders(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		renderAdminUsersPage(c, adminPageData{
			TenantID:    principal.TenantID,
			DisplayName: principal.DisplayName,
			Users:       users.Items,
			Providers:   providers.Items,
		})
	})
	engine.POST("/admin/users", requireAdminWebSession(svc), func(c *gin.Context) {
		principal, _ := auth.PrincipalFromContext(c.Request.Context())
		_, err := svc.CreateUser(c.Request.Context(), principal.TenantID, api.CreateUserRequest{
			Username:    c.PostForm("username"),
			DisplayName: c.PostForm("display_name"),
			Password:    c.PostForm("password"),
			IsAdmin:     c.PostForm("is_admin") == "on",
		})
		if err != nil {
			users, _ := svc.ListUsers(c.Request.Context(), principal.TenantID)
			providers, _ := svc.ListAdminOAuthProviders(c.Request.Context())
			renderAdminUsersPage(c, adminPageData{
				TenantID:     principal.TenantID,
				DisplayName:  principal.DisplayName,
				Users:        users.Items,
				Providers:    providers.Items,
				ErrorMessage: err.Error(),
			})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/users?success=user+created")
	})
	engine.POST("/admin/auth-providers", requireAdminWebSession(svc), func(c *gin.Context) {
		enabled := c.PostForm("enabled") == "on"
		autoCreateUsers := c.PostForm("auto_create_users") == "on"
		autoCreateTenants := c.PostForm("auto_create_tenants") == "on"
		_, err := svc.UpsertOAuthProvider(c.Request.Context(), api.UpsertOAuthProviderRequest{
			Name:              c.PostForm("name"),
			DisplayName:       c.PostForm("display_name"),
			AuthURL:           c.PostForm("auth_url"),
			TokenURL:          c.PostForm("token_url"),
			UserinfoURL:       c.PostForm("userinfo_url"),
			ClientID:          c.PostForm("client_id"),
			ClientSecret:      c.PostForm("client_secret"),
			Scopes:            strings.Fields(c.PostForm("scopes")),
			Enabled:           enabled,
			AutoCreateUsers:   autoCreateUsers,
			AutoCreateTenants: autoCreateTenants,
		})
		if err != nil {
			principal, _ := auth.PrincipalFromContext(c.Request.Context())
			users, _ := svc.ListUsers(c.Request.Context(), principal.TenantID)
			providers, _ := svc.ListAdminOAuthProviders(c.Request.Context())
			renderAdminUsersPage(c, adminPageData{
				TenantID:     principal.TenantID,
				DisplayName:  principal.DisplayName,
				Users:        users.Items,
				Providers:    providers.Items,
				ErrorMessage: err.Error(),
			})
			return
		}
		c.Redirect(http.StatusSeeOther, "/admin/users?success=provider+saved")
	})

	engine.GET("/auth/authorize", func(c *gin.Context) {
		requestID := c.Query("request_id")
		redirectURL, request, providers, err := svc.ResolveOAuthAuthorizeURL(c.Request.Context(), publicBaseURL(c), requestID, c.Query("provider"))
		if err != nil {
			handleError(c, err)
			return
		}
		if redirectURL != "" {
			c.Redirect(http.StatusSeeOther, redirectURL)
			return
		}
		renderAuthShell(c, authPageData{
			Title:        "Approve CLI Login",
			Heading:      "Approve CLI Login",
			Description:  "Choose a login provider or fall back to local username and password approval.",
			Action:       "/auth/authorize",
			TenantID:     request.TenantID,
			RequestID:    request.ID,
			ProviderHTML: renderOAuthProviderButtons(request.ID, providers),
		}, oauthApproveFormHTML)
	})
	engine.POST("/auth/authorize", func(c *gin.Context) {
		request, err := svc.ApproveAuthRequestWithPassword(c.Request.Context(), c.PostForm("request_id"), c.PostForm("username"), c.PostForm("password"))
		if err != nil {
			renderAuthShell(c, authPageData{
				Title:        "Approve CLI Login",
				Heading:      "Approve CLI Login",
				Description:  "Choose a login provider or fall back to local username and password approval.",
				Action:       "/auth/authorize",
				TenantID:     c.PostForm("tenant_id"),
				RequestID:    c.PostForm("request_id"),
				ErrorMessage: "Login failed.",
			}, oauthApproveFormHTML)
			return
		}
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", request.RedirectURI, request.AuthCode, request.State)
		c.Redirect(http.StatusSeeOther, redirectURL)
	})
	engine.GET("/auth/oauth/callback", func(c *gin.Context) {
		requestID := c.Query("request_id")
		request, err := svc.Repo().GetAuthRequest(c.Request.Context(), requestID)
		if err != nil {
			handleError(c, err)
			return
		}
		providerName := strings.TrimSpace(c.Query("provider"))
		if providerName == "" {
			providerName = request.OAuthProvider
		}
		if providerName == "" {
			providerName = c.Query("state")
		}
		_, _, providers, _ := svc.ResolveOAuthAuthorizeURL(c.Request.Context(), publicBaseURL(c), requestID, "")
		approved, err := svc.CompleteOAuthBrowserLogin(c.Request.Context(), requestID, providerName, c.Query("code"), publicBaseURL(c))
		if err != nil {
			renderAuthShell(c, authPageData{
				Title:        "Approve CLI Login",
				Heading:      "Approve CLI Login",
				Description:  "OAuth login failed.",
				Action:       "/auth/authorize",
				TenantID:     request.TenantID,
				RequestID:    request.ID,
				ErrorMessage: err.Error(),
				ProviderHTML: renderOAuthProviderButtons(request.ID, providers),
			}, oauthApproveFormHTML)
			return
		}
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", approved.RedirectURI, approved.AuthCode, approved.State)
		c.Redirect(http.StatusSeeOther, redirectURL)
	})

	engine.GET("/auth/device", func(c *gin.Context) {
		renderAuthShell(c, authPageData{
			Title:       "Device Sign In",
			Heading:     "Device Sign In",
			Description: "Enter your one-time code, username, and password.",
			Action:      "/auth/device",
			UserCode:    strings.ToUpper(c.Query("code")),
		}, deviceFormHTML)
	})
	engine.POST("/auth/device", func(c *gin.Context) {
		_, err := svc.ApproveDeviceCodeWithPassword(c.Request.Context(), c.PostForm("user_code"), c.PostForm("username"), c.PostForm("password"))
		if err != nil {
			renderAuthShell(c, authPageData{
				Title:        "Device Sign In",
				Heading:      "Device Sign In",
				Description:  "Enter your one-time code, username, and password.",
				Action:       "/auth/device",
				UserCode:     c.PostForm("user_code"),
				ErrorMessage: "Login failed or code expired.",
			}, deviceFormHTML)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(deviceSuccessHTML()))
	})
}

func requireAdminWebSession(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(webSessionCookie)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/admin/login")
			c.Abort()
			return
		}
		_, user, principal, err := svc.GetWebSession(c.Request.Context(), sessionID)
		if err != nil || !user.IsAdmin {
			clearWebSessionCookie(c)
			c.Redirect(http.StatusSeeOther, "/admin/login")
			c.Abort()
			return
		}
		ctx := auth.WithPrincipal(c.Request.Context(), principal)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func setWebSessionCookie(c *gin.Context, sessionID string) {
	c.SetCookie(webSessionCookie, sessionID, int(webSessionCookieTTL.Seconds()), "/", "", false, true)
}

func clearWebSessionCookie(c *gin.Context) {
	c.SetCookie(webSessionCookie, "", -1, "/", "", false, true)
}

func renderAuthShell(c *gin.Context, data authPageData, content string) {
	html := fmt.Sprintf(authShellHTML, template.HTMLEscapeString(data.Title), template.HTMLEscapeString(data.Heading), template.HTMLEscapeString(data.Description), content)
	replacer := strings.NewReplacer(
		"{{action}}", template.HTMLEscapeString(data.Action),
		"{{tenant_id}}", template.HTMLEscapeString(data.TenantID),
		"{{request_id}}", template.HTMLEscapeString(data.RequestID),
		"{{user_code}}", template.HTMLEscapeString(data.UserCode),
		"{{error}}", template.HTMLEscapeString(data.ErrorMessage),
		"{{provider_html}}", data.ProviderHTML,
	)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(replacer.Replace(html)))
}

func renderAdminUsersPage(c *gin.Context, data adminPageData) {
	rows := make([]string, 0, len(data.Users))
	for _, item := range data.Users {
		role := "member"
		if item.IsAdmin {
			role = "admin"
		}
		rows = append(rows, fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			template.HTMLEscapeString(item.Username),
			template.HTMLEscapeString(item.DisplayName),
			template.HTMLEscapeString(role),
			template.HTMLEscapeString(item.CreatedAt),
		))
	}
	providerRows := make([]string, 0, len(data.Providers))
	for _, item := range data.Providers {
		providerRows = append(providerRows, fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%t</td><td>%t</td><td>%t</td></tr>`,
			template.HTMLEscapeString(item.Name),
			template.HTMLEscapeString(item.DisplayName),
			item.Enabled,
			item.AutoCreateUsers,
			item.AutoCreateTenants,
		))
	}
	html := fmt.Sprintf(adminUsersHTML,
		template.HTMLEscapeString(data.TenantID),
		template.HTMLEscapeString(data.DisplayName),
		template.HTMLEscapeString(data.ErrorMessage),
		template.HTMLEscapeString(c.Query("success")),
		strings.Join(rows, ""),
		strings.Join(providerRows, ""),
	)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func renderOAuthProviderButtons(requestID string, providers []repository.OAuthProviderRecord) string {
	if len(providers) == 0 {
		return ""
	}
	parts := []string{`<div style="margin:0 0 14px;">`}
	for _, item := range providers {
		parts = append(parts, fmt.Sprintf(`<a href="/auth/authorize?request_id=%s&provider=%s" style="display:block;text-decoration:none;margin:0 0 10px;"><button type="button">%s</button></a>`,
			template.URLQueryEscaper(requestID),
			template.URLQueryEscaper(item.Name),
			template.HTMLEscapeString("Continue with "+item.DisplayName),
		))
	}
	parts = append(parts, `</div>`)
	return strings.Join(parts, "")
}

const authShellHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
  <style>
    :root { --bg:#14110d; --panel:#1e1913; --line:#f2c94c; --text:#f7f1e4; --muted:#c0b391; --danger:#ff8d7b; --pixel:"Courier New",monospace; --body:"Trebuchet MS",sans-serif; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; color:var(--text); font-family:var(--body); background:linear-gradient(rgba(242,201,76,.08) 1px, transparent 1px),linear-gradient(90deg, rgba(242,201,76,.08) 1px, transparent 1px),radial-gradient(circle at top, rgba(242,201,76,.18), transparent 35%%), var(--bg); background-size:24px 24px,24px 24px,100%% 100%%,100%% 100%%; display:grid; place-items:center; padding:20px; }
    .card { width:min(560px,100%%); border:4px solid #050402; background:linear-gradient(180deg, rgba(255,255,255,.03), rgba(0,0,0,.22)), var(--panel); box-shadow:0 0 0 3px rgba(242,201,76,.15), 0 18px 40px rgba(0,0,0,.45); padding:22px; }
    .mark { width:52px; height:52px; display:grid; place-items:center; background:var(--line); color:#1e1710; border:3px solid #050402; font:700 12px/1 var(--pixel); box-shadow:4px 4px 0 rgba(0,0,0,.3); }
    h1 { margin:14px 0 8px; font:700 24px/1.1 var(--pixel); text-transform:uppercase; }
    p { margin:0 0 16px; color:var(--muted); line-height:1.6; }
    label { display:block; margin:0 0 6px; font:700 12px/1 var(--pixel); text-transform:uppercase; }
    input { width:100%%; padding:12px 14px; border:3px solid #050402; background:#fff8ea; color:#231b11; margin:0 0 14px; font:700 14px/1 var(--pixel); }
    button { width:100%%; border:3px solid #050402; background:var(--line); color:#1c150d; padding:12px 14px; font:700 13px/1 var(--pixel); text-transform:uppercase; cursor:pointer; box-shadow:4px 4px 0 rgba(0,0,0,.24); }
    .error { min-height:20px; color:var(--danger); font:700 12px/1.4 var(--pixel); margin-bottom:10px; }
    .grid { display:grid; gap:10px; grid-template-columns:1fr 1fr; }
    @media (max-width:640px){ .grid { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <div class="card">
    <div class="mark">LW</div>
    <h1>%s</h1>
    <p>%s</p>
    %s
  </div>
</body>
</html>`

const setupFormHTML = `<form method="post" action="{{action}}">
  <div class="error">{{error}}</div>
  <label>Tenant</label>
  <input name="tenant_id" value="{{tenant_id}}" placeholder="default" required>
  <label>Username</label>
  <input name="username" placeholder="admin" required>
  <label>Display Name</label>
  <input name="display_name" placeholder="Workspace Admin" required>
  <label>Password</label>
  <input type="password" name="password" placeholder="Choose a strong password" required>
  <button type="submit">Initialize Workspace</button>
</form>`

const loginFormHTML = `<form method="post" action="{{action}}">
  <div class="error">{{error}}</div>
  <label>Tenant</label>
  <input name="tenant_id" value="{{tenant_id}}" placeholder="default" required>
  <label>Username</label>
  <input name="username" placeholder="admin" required>
  <label>Password</label>
  <input type="password" name="password" placeholder="Password" required>
  <button type="submit">Sign In</button>
</form>`

const oauthApproveFormHTML = `{{provider_html}}<form method="post" action="{{action}}">
  <div class="error">{{error}}</div>
  <input type="hidden" name="request_id" value="{{request_id}}">
  <label>Tenant</label>
  <input name="tenant_id" value="{{tenant_id}}" readonly>
  <label>Username</label>
  <input name="username" placeholder="admin" required>
  <label>Password</label>
  <input type="password" name="password" placeholder="Password" required>
  <button type="submit">Approve Login</button>
</form>`

const deviceFormHTML = `<form method="post" action="{{action}}">
  <div class="error">{{error}}</div>
  <label>Device Code</label>
  <input name="user_code" value="{{user_code}}" placeholder="13E37E23" required>
  <label>Username</label>
  <input name="username" placeholder="admin" required>
  <label>Password</label>
  <input type="password" name="password" placeholder="Password" required>
  <button type="submit">Approve Device Login</button>
</form>`

const adminUsersHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LLM-Wiki Admin</title>
  <style>
    :root { --bg:#13110d; --panel:#1f1811; --panel2:#17120d; --line:#f2c94c; --text:#f7f1e4; --muted:#c0b391; --danger:#ff8d7b; --success:#8ddf8b; --pixel:"Courier New",monospace; --body:"Trebuchet MS",sans-serif; }
    * { box-sizing:border-box; } body { margin:0; color:var(--text); font-family:var(--body); background:linear-gradient(rgba(242,201,76,.07) 1px, transparent 1px),linear-gradient(90deg, rgba(242,201,76,.07) 1px, transparent 1px),var(--bg); background-size:24px 24px,24px 24px,100%% 100%%; }
    .page { max-width:1200px; margin:0 auto; padding:20px; } .frame { border:4px solid #050402; background:var(--panel); box-shadow:0 0 0 3px rgba(242,201,76,.14), 0 18px 40px rgba(0,0,0,.45); }
    .top { display:flex; justify-content:space-between; gap:16px; align-items:center; padding:18px; } h1 { margin:0; font:700 24px/1 var(--pixel); text-transform:uppercase; } .muted { color:var(--muted); }
    .shell { display:grid; grid-template-columns:340px 1fr; gap:16px; padding:0 18px 18px; } .panel { border:3px solid #050402; background:var(--panel2); padding:16px; } .stack { display:grid; gap:16px; }
    label { display:block; margin:0 0 6px; font:700 12px/1 var(--pixel); text-transform:uppercase; } input { width:100%%; padding:12px 14px; border:3px solid #050402; background:#fff8ea; color:#231b11; margin:0 0 14px; font:700 14px/1 var(--pixel); }
    button { border:3px solid #050402; background:var(--line); color:#20170d; padding:12px 14px; font:700 13px/1 var(--pixel); text-transform:uppercase; cursor:pointer; box-shadow:4px 4px 0 rgba(0,0,0,.24); }
    table { width:100%%; border-collapse:collapse; } th, td { border:2px solid #050402; padding:10px 12px; text-align:left; } th { font:700 12px/1 var(--pixel); text-transform:uppercase; color:var(--muted); }
    .error { color:var(--danger); font:700 12px/1.4 var(--pixel); margin-bottom:10px; } .success { color:var(--success); font:700 12px/1.4 var(--pixel); margin-bottom:10px; }
    @media (max-width:900px){ .shell { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <div class="page">
    <div class="frame">
      <div class="top">
        <div>
          <h1>LLM-Wiki Admin</h1>
          <div class="muted">Tenant: %s · Signed in as %s</div>
        </div>
        <form method="post" action="/admin/logout"><button type="submit">Log Out</button></form>
      </div>
      <div class="shell">
        <div class="stack">
        <div class="panel">
          <div class="error">%s</div>
          <div class="success">%s</div>
          <form method="post" action="/admin/users">
            <label>Username</label>
            <input name="username" placeholder="new-user" required>
            <label>Display Name</label>
            <input name="display_name" placeholder="New User" required>
            <label>Password</label>
            <input type="password" name="password" placeholder="Temporary password" required>
            <label><input type="checkbox" name="is_admin" style="width:auto; margin-right:8px;"> Admin</label>
            <button type="submit">Create User</button>
          </form>
        </div>
        <div class="panel">
          <form method="post" action="/admin/auth-providers">
            <label>Provider Key</label>
            <input name="name" placeholder="google" required>
            <label>Display Name</label>
            <input name="display_name" placeholder="Google" required>
            <label>Authorization URL</label>
            <input name="auth_url" placeholder="https://accounts.google.com/o/oauth2/v2/auth" required>
            <label>Token URL</label>
            <input name="token_url" placeholder="https://oauth2.googleapis.com/token" required>
            <label>Userinfo URL</label>
            <input name="userinfo_url" placeholder="https://openidconnect.googleapis.com/v1/userinfo" required>
            <label>Client ID</label>
            <input name="client_id" placeholder="oauth client id" required>
            <label>Client Secret</label>
            <input name="client_secret" placeholder="oauth client secret" required>
            <label>Scopes</label>
            <input name="scopes" placeholder="openid email profile">
            <label><input type="checkbox" name="enabled" style="width:auto; margin-right:8px;" checked> Enabled</label>
            <label><input type="checkbox" name="auto_create_users" style="width:auto; margin-right:8px;" checked> Auto Create Users</label>
            <label><input type="checkbox" name="auto_create_tenants" style="width:auto; margin-right:8px;" checked> Auto Create Tenants</label>
            <button type="submit">Save OAuth Provider</button>
          </form>
        </div>
        </div>
        <div class="panel">
          <table>
            <thead><tr><th>Username</th><th>Display Name</th><th>Role</th><th>Created</th></tr></thead>
            <tbody>%s</tbody>
          </table>
          <br>
          <table>
            <thead><tr><th>Provider</th><th>Display Name</th><th>Enabled</th><th>Auto Users</th><th>Auto Tenants</th></tr></thead>
            <tbody>%s</tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</body>
</html>`

func deviceSuccessHTML() string {
	return `<!doctype html>
<html>
<head><meta charset="utf-8"><title>LLM-Wiki Device Sign In</title></head>
<body style="font-family:monospace;max-width:720px;margin:40px auto;padding:0 16px;background:#14110d;color:#f7f1e4;">
  <h1 style="text-transform:uppercase;">Authorization Complete</h1>
  <p>You can return to the CLI.</p>
</body>
</html>`
}
