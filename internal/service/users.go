package service

import (
	"context"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const webSessionTTL = 24 * time.Hour

func (s *Service) SetupStatus(ctx context.Context, defaultTenant string) (api.SetupStatusResponse, error) {
	exists, err := s.repo.AnyAdminUsersExist(ctx)
	if err != nil {
		return api.SetupStatusResponse{}, err
	}
	return api.SetupStatusResponse{
		Initialized:   exists,
		DefaultTenant: defaultTenant,
	}, nil
}

func (s *Service) Initialize(ctx context.Context, req api.InitializeRequest) (api.UserResponse, error) {
	exists, err := s.repo.AnyAdminUsersExist(ctx)
	if err != nil {
		return api.UserResponse{}, err
	}
	if exists {
		return api.UserResponse{}, repository.ErrConflict
	}
	return s.createUser(ctx, strings.TrimSpace(req.TenantID), api.CreateUserRequest{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		IsAdmin:     true,
	})
}

func (s *Service) LoginUser(ctx context.Context, tenantID string, username string, password string) (repository.UserRecord, auth.Principal, error) {
	user, err := s.repo.GetUserByUsername(ctx, strings.TrimSpace(tenantID), normalizeUsername(username))
	if err != nil {
		return repository.UserRecord{}, auth.Principal{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return repository.UserRecord{}, auth.Principal{}, repository.ErrUnauthorized
	}
	tokenType := "access"
	principalType := "user"
	if user.IsAdmin {
		principalType = "admin"
	}
	principal := auth.Principal{
		PrincipalID:   user.PrincipalID,
		PrincipalType: principalType,
		TenantID:      user.TenantID,
		DisplayName:   user.DisplayName,
		Scopes:        adminScopes(user.IsAdmin),
		TokenType:     tokenType,
	}
	return user, principal, nil
}

func (s *Service) CreateWebSession(ctx context.Context, principalID string, tenantID string) (repository.WebSessionRecord, error) {
	return s.repo.CreateWebSession(ctx, principalID, tenantID, webSessionTTL)
}

func (s *Service) GetWebSession(ctx context.Context, sessionID string) (repository.WebSessionRecord, repository.UserRecord, auth.Principal, error) {
	session, user, err := s.repo.GetWebSession(ctx, sessionID)
	if err != nil {
		return repository.WebSessionRecord{}, repository.UserRecord{}, auth.Principal{}, err
	}
	if session.ExpiresAt.Before(time.Now()) {
		return repository.WebSessionRecord{}, repository.UserRecord{}, auth.Principal{}, repository.ErrExpired
	}
	principalType := "user"
	if user.IsAdmin {
		principalType = "admin"
	}
	principal := auth.Principal{
		PrincipalID:   user.PrincipalID,
		PrincipalType: principalType,
		TenantID:      user.TenantID,
		DisplayName:   user.DisplayName,
		Scopes:        adminScopes(user.IsAdmin),
		TokenType:     "web_session",
	}
	return session, user, principal, nil
}

func (s *Service) DeleteWebSession(ctx context.Context, sessionID string) error {
	return s.repo.DeleteWebSession(ctx, sessionID)
}

func (s *Service) ListUsers(ctx context.Context, tenantID string) (api.ListUsersResponse, error) {
	items, err := s.repo.ListUsers(ctx, tenantID)
	if err != nil {
		return api.ListUsersResponse{}, err
	}
	resp := make([]api.UserResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, userResponse(item))
	}
	return api.ListUsersResponse{Items: resp}, nil
}

func (s *Service) CreateUser(ctx context.Context, tenantID string, req api.CreateUserRequest) (api.UserResponse, error) {
	return s.createUser(ctx, tenantID, req)
}

func (s *Service) ApproveAuthRequestWithPassword(ctx context.Context, requestID string, username string, password string) (repository.AuthRequest, error) {
	request, err := s.repo.GetAuthRequest(ctx, requestID)
	if err != nil {
		return repository.AuthRequest{}, err
	}
	user, _, err := s.LoginUser(ctx, request.TenantID, username, password)
	if err != nil {
		return repository.AuthRequest{}, err
	}
	return s.ApproveAuthRequest(ctx, request.ID, user.DisplayName)
}

func (s *Service) ApproveDeviceCodeWithPassword(ctx context.Context, userCode string, username string, password string) (repository.AuthRequest, error) {
	request, err := s.repo.GetAuthRequestByUserCode(ctx, strings.TrimSpace(strings.ToUpper(userCode)))
	if err != nil {
		return repository.AuthRequest{}, err
	}
	return s.ApproveAuthRequestWithPassword(ctx, request.ID, username, password)
}

func (s *Service) createUser(ctx context.Context, tenantID string, req api.CreateUserRequest) (api.UserResponse, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return api.UserResponse{}, err
	}
	item, err := s.repo.CreateUser(
		ctx,
		tenantID,
		normalizeUsername(req.Username),
		strings.TrimSpace(req.DisplayName),
		string(passwordHash),
		req.IsAdmin,
	)
	if err != nil {
		return api.UserResponse{}, err
	}
	return userResponse(item), nil
}

func userResponse(item repository.UserRecord) api.UserResponse {
	return api.UserResponse{
		ID:          item.ID,
		PrincipalID: item.PrincipalID,
		TenantID:    item.TenantID,
		Username:    item.Username,
		DisplayName: item.DisplayName,
		IsAdmin:     item.IsAdmin,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
	}
}

func normalizeUsername(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "_", "-")
	if value == "" {
		return value
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func adminScopes(isAdmin bool) []string {
	scopes := []string{
		auth.ScopeSpacesRead,
		auth.ScopeNamespacesRead,
		auth.ScopeNamespacesWrite,
		auth.ScopeDocumentsRead,
		auth.ScopeDocumentsWrite,
		auth.ScopeDocumentsArchive,
		auth.ScopeRevisionsRead,
		auth.ScopeMCPInvoke,
	}
	if isAdmin {
		scopes = append(scopes, auth.ScopeTokensIssue, auth.ScopeTokensRevoke, auth.ScopeAdminTenants)
	}
	return scopes
}
