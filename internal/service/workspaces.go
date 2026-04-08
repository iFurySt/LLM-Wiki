package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/repository"
)

func (s *Service) ListWorkspaces(ctx context.Context) (api.ListWorkspacesResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.ListWorkspacesResponse{}, repository.ErrUnauthorized
	}
	account, err := s.repo.GetOAuthAccountByPrincipalID(ctx, principal.PrincipalID)
	if err == repository.ErrNotFound {
		spaces, err := s.repo.ListSpaces(ctx, principal.TenantID)
		if err != nil {
			return api.ListWorkspacesResponse{}, err
		}
		return api.ListWorkspacesResponse{Items: workspaceResponses(spaces)}, nil
	}
	if err != nil {
		return api.ListWorkspacesResponse{}, err
	}
	accounts, err := s.repo.ListOAuthAccountsByIdentity(ctx, account.ProviderName, account.ExternalSubject)
	if err != nil {
		return api.ListWorkspacesResponse{}, err
	}
	items := make([]repository.Space, 0, len(accounts))
	for _, item := range accounts {
		space, err := s.repo.EnsureTenantSpace(ctx, item.TenantID)
		if err != nil {
			return api.ListWorkspacesResponse{}, err
		}
		space.Role = "member"
		user, err := s.repo.GetUserByPrincipalID(ctx, item.PrincipalID)
		if err == nil && user.IsAdmin {
			space.Role = "owner"
		}
		items = append(items, space)
	}
	return api.ListWorkspacesResponse{Items: workspaceResponses(items)}, nil
}

func (s *Service) CreateWorkspace(ctx context.Context, req api.CreateWorkspaceRequest) (api.WorkspaceResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.WorkspaceResponse{}, repository.ErrUnauthorized
	}
	account, err := s.repo.GetOAuthAccountByPrincipalID(ctx, principal.PrincipalID)
	if err != nil {
		return api.WorkspaceResponse{}, repository.ErrForbidden
	}
	tenantID := normalizeKey(req.TenantID)
	if tenantID == "" {
		tenantID, err = s.allocatePersonalTenantID(ctx, oauthIdentity{
			Email:       account.Email,
			Username:    account.Username,
			DisplayName: req.DisplayName,
		})
		if err != nil {
			return api.WorkspaceResponse{}, err
		}
	}
	if err := validateKeyLike("workspace key", tenantID); err != nil {
		return api.WorkspaceResponse{}, err
	}
	exists, err := s.repo.TenantExists(ctx, tenantID)
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	if exists {
		return api.WorkspaceResponse{}, repository.ErrConflict
	}
	user, err := s.cloneOAuthIdentityIntoTenant(ctx, account, tenantID, strings.TrimSpace(req.DisplayName), true)
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	space, err := s.repo.EnsureTenantSpaceWithDisplayName(ctx, tenantID, strings.TrimSpace(req.DisplayName))
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	space.Role = "owner"
	if user.PrincipalID == "" {
		return api.WorkspaceResponse{}, fmt.Errorf("workspace principal not created")
	}
	return workspaceResponse(space), nil
}

func (s *Service) ListInvites(ctx context.Context) (api.ListInvitesResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.ListInvitesResponse{}, repository.ErrUnauthorized
	}
	items, err := s.repo.ListTenantInvites(ctx, principal.TenantID)
	if err != nil {
		return api.ListInvitesResponse{}, err
	}
	resp := make([]api.InviteResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, inviteResponse(item, false))
	}
	return api.ListInvitesResponse{Items: resp}, nil
}

func (s *Service) CreateInvite(ctx context.Context, req api.CreateInviteRequest) (api.InviteResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.InviteResponse{}, repository.ErrUnauthorized
	}
	token, err := auth.RandomID("invite", 16)
	if err != nil {
		return api.InviteResponse{}, err
	}
	expiresIn := req.ExpiresInHours
	if expiresIn <= 0 {
		expiresIn = 72
	}
	item, err := s.repo.CreateTenantInvite(ctx, repository.TenantInviteRecord{
		TenantID:             principal.TenantID,
		Email:                strings.TrimSpace(strings.ToLower(req.Email)),
		Role:                 normalizeInviteRole(req.Role),
		InviteToken:          token,
		InvitedByPrincipalID: &principal.PrincipalID,
		ExpiresAt:            time.Now().Add(time.Duration(expiresIn) * time.Hour),
	})
	if err != nil {
		return api.InviteResponse{}, err
	}
	return inviteResponse(item, true), nil
}

