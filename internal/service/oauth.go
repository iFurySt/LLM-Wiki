package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type oauthIdentity struct {
	Subject       string
	Email         string
	Username      string
	DisplayName   string
	EmailVerified bool
}

func (s *Service) ListOAuthProviders(ctx context.Context, enabledOnly bool) (api.ListOAuthProvidersResponse, error) {
	items, err := s.repo.ListOAuthProviders(ctx, enabledOnly)
	if err != nil {
		return api.ListOAuthProvidersResponse{}, err
	}
	resp := make([]api.OAuthProviderResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, oauthProviderResponse(item, false))
	}
	return api.ListOAuthProvidersResponse{Items: resp}, nil
}

func (s *Service) ListAdminOAuthProviders(ctx context.Context) (api.ListOAuthProvidersResponse, error) {
	items, err := s.repo.ListOAuthProviders(ctx, false)
	if err != nil {
		return api.ListOAuthProvidersResponse{}, err
	}
	resp := make([]api.OAuthProviderResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, oauthProviderResponse(item, true))
	}
	return api.ListOAuthProvidersResponse{Items: resp}, nil
}

func (s *Service) UpsertOAuthProvider(ctx context.Context, req api.UpsertOAuthProviderRequest) (api.OAuthProviderResponse, error) {
	item, err := s.repo.UpsertOAuthProvider(ctx, repository.OAuthProviderRecord{
		Name:              normalizeKey(req.Name),
		DisplayName:       strings.TrimSpace(req.DisplayName),
		AuthURL:           strings.TrimSpace(req.AuthURL),
		TokenURL:          strings.TrimSpace(req.TokenURL),
		UserinfoURL:       strings.TrimSpace(req.UserinfoURL),
		ClientID:          strings.TrimSpace(req.ClientID),
		ClientSecret:      strings.TrimSpace(req.ClientSecret),
		Scopes:            uniqueScopes(req.Scopes),
		Enabled:           req.Enabled,
		AutoCreateUsers:   req.AutoCreateUsers,
		AutoCreateTenants: req.AutoCreateNS,
	})
	if err != nil {
		return api.OAuthProviderResponse{}, err
	}
	return oauthProviderResponse(item, true), nil
}

