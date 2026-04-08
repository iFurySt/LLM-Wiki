package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/repository"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
	tokenTypeService = "service"
)

var interactiveScopes = []string{
	auth.ScopeNSRead,
	auth.ScopeFoldersRead,
	auth.ScopeFoldersWrite,
	auth.ScopeDocumentsRead,
	auth.ScopeDocumentsWrite,
	auth.ScopeDocumentsArchive,
	auth.ScopeRevisionsRead,
	auth.ScopeMCPInvoke,
}

func (s *Service) BootstrapToken(ctx context.Context, tenantID string, displayName string, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return nil
	}
	principal, err := s.repo.EnsurePrincipal(ctx, tenantID, "admin", displayName)
	if err != nil {
		return err
	}
	if _, err := s.repo.AuthenticateToken(ctx, rawToken); err == nil {
		return nil
	} else if err != nil && err != repository.ErrUnauthorized && err != repository.ErrExpired {
		return err
	}
	hashExpiry := time.Now().AddDate(10, 0, 0)
	_, err = s.repo.UpsertStaticToken(ctx, repository.IssueTokenParams{
		TokenType:   tokenTypeService,
		PrincipalID: principal.ID,
		NS:          tenantID,
		DisplayName: "bootstrap-admin",
		Scopes: []string{
			auth.ScopeNSRead,
			auth.ScopeFoldersRead,
			auth.ScopeFoldersWrite,
			auth.ScopeDocumentsRead,
			auth.ScopeDocumentsWrite,
			auth.ScopeDocumentsArchive,
			auth.ScopeRevisionsRead,
			auth.ScopeMCPInvoke,
			auth.ScopeTokensIssue,
			auth.ScopeTokensRevoke,
			auth.ScopeAdminNS,
		},
		ExpiresAt: &hashExpiry,
	}, rawToken)
	return err
}

func (s *Service) AuthenticateBearerToken(ctx context.Context, rawToken string) (auth.Principal, error) {
	result, err := s.repo.AuthenticateToken(ctx, rawToken)
	if err != nil {
		return auth.Principal{}, err
	}
	return auth.Principal{
		PrincipalID:   result.Principal.ID,
		PrincipalType: result.Principal.PrincipalType,
		NS:            result.Principal.NS,
		DisplayName:   result.Principal.DisplayName,
		Scopes:        result.Token.Scopes,
		TokenID:       result.Token.ID,
		TokenType:     result.Token.TokenType,
	}, nil
}

func (s *Service) WhoAmI(ctx context.Context) (api.WhoAmIResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.WhoAmIResponse{}, repository.ErrUnauthorized
	}
	return api.WhoAmIResponse{
		PrincipalID:   principal.PrincipalID,
		PrincipalType: principal.PrincipalType,
		DisplayName:   principal.DisplayName,
		NS:            principal.NS,
		Scopes:        principal.Scopes,
		TokenID:       principal.TokenID,
		TokenType:     principal.TokenType,
	}, nil
}