func (s *Service) AcceptInvite(ctx context.Context, req api.AcceptInviteRequest) (api.WorkspaceResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.WorkspaceResponse{}, repository.ErrUnauthorized
	}
	account, err := s.repo.GetOAuthAccountByPrincipalID(ctx, principal.PrincipalID)
	if err != nil {
		return api.WorkspaceResponse{}, repository.ErrForbidden
	}
	invite, err := s.repo.GetTenantInviteByToken(ctx, strings.TrimSpace(req.InviteToken))
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	if invite.AcceptedAt != nil {
		return api.WorkspaceResponse{}, repository.ErrConflict
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return api.WorkspaceResponse{}, repository.ErrExpired
	}
	if strings.TrimSpace(strings.ToLower(account.Email)) != strings.TrimSpace(strings.ToLower(invite.Email)) {
		return api.WorkspaceResponse{}, repository.ErrForbidden
	}
	accounts, err := s.repo.ListOAuthAccountsByIdentity(ctx, account.ProviderName, account.ExternalSubject)
	if err == nil {
		for _, item := range accounts {
			if item.TenantID == invite.TenantID {
				space, err := s.repo.EnsureTenantSpace(ctx, item.TenantID)
				if err != nil {
					return api.WorkspaceResponse{}, err
				}
				return workspaceResponse(space), nil
			}
		}
	}
	user, err := s.cloneOAuthIdentityIntoTenant(ctx, account, invite.TenantID, "", invite.Role == "owner" || invite.Role == "admin")
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	if _, err := s.repo.AcceptTenantInvite(ctx, invite.ID, user.PrincipalID); err != nil {
		return api.WorkspaceResponse{}, err
	}
	space, err := s.repo.EnsureTenantSpace(ctx, invite.TenantID)
	if err != nil {
		return api.WorkspaceResponse{}, err
	}
	space.Role = invite.Role
	return workspaceResponse(space), nil
}

func (s *Service) SwitchTenant(ctx context.Context, tenantID string) (api.TokenExchangeResponse, error) {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return api.TokenExchangeResponse{}, repository.ErrUnauthorized
	}
	account, err := s.repo.GetOAuthAccountByPrincipalID(ctx, principal.PrincipalID)
	if err != nil {
		return api.TokenExchangeResponse{}, repository.ErrForbidden
	}
	accounts, err := s.repo.ListOAuthAccountsByIdentity(ctx, account.ProviderName, account.ExternalSubject)
	if err != nil {
		return api.TokenExchangeResponse{}, err
	}
	for _, item := range accounts {
		if item.TenantID == tenantID {
			request := repository.AuthRequest{
				TenantID:    tenantID,
				Scopes:      interactiveScopes,
				PrincipalID: &item.PrincipalID,
			}
			return s.issueInteractiveTokens(ctx, request)
		}
	}
	return api.TokenExchangeResponse{}, repository.ErrForbidden
}

func (s *Service) cloneOAuthIdentityIntoTenant(ctx context.Context, account repository.OAuthAccountRecord, tenantID string, displayName string, isAdmin bool) (repository.UserRecord, error) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = strings.TrimSpace(account.DisplayName)
	}
	passwordHash, err := oauthPasswordHash()
	if err != nil {
		return repository.UserRecord{}, err
	}
	user, err := s.repo.CreateUser(ctx, tenantID, normalizeUsername(firstNonEmpty(account.Username, emailLocalPart(account.Email), account.DisplayName)), firstNonEmpty(name, account.DisplayName, account.Username), passwordHash, isAdmin)
	if err != nil {
		return repository.UserRecord{}, err
	}
	if _, err := s.repo.CreateOAuthAccount(ctx, repository.OAuthAccountRecord{
		ProviderName:    account.ProviderName,
		ExternalSubject: account.ExternalSubject,
		PrincipalID:     user.PrincipalID,
		TenantID:        tenantID,
		Email:           account.Email,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
	}); err != nil {
		return repository.UserRecord{}, err
	}
	_, err = s.repo.EnsureTenantSpaceWithDisplayName(ctx, tenantID, firstNonEmpty(name, account.DisplayName, tenantID))
	return user, err
}

func normalizeInviteRole(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "owner", "admin":
		return "owner"
	default:
		return "member"
	}
}

func workspaceResponses(items []repository.Space) []api.WorkspaceResponse {
	resp := make([]api.WorkspaceResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, workspaceResponse(item))
	}
	return resp
}

func workspaceResponse(item repository.Space) api.WorkspaceResponse {
	return api.WorkspaceResponse{
		TenantID:    item.TenantID,
		Key:         item.Key,
		DisplayName: item.DisplayName,
		Role:        item.Role,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
	}
}

func inviteResponse(item repository.TenantInviteRecord, includeToken bool) api.InviteResponse {
	resp := api.InviteResponse{
		ID:        item.ID,
		TenantID:  item.TenantID,
		Email:     item.Email,
		Role:      item.Role,
		ExpiresAt: item.ExpiresAt.Format(time.RFC3339),
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
	if includeToken {
		resp.Token = item.InviteToken
	}
	if item.AcceptedAt != nil {
		resp.AcceptedAt = item.AcceptedAt.Format(time.RFC3339)
	}
	return resp
}
