package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/mcpserver"
	"github.com/ifuryst/llm-wiki/internal/service"
)

func registerMCPRoutes(engine *gin.Engine, svc *service.Service, manager *mcpserver.Manager) {
	engine.Any("/mcp", gin.WrapH(mcpAuthHandler(svc, mcpserver.StreamableHTTPHandler(manager))))
	engine.Any("/sse", gin.WrapH(mcpAuthHandler(svc, mcpserver.SSEHandler(manager))))
}

func mcpAuthHandler(svc *service.Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerTokenFromRequest(r)
		if token == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		principal, err := svc.AuthenticateBearerToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		if !auth.HasScopes(principal, auth.ScopeMCPInvoke) {
			http.Error(w, "missing mcp.invoke scope", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithPrincipal(r.Context(), principal)))
	})
}