func (s *Service) ResolveOAuthAuthorizeURL(ctx context.Context, baseURL string, requestID string, explicitProvider string) (string, repository.AuthRequest, []repository.OAuthProviderRecord, error) {
	request, err := s.repo.GetAuthRequest(ctx, requestID)
	if err != nil {
		return "", repository.AuthRequest{}, nil, err
	}
	providers, err := s.repo.ListOAuthProviders(ctx, true)
	if err != nil {
		return "", repository.AuthRequest{}, nil, err
	}
	providerName := strings.TrimSpace(explicitProvider)
	if providerName == "" {
		providerName = strings.TrimSpace(request.OAuthProvider)
	}
	if providerName == "" && len(providers) == 1 {
		providerName = providers[0].Name
	}
	if providerName == "" {
		return "", request, providers, nil
	}
	provider, err := s.repo.GetOAuthProvider(ctx, providerName)
	if err != nil {
		return "", repository.AuthRequest{}, nil, err
	}
	if !provider.Enabled {
		return "", repository.AuthRequest{}, nil, repository.ErrForbidden
	}
	redirectURI := strings.TrimRight(baseURL, "/") + "/auth/oauth/callback?request_id=" + url.QueryEscape(request.ID) + "&provider=" + url.QueryEscape(provider.Name)
	values := url.Values{}
	values.Set("client_id", provider.ClientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("response_type", "code")
	values.Set("scope", strings.Join(provider.Scopes, " "))
	values.Set("state", request.ID)
	return provider.AuthURL + "?" + values.Encode(), request, providers, nil
}

func (s *Service) CompleteOAuthBrowserLogin(ctx context.Context, requestID string, providerName string, code string, baseURL string) (repository.AuthRequest, error) {
	request, err := s.repo.GetAuthRequest(ctx, requestID)
	if err != nil {
		return repository.AuthRequest{}, err
	}
	provider, err := s.repo.GetOAuthProvider(ctx, providerName)
	if err != nil {
		return repository.AuthRequest{}, err
	}
	identity, err := fetchOAuthIdentity(ctx, provider, code, strings.TrimRight(baseURL, "/")+"/auth/oauth/callback?request_id="+url.QueryEscape(requestID))
	if err != nil {
		return repository.AuthRequest{}, err
	}
	principalID, tenantID, displayName, err := s.resolveOrProvisionOAuthPrincipal(ctx, provider, identity)
	if err != nil {
		return repository.AuthRequest{}, err
	}
	request.NS = tenantID
	request.DisplayName = displayName
	return s.repo.ApproveAuthRequestForPrincipal(ctx, request.ID, principalID)
}

func (s *Service) resolveOrProvisionOAuthPrincipal(ctx context.Context, provider repository.OAuthProviderRecord, identity oauthIdentity) (string, string, string, error) {
	account, err := s.repo.GetOAuthAccount(ctx, provider.Name, identity.Subject)
	if err == nil {
		principal, err := s.repo.GetPrincipal(ctx, account.PrincipalID)
		if err != nil {
			return "", "", "", err
		}
		return principal.ID, principal.NS, principal.DisplayName, nil
	}
	if err != repository.ErrNotFound {
		return "", "", "", err
	}
	if !provider.AutoCreateUsers {
		return "", "", "", repository.ErrUnauthorized
	}

	username := deriveUsername(identity)
	displayName := deriveDisplayName(identity)
	tenantID := ""
	if provider.AutoCreateTenants {
		var err error
		tenantID, err = s.allocatePersonalTenantID(ctx, identity)
		if err != nil {
			return "", "", "", err
		}
	}
	if tenantID == "" {
		tenantID = username
	}
	passwordHash, err := oauthPasswordHash()
	if err != nil {
		return "", "", "", err
	}
	user, err := s.repo.CreateUser(ctx, tenantID, username, displayName, passwordHash, true)
	if err != nil {
		return "", "", "", err
	}
	_, err = s.repo.CreateOAuthAccount(ctx, repository.OAuthAccountRecord{
		ProviderName:    provider.Name,
		ExternalSubject: identity.Subject,
		PrincipalID:     user.PrincipalID,
		NS:              user.NS,
		Email:           identity.Email,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
	})
	if err != nil {
		return "", "", "", err
	}
	return user.PrincipalID, user.NS, user.DisplayName, nil
}

func (s *Service) allocatePersonalTenantID(ctx context.Context, identity oauthIdentity) (string, error) {
	base := normalizeKey(firstNonEmpty(identity.Username, emailLocalPart(identity.Email), identity.DisplayName, "ns"))
	if base == "" {
		base = "ns"
	}
	for i := 0; i < 100; i++ {
		candidate := base
		if i > 0 {
			candidate = fmt.Sprintf("%s-%d", base, i+1)
		}
		exists, err := s.repo.TenantExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to allocate ns")
}

func fetchOAuthIdentity(ctx context.Context, provider repository.OAuthProviderRecord, code string, redirectURI string) (oauthIdentity, error) {
	tokenValues := url.Values{}
	tokenValues.Set("grant_type", "authorization_code")
	tokenValues.Set("code", strings.TrimSpace(code))
	tokenValues.Set("redirect_uri", redirectURI)
	tokenValues.Set("client_id", provider.ClientID)
	tokenValues.Set("client_secret", provider.ClientSecret)

	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(tokenValues.Encode()))
	if err != nil {
		return oauthIdentity{}, err
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	tokenResp, err := httpClient.Do(tokenReq)
	if err != nil {
		return oauthIdentity{}, err
	}
	defer tokenResp.Body.Close()
	if tokenResp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(tokenResp.Body, 4096))
		return oauthIdentity{}, fmt.Errorf("oauth token exchange failed: %s", strings.TrimSpace(string(body)))
	}
	var tokenPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenPayload); err != nil {
		return oauthIdentity{}, err
	}
	if strings.TrimSpace(tokenPayload.AccessToken) == "" {
		return oauthIdentity{}, fmt.Errorf("oauth token response missing access_token")
	}

	userReq, err := http.NewRequestWithContext(ctx, http.MethodGet, provider.UserinfoURL, nil)
	if err != nil {
		return oauthIdentity{}, err
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenPayload.AccessToken)
	userReq.Header.Set("Accept", "application/json")
	userResp, err := httpClient.Do(userReq)
	if err != nil {
		return oauthIdentity{}, err
	}
	defer userResp.Body.Close()
	if userResp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(userResp.Body, 4096))
		return oauthIdentity{}, fmt.Errorf("oauth userinfo fetch failed: %s", strings.TrimSpace(string(body)))
	}
	var profile map[string]any
	if err := json.NewDecoder(userResp.Body).Decode(&profile); err != nil {
		return oauthIdentity{}, err
	}
	return oauthIdentity{
		Subject:       firstMapString(profile, "sub", "id"),
		Email:         firstMapString(profile, "email"),
		Username:      firstMapString(profile, "preferred_username", "login", "username"),
		DisplayName:   firstMapString(profile, "name", "display_name"),
		EmailVerified: firstMapBool(profile, "email_verified"),
	}, nil
}

func oauthProviderResponse(item repository.OAuthProviderRecord, includeSecrets bool) api.OAuthProviderResponse {
	resp := api.OAuthProviderResponse{
		Name:            item.Name,
		DisplayName:     item.DisplayName,
		Scopes:          item.Scopes,
		Enabled:         item.Enabled,
		AutoCreateUsers: item.AutoCreateUsers,
		AutoCreateNS:    item.AutoCreateTenants,
		CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       item.UpdatedAt.Format(time.RFC3339),
	}
	if includeSecrets {
		resp.AuthURL = item.AuthURL
		resp.TokenURL = item.TokenURL
		resp.UserinfoURL = item.UserinfoURL
		resp.ClientID = item.ClientID
	}
	return resp
}

func firstMapString(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if raw, ok := values[key]; ok {
			switch value := raw.(type) {
			case string:
				if strings.TrimSpace(value) != "" {
					return strings.TrimSpace(value)
				}
			case float64:
				return strings.TrimSpace(fmt.Sprintf("%.0f", value))
			}
		}
	}
	return ""
}

func firstMapBool(values map[string]any, keys ...string) bool {
	for _, key := range keys {
		if raw, ok := values[key]; ok {
			if value, ok := raw.(bool); ok {
				return value
			}
		}
	}
	return false
}

func deriveUsername(identity oauthIdentity) string {
	value := normalizeUsername(firstNonEmpty(identity.Username, emailLocalPart(identity.Email), identity.DisplayName, identity.Subject))
	if value == "" {
		return "user"
	}
	return value
}

func deriveDisplayName(identity oauthIdentity) string {
	value := strings.TrimSpace(firstNonEmpty(identity.DisplayName, identity.Username, emailLocalPart(identity.Email), "LLM-Wiki User"))
	if value == "" {
		return "LLM-Wiki User"
	}
	return value
}

func emailLocalPart(email string) string {
	value := strings.TrimSpace(email)
	if idx := strings.Index(value, "@"); idx > 0 {
		return value[:idx]
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func oauthPasswordHash() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(raw)
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