func (s *Service) StartBrowserLogin(ctx context.Context, baseURL string, req api.StartBrowserLoginRequest) (api.StartBrowserLoginResponse, error) {
	request, err := s.repo.CreateAuthRequest(ctx, repository.CreateAuthRequestParams{
		FlowType:            "browser",
		NS:                  strings.TrimSpace(req.NS),
		OAuthProvider:       normalizeKey(req.Provider),
		DisplayName:         strings.TrimSpace(req.DisplayName),
		Scopes:              filterInteractiveScopes(req.Scopes),
		State:               req.State,
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute),
	})
	if err != nil {
		return api.StartBrowserLoginResponse{}, err
	}
	return api.StartBrowserLoginResponse{
		RequestID:    request.ID,
		AuthorizeURL: strings.TrimRight(baseURL, "/") + "/auth/authorize?request_id=" + request.ID,
		ExpiresAt:    request.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) StartDeviceLogin(ctx context.Context, baseURL string, req api.StartDeviceLoginRequest) (api.StartDeviceLoginResponse, error) {
	deviceCode, err := auth.RandomID("device", 16)
	if err != nil {
		return api.StartDeviceLoginResponse{}, err
	}
	userCode, err := auth.RandomCode(8)
	if err != nil {
		return api.StartDeviceLoginResponse{}, err
	}
	request, err := s.repo.CreateAuthRequest(ctx, repository.CreateAuthRequestParams{
		FlowType:    "device",
		NS:          strings.TrimSpace(req.NS),
		DisplayName: strings.TrimSpace(req.DisplayName),
		Scopes:      filterInteractiveScopes(req.Scopes),
		DeviceCode:  deviceCode,
		UserCode:    userCode,
		ExpiresAt:   time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return api.StartDeviceLoginResponse{}, err
	}
	return api.StartDeviceLoginResponse{
		RequestID:       request.ID,
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: strings.TrimRight(baseURL, "/") + "/auth/device",
		IntervalSeconds: 5,
		ExpiresAt:       request.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) ApproveAuthRequest(ctx context.Context, requestID string, displayName string) (repository.AuthRequest, error) {
	request, _, err := s.repo.ApproveAuthRequest(ctx, requestID, strings.TrimSpace(displayName))
	return request, err
}

func (s *Service) ApproveDeviceCode(ctx context.Context, userCode string, displayName string) (repository.AuthRequest, error) {
	request, err := s.repo.GetAuthRequestByUserCode(ctx, strings.TrimSpace(strings.ToUpper(userCode)))
	if err != nil {
		return repository.AuthRequest{}, err
	}
	return s.ApproveAuthRequest(ctx, request.ID, displayName)
}

func (s *Service) ExchangeToken(ctx context.Context, req api.TokenExchangeRequest) (api.TokenExchangeResponse, error) {
	switch req.GrantType {
	case "authorization_code":
		return s.exchangeAuthorizationCode(ctx, req.Code, req.CodeVerifier)
	case "urn:ietf:params:oauth:grant-type:device_code":
		return s.exchangeDeviceCode(ctx, req.DeviceCode)
	case "refresh_token":
		return s.exchangeRefreshToken(ctx, req.RefreshToken)
	default:
		return api.TokenExchangeResponse{}, fmt.Errorf("unsupported grant_type")
	}
}

func (s *Service) CreateServicePrincipal(ctx context.Context, tenantID string, req api.CreateServicePrincipalRequest) (api.ServicePrincipalResponse, error) {
	item, err := s.repo.EnsurePrincipal(ctx, tenantID, "service", strings.TrimSpace(req.DisplayName))
	if err != nil {
		return api.ServicePrincipalResponse{}, err
	}
	return api.ServicePrincipalResponse{
		ID:            item.ID,
		NS:            item.NS,
		PrincipalType: item.PrincipalType,
		DisplayName:   item.DisplayName,
		CreatedAt:     item.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) ListServicePrincipals(ctx context.Context, tenantID string) (api.ListServicePrincipalsResponse, error) {
	items, err := s.repo.ListPrincipalsByType(ctx, tenantID, "service")
	if err != nil {
		return api.ListServicePrincipalsResponse{}, err
	}
	resp := make([]api.ServicePrincipalResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, api.ServicePrincipalResponse{
			ID:            item.ID,
			NS:            item.NS,
			PrincipalType: item.PrincipalType,
			DisplayName:   item.DisplayName,
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
		})
	}
	return api.ListServicePrincipalsResponse{Items: resp}, nil
}

func (s *Service) IssueServiceToken(ctx context.Context, tenantID string, req api.IssueTokenRequest) (api.TokenResponse, error) {
	principal, err := s.repo.GetPrincipal(ctx, req.PrincipalID)
	if err != nil {
		return api.TokenResponse{}, err
	}
	if principal.NS != tenantID {
		return api.TokenResponse{}, repository.ErrForbidden
	}
	var expiresAt *time.Time
	if req.ExpiresInSeconds > 0 {
		value := time.Now().Add(time.Duration(req.ExpiresInSeconds) * time.Second)
		expiresAt = &value
	}
	var createdBy *string
	if caller, ok := auth.PrincipalFromContext(ctx); ok {
		createdBy = &caller.PrincipalID
	}
	record, plaintext, err := s.repo.IssueToken(ctx, repository.IssueTokenParams{
		TokenType:            tokenTypeService,
		PrincipalID:          principal.ID,
		NS:                   tenantID,
		DisplayName:          strings.TrimSpace(req.DisplayName),
		Scopes:               uniqueScopes(req.Scopes),
		ExpiresAt:            expiresAt,
		CreatedByPrincipalID: createdBy,
	})
	if err != nil {
		return api.TokenResponse{}, err
	}
	return toTokenResponse(record, plaintext), nil
}

func (s *Service) ListTokens(ctx context.Context, tenantID string) (api.ListTokensResponse, error) {
	items, err := s.repo.ListTokens(ctx, tenantID)
	if err != nil {
		return api.ListTokensResponse{}, err
	}
	resp := make([]api.TokenResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toTokenResponse(item, ""))
	}
	return api.ListTokensResponse{Items: resp}, nil
}

func (s *Service) RevokeToken(ctx context.Context, tenantID string, tokenID int64) (api.TokenResponse, error) {
	item, err := s.repo.RevokeToken(ctx, tenantID, tokenID)
	if err != nil {
		return api.TokenResponse{}, err
	}
	return toTokenResponse(item, ""), nil
}

func (s *Service) exchangeAuthorizationCode(ctx context.Context, code string, verifier string) (api.TokenExchangeResponse, error) {
	request, err := s.repo.ExchangeAuthorizationCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	if request.PrincipalID == nil {
		return api.TokenExchangeResponse{}, repository.ErrPending
	}
	if request.ExpiresAt.Before(time.Now()) {
		return api.TokenExchangeResponse{}, repository.ErrExpired
	}
	if !verifyPKCE(verifier, request.CodeChallenge, request.CodeChallengeMethod) {
		return api.TokenExchangeResponse{}, repository.ErrUnauthorized
	}
	return s.issueInteractiveTokens(ctx, request)
}

func (s *Service) exchangeDeviceCode(ctx context.Context, deviceCode string) (api.TokenExchangeResponse, error) {
	request, err := s.repo.GetAuthRequestByDeviceCode(ctx, strings.TrimSpace(deviceCode))
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	if request.ExpiresAt.Before(time.Now()) {
		return api.TokenExchangeResponse{}, repository.ErrExpired
	}
	if request.PrincipalID == nil {
		return api.TokenExchangeResponse{}, repository.ErrPending
	}
	return s.issueInteractiveTokens(ctx, request)
}

func (s *Service) exchangeRefreshToken(ctx context.Context, rawToken string) (api.TokenExchangeResponse, error) {
	authn, err := s.repo.AuthenticateRefreshToken(ctx, strings.TrimSpace(rawToken))
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	request := repository.AuthRequest{
		NS:          authn.Principal.NS,
		Scopes:      authn.Token.Scopes,
		PrincipalID: &authn.Principal.ID,
	}
	return s.issueInteractiveTokens(ctx, request)
}

func (s *Service) issueInteractiveTokens(ctx context.Context, request repository.AuthRequest) (api.TokenExchangeResponse, error) {
	if request.PrincipalID == nil {
		return api.TokenExchangeResponse{}, repository.ErrUnauthorized
	}
	if user, err := s.repo.GetUserByPrincipalID(ctx, *request.PrincipalID); err == nil && user.IsAdmin {
		request.Scopes = uniqueScopes(append(request.Scopes, auth.ScopeTokensIssue, auth.ScopeTokensRevoke, auth.ScopeAdminNS))
	}
	accessExpiry := time.Now().Add(1 * time.Hour)
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour)

	access, accessToken, err := s.repo.IssueToken(ctx, repository.IssueTokenParams{
		TokenType:   tokenTypeAccess,
		PrincipalID: *request.PrincipalID,
		NS:          request.NS,
		DisplayName: "cli-access",
		Scopes:      request.Scopes,
		ExpiresAt:   &accessExpiry,
	})
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	_, refreshToken, err := s.repo.IssueToken(ctx, repository.IssueTokenParams{
		TokenType:   tokenTypeRefresh,
		PrincipalID: *request.PrincipalID,
		NS:          request.NS,
		DisplayName: "cli-refresh",
		Scopes:      request.Scopes,
		ExpiresAt:   &refreshExpiry,
	})
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	return api.TokenExchangeResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(time.Until(accessExpiry).Seconds()),
		RefreshToken: refreshToken,
		Scopes:       access.Scopes,
		NS:           access.NS,
		PrincipalID:  access.PrincipalID,
	}, nil
}

