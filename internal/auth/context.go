package auth

import (
	"context"
	"slices"
)

type Principal struct {
	PrincipalID   string
	PrincipalType string
	TenantID      string
	DisplayName   string
	Scopes        []string
	TokenID       int64
	TokenType     string
}

type contextKey string

const principalContextKey contextKey = "llm_wiki_principal"

const (
	ScopeSpacesRead       = "spaces.read"
	ScopeNamespacesRead   = "namespaces.read"
	ScopeNamespacesWrite  = "namespaces.write"
	ScopeDocumentsRead    = "documents.read"
	ScopeDocumentsWrite   = "documents.write"
	ScopeDocumentsArchive = "documents.archive"
	ScopeRevisionsRead    = "revisions.read"
	ScopeMCPInvoke        = "mcp.invoke"
	ScopeTokensIssue      = "tokens.issue"
	ScopeTokensRevoke     = "tokens.revoke"
	ScopeAdminTenants     = "admin.tenants"
)

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value := ctx.Value(principalContextKey)
	principal, ok := value.(Principal)
	return principal, ok
}

func HasScopes(principal Principal, required ...string) bool {
	if len(required) == 0 {
		return true
	}
	for _, item := range required {
		if !slices.Contains(principal.Scopes, item) {
			return false
		}
	}
	return true
}
