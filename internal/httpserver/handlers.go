package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"github.com/ifuryst/llm-wiki/internal/service"
)

func registerAPIRoutes(engine *gin.Engine, svc *service.Service) {
	engine.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := svc.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"ok": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	v1 := engine.Group("/v1")
	authGroup := v1.Group("/auth")
	authGroup.POST("/browser/start", func(c *gin.Context) {
		var req api.StartBrowserLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.StartBrowserLogin(c.Request.Context(), publicBaseURL(c), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authGroup.GET("/providers", func(c *gin.Context) {
		resp, err := svc.ListOAuthProviders(c.Request.Context(), true)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authGroup.POST("/device/start", func(c *gin.Context) {
		var req api.StartDeviceLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.StartDeviceLogin(c.Request.Context(), publicBaseURL(c), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authGroup.POST("/token", func(c *gin.Context) {
		var req api.TokenExchangeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.ExchangeToken(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	protected := v1.Group("")
	protected.Use(authenticateRequest(svc))

	protected.POST("/auth/switch-tenant", func(c *gin.Context) {
		var req api.SwitchTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.SwitchTenant(c.Request.Context(), strings.TrimSpace(req.TenantID))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	protected.GET("/auth/whoami", func(c *gin.Context) {
		resp, err := svc.WhoAmI(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	authAdmin := protected.Group("/auth")
	authAdmin.Use(requireScopes(auth.ScopeTokensIssue))
	authAdmin.GET("/providers/admin", func(c *gin.Context) {
		resp, err := svc.ListAdminOAuthProviders(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authAdmin.POST("/providers", func(c *gin.Context) {
		var req api.UpsertOAuthProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.UpsertOAuthProvider(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	authAdmin.GET("/service-principals", func(c *gin.Context) {
		resp, err := svc.ListServicePrincipals(c.Request.Context(), principalFromRequest(c).TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authAdmin.POST("/service-principals", func(c *gin.Context) {
		var req api.CreateServicePrincipalRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateServicePrincipal(c.Request.Context(), principalFromRequest(c).TenantID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	authAdmin.GET("/tokens", func(c *gin.Context) {
		resp, err := svc.ListTokens(c.Request.Context(), principalFromRequest(c).TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	authAdmin.POST("/tokens", func(c *gin.Context) {
		var req api.IssueTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.IssueServiceToken(c.Request.Context(), principalFromRequest(c).TenantID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})

	authRevoke := protected.Group("/auth")
	authRevoke.Use(requireScopes(auth.ScopeTokensRevoke))
	authRevoke.POST("/tokens/:id/revoke", func(c *gin.Context) {
		tokenID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.RevokeToken(c.Request.Context(), principalFromRequest(c).TenantID, tokenID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	spaces := protected.Group("")
	spaces.Use(requireScopes(auth.ScopeSpacesRead))
	spaces.GET("/spaces", func(c *gin.Context) {
		resp, err := svc.ListSpaces(c.Request.Context(), principalFromRequest(c).TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	spaces.GET("/workspaces", func(c *gin.Context) {
		resp, err := svc.ListWorkspaces(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	spaces.POST("/workspaces", func(c *gin.Context) {
		var req api.CreateWorkspaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateWorkspace(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	spaces.GET("/workspaces/invites", func(c *gin.Context) {
		resp, err := svc.ListInvites(c.Request.Context())
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	spaces.POST("/workspaces/invites", func(c *gin.Context) {
		var req api.CreateInviteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateInvite(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	spaces.POST("/workspaces/invites/accept", func(c *gin.Context) {
		var req api.AcceptInviteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.AcceptInvite(c.Request.Context(), req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	namespacesRead := protected.Group("")
	namespacesRead.Use(requireScopes(auth.ScopeNamespacesRead))
	namespacesRead.GET("/namespaces", func(c *gin.Context) {
		resp, err := svc.ListNamespaces(c.Request.Context(), principalFromRequest(c).TenantID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	namespacesRead.GET("/namespaces/:id", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.GetNamespace(c.Request.Context(), principalFromRequest(c).TenantID, namespaceID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	namespacesWrite := protected.Group("")
	namespacesWrite.Use(requireScopes(auth.ScopeNamespacesWrite))
	namespacesWrite.POST("/namespaces", func(c *gin.Context) {
		var req api.CreateNamespaceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateNamespace(c.Request.Context(), principalFromRequest(c).TenantID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	namespacesWrite.POST("/namespaces/:id/archive", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.ArchiveNamespace(c.Request.Context(), principalFromRequest(c).TenantID, namespaceID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	documentsRead := protected.Group("")
	documentsRead.Use(requireScopes(auth.ScopeDocumentsRead))
	documentsRead.GET("/documents", func(c *gin.Context) {
		var namespaceID *int64
		var status *string
		if raw := c.Query("namespace_id"); raw != "" {
			parsed, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				badRequest(c, err)
				return
			}
			namespaceID = &parsed
		}
		if raw := c.Query("status"); raw != "" {
			status = &raw
		}
		resp, err := svc.ListDocuments(c.Request.Context(), principalFromRequest(c).TenantID, namespaceID, status)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	documentsRead.GET("/documents/:id", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.GetDocument(c.Request.Context(), principalFromRequest(c).TenantID, documentID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
	documentsRead.GET("/document-by-slug", func(c *gin.Context) {
		namespaceID, err := strconv.ParseInt(c.Query("namespace_id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.GetDocumentBySlug(c.Request.Context(), principalFromRequest(c).TenantID, namespaceID, c.Query("slug"))
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	documentsWrite := protected.Group("")
	documentsWrite.Use(requireScopes(auth.ScopeDocumentsWrite))
	documentsWrite.POST("/documents", func(c *gin.Context) {
		var req api.CreateDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.CreateDocument(c.Request.Context(), principalFromRequest(c).TenantID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusCreated, resp)
	})
	documentsWrite.PUT("/documents/:id", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		var req api.UpdateDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.UpdateDocument(c.Request.Context(), principalFromRequest(c).TenantID, documentID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})

	documentsArchive := protected.Group("")
	documentsArchive.Use(requireScopes(auth.ScopeDocumentsArchive))
	documentsArchive.POST("/documents/:id/archive", func(c *gin.Context) {
		documentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			badRequest(c, err)
			return
		}
		var req api.ArchiveDocumentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			badRequest(c, err)
			return
		}
		resp, err := svc.ArchiveDocument(c.Request.Context(), principalFromRequest(c).TenantID, documentID, req)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, resp)
	})
}

func authenticateRequest(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerTokenFromRequest(c.Request)
		if token == "" {
			unauthorized(c)
			return
		}
		principal, err := svc.AuthenticateBearerToken(c.Request.Context(), token)
		if err != nil {
			handleError(c, err)
			return
		}
		ctx := auth.WithPrincipal(c.Request.Context(), principal)
		c.Request = c.Request.WithContext(ctx)
		c.Set("principal", principal)
		c.Next()
	}
}

func requireScopes(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal := principalFromRequest(c)
		if !auth.HasScopes(principal, required...) {
			forbidden(c)
			return
		}
		c.Next()
	}
}

func principalFromRequest(c *gin.Context) auth.Principal {
	value, _ := c.Get("principal")
	principal, _ := value.(auth.Principal)
	return principal
}

func tenantIDFromRequest(c *gin.Context) string {
	principal := principalFromRequest(c)
	if principal.TenantID != "" {
		return principal.TenantID
	}
	if principal, ok := auth.PrincipalFromContext(c.Request.Context()); ok && principal.TenantID != "" {
		return principal.TenantID
	}
	return "default"
}

func bearerTokenFromRequest(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return ""
	}
	return strings.TrimSpace(value[len("Bearer "):])
}

func publicBaseURL(c *gin.Context) string {
	if host := c.Request.Header.Get("X-Forwarded-Host"); host != "" {
		scheme := c.Request.Header.Get("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https"
		}
		return scheme + "://" + host
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrUnauthorized):
		unauthorized(c)
	case errors.Is(err, repository.ErrForbidden):
		forbidden(c)
	case errors.Is(err, repository.ErrExpired):
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: api.ErrorDetail{Code: "expired_token", Message: "token or auth request expired"}})
	case errors.Is(err, repository.ErrPending):
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: api.ErrorDetail{Code: "authorization_pending", Message: "authorization pending"}})
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: api.ErrorDetail{Code: "not_found", Message: "resource not found"}})
	case errors.Is(err, repository.ErrConflict):
		c.JSON(http.StatusConflict, api.ErrorResponse{Error: api.ErrorDetail{Code: "conflict", Message: "resource state conflict"}})
	default:
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: api.ErrorDetail{Code: "internal_error", Message: err.Error()}})
	}
}

func badRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: api.ErrorDetail{Code: "bad_request", Message: err.Error()}})
}

func unauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: api.ErrorDetail{Code: "unauthorized", Message: "missing or invalid bearer token"}})
}

func forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, api.ErrorResponse{Error: api.ErrorDetail{Code: "forbidden", Message: "missing required scope"}})
}