func toTokenResponse(item repository.TokenRecord, plaintext string) api.TokenResponse {
	resp := api.TokenResponse{
		ID:          item.ID,
		TokenType:   item.TokenType,
		PrincipalID: item.PrincipalID,
		NS:          item.NS,
		DisplayName: item.DisplayName,
		TokenPrefix: item.TokenPrefix,
		Scopes:      item.Scopes,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
	}
	if item.ExpiresAt != nil {
		resp.ExpiresAt = item.ExpiresAt.Format(time.RFC3339)
	}
	if item.LastUsedAt != nil {
		resp.LastUsedAt = item.LastUsedAt.Format(time.RFC3339)
	}
	if item.RevokedAt != nil {
		resp.RevokedAt = item.RevokedAt.Format(time.RFC3339)
	}
	if item.CreatedByPrincipalID != nil {
		resp.CreatedByPrincipalID = *item.CreatedByPrincipalID
	}
	resp.PlaintextToken = plaintext
	return resp
}

func filterInteractiveScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return slices.Clone(interactiveScopes)
	}
	filtered := make([]string, 0, len(scopes))
	for _, item := range uniqueScopes(scopes) {
		if slices.Contains(interactiveScopes, item) {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return slices.Clone(interactiveScopes)
	}
	return filtered
}

func uniqueScopes(scopes []string) []string {
	items := make([]string, 0, len(scopes))
	seen := make(map[string]struct{}, len(scopes))
	for _, item := range scopes {
		scope := strings.TrimSpace(item)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		items = append(items, scope)
	}
	return items
}

func verifyPKCE(verifier string, challenge string, method string) bool {
	if method != "S256" {
		return false
	}
	sum := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return encoded == challenge
}
