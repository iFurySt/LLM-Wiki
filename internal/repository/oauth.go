package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type OAuthProviderRecord struct {
	Name              string
	DisplayName       string
	AuthURL           string
	TokenURL          string
	UserinfoURL       string
	ClientID          string
	ClientSecret      string
	Scopes            []string
	Enabled           bool
	AutoCreateUsers   bool
	AutoCreateTenants bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type OAuthAccountRecord struct {
	ID              int64
	ProviderName    string
	ExternalSubject string
	PrincipalID     string
	NS              string
	Email           string
	Username        string
	DisplayName     string
	CreatedAt       time.Time
}

type TenantInviteRecord struct {
	ID                    int64
	NS                    string
	Email                 string
	Role                  string
	InviteToken           string
	InvitedByPrincipalID  *string
	AcceptedByPrincipalID *string
	AcceptedAt            *time.Time
	ExpiresAt             time.Time
	CreatedAt             time.Time
}

func (r *Repository) ListOAuthProviders(ctx context.Context, enabledOnly bool) ([]OAuthProviderRecord, error) {
	query := `
		SELECT name, display_name, auth_url, token_url, userinfo_url, client_id, client_secret, scopes_json, enabled, auto_create_users, auto_create_tenants, created_at, updated_at
		FROM oauth_providers
	`
	args := []any{}
	if enabledOnly {
		query += ` WHERE enabled = TRUE`
	}
	query += ` ORDER BY display_name ASC, name ASC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OAuthProviderRecord, 0)
	for rows.Next() {
		var item OAuthProviderRecord
		var scopesJSON string
		if err := rows.Scan(
			&item.Name,
			&item.DisplayName,
			&item.AuthURL,
			&item.TokenURL,
			&item.UserinfoURL,
			&item.ClientID,
			&item.ClientSecret,
			&scopesJSON,
			&item.Enabled,
			&item.AutoCreateUsers,
			&item.AutoCreateTenants,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Scopes, err = unmarshalScopes(scopesJSON)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetOAuthProvider(ctx context.Context, name string) (OAuthProviderRecord, error) {
	var item OAuthProviderRecord
	var scopesJSON string
	err := r.pool.QueryRow(ctx, `
		SELECT name, display_name, auth_url, token_url, userinfo_url, client_id, client_secret, scopes_json, enabled, auto_create_users, auto_create_tenants, created_at, updated_at
		FROM oauth_providers
		WHERE name = $1
	`, name).Scan(
		&item.Name,
		&item.DisplayName,
		&item.AuthURL,
		&item.TokenURL,
		&item.UserinfoURL,
		&item.ClientID,
		&item.ClientSecret,
		&scopesJSON,
		&item.Enabled,
		&item.AutoCreateUsers,
		&item.AutoCreateTenants,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OAuthProviderRecord{}, ErrNotFound
	}
	if err != nil {
		return OAuthProviderRecord{}, err
	}
	item.Scopes, err = unmarshalScopes(scopesJSON)
	return item, err
}

func (r *Repository) UpsertOAuthProvider(ctx context.Context, item OAuthProviderRecord) (OAuthProviderRecord, error) {
	scopesJSON, err := marshalScopes(item.Scopes)
	if err != nil {
		return OAuthProviderRecord{}, err
	}
	var saved OAuthProviderRecord
	err = r.pool.QueryRow(ctx, `
		INSERT INTO oauth_providers (
			name, display_name, auth_url, token_url, userinfo_url, client_id, client_secret, scopes_json, enabled, auto_create_users, auto_create_tenants, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			auth_url = EXCLUDED.auth_url,
			token_url = EXCLUDED.token_url,
			userinfo_url = EXCLUDED.userinfo_url,
			client_id = EXCLUDED.client_id,
			client_secret = EXCLUDED.client_secret,
			scopes_json = EXCLUDED.scopes_json,
			enabled = EXCLUDED.enabled,
			auto_create_users = EXCLUDED.auto_create_users,
			auto_create_tenants = EXCLUDED.auto_create_tenants,
			updated_at = NOW()
		RETURNING name, display_name, auth_url, token_url, userinfo_url, client_id, client_secret, scopes_json, enabled, auto_create_users, auto_create_tenants, created_at, updated_at
	`, item.Name, item.DisplayName, item.AuthURL, item.TokenURL, item.UserinfoURL, item.ClientID, item.ClientSecret, scopesJSON, item.Enabled, item.AutoCreateUsers, item.AutoCreateTenants).Scan(
		&saved.Name,
		&saved.DisplayName,
		&saved.AuthURL,
		&saved.TokenURL,
		&saved.UserinfoURL,
		&saved.ClientID,
		&saved.ClientSecret,
		&scopesJSON,
		&saved.Enabled,
		&saved.AutoCreateUsers,
		&saved.AutoCreateTenants,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return OAuthProviderRecord{}, err
	}
	saved.Scopes, err = unmarshalScopes(scopesJSON)
	return saved, err
}

func (r *Repository) GetOAuthAccount(ctx context.Context, providerName string, externalSubject string) (OAuthAccountRecord, error) {
	var item OAuthAccountRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, provider_name, external_subject, principal_id, ns, email, username, display_name, created_at
		FROM oauth_accounts
		WHERE provider_name = $1 AND external_subject = $2
		ORDER BY created_at ASC
		LIMIT 1
	`, providerName, externalSubject).Scan(
		&item.ID,
		&item.ProviderName,
		&item.ExternalSubject,
		&item.PrincipalID,
		&item.NS,
		&item.Email,
		&item.Username,
		&item.DisplayName,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OAuthAccountRecord{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateOAuthAccount(ctx context.Context, item OAuthAccountRecord) (OAuthAccountRecord, error) {
	var saved OAuthAccountRecord
	err := r.pool.QueryRow(ctx, `
		INSERT INTO oauth_accounts (provider_name, external_subject, principal_id, ns, email, username, display_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, provider_name, external_subject, principal_id, ns, email, username, display_name, created_at
	`, item.ProviderName, item.ExternalSubject, item.PrincipalID, item.NS, item.Email, item.Username, item.DisplayName).Scan(
		&saved.ID,
		&saved.ProviderName,
		&saved.ExternalSubject,
		&saved.PrincipalID,
		&saved.NS,
		&saved.Email,
		&saved.Username,
		&saved.DisplayName,
		&saved.CreatedAt,
	)
	return saved, err
}

func (r *Repository) GetOAuthAccountByPrincipalID(ctx context.Context, principalID string) (OAuthAccountRecord, error) {
	var item OAuthAccountRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, provider_name, external_subject, principal_id, ns, email, username, display_name, created_at
		FROM oauth_accounts
		WHERE principal_id = $1
	`, principalID).Scan(
		&item.ID,
		&item.ProviderName,
		&item.ExternalSubject,
		&item.PrincipalID,
		&item.NS,
		&item.Email,
		&item.Username,
		&item.DisplayName,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return OAuthAccountRecord{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) ListOAuthAccountsByIdentity(ctx context.Context, providerName string, externalSubject string) ([]OAuthAccountRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, provider_name, external_subject, principal_id, ns, email, username, display_name, created_at
		FROM oauth_accounts
		WHERE provider_name = $1 AND external_subject = $2
		ORDER BY ns ASC
	`, providerName, externalSubject)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]OAuthAccountRecord, 0)
	for rows.Next() {
		var item OAuthAccountRecord
		if err := rows.Scan(
			&item.ID,
			&item.ProviderName,
			&item.ExternalSubject,
			&item.PrincipalID,
			&item.NS,
			&item.Email,
			&item.Username,
			&item.DisplayName,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) CreateTenantInvite(ctx context.Context, item TenantInviteRecord) (TenantInviteRecord, error) {
	var saved TenantInviteRecord
	err := r.pool.QueryRow(ctx, `
		INSERT INTO ns_invites (ns, email, role, invite_token, invited_by_principal_id, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, ns, email, role, invite_token, invited_by_principal_id, accepted_by_principal_id, accepted_at, expires_at, created_at
	`, item.NS, item.Email, item.Role, item.InviteToken, item.InvitedByPrincipalID, item.ExpiresAt).Scan(
		&saved.ID,
		&saved.NS,
		&saved.Email,
		&saved.Role,
		&saved.InviteToken,
		&saved.InvitedByPrincipalID,
		&saved.AcceptedByPrincipalID,
		&saved.AcceptedAt,
		&saved.ExpiresAt,
		&saved.CreatedAt,
	)
	return saved, err
}

func (r *Repository) ListTenantInvites(ctx context.Context, tenantID string) ([]TenantInviteRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ns, email, role, invite_token, invited_by_principal_id, accepted_by_principal_id, accepted_at, expires_at, created_at
		FROM ns_invites
		WHERE ns = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantInviteRecord, 0)
	for rows.Next() {
		var item TenantInviteRecord
		if err := rows.Scan(
			&item.ID,
			&item.NS,
			&item.Email,
			&item.Role,
			&item.InviteToken,
			&item.InvitedByPrincipalID,
			&item.AcceptedByPrincipalID,
			&item.AcceptedAt,
			&item.ExpiresAt,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetTenantInviteByToken(ctx context.Context, token string) (TenantInviteRecord, error) {
	var item TenantInviteRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, ns, email, role, invite_token, invited_by_principal_id, accepted_by_principal_id, accepted_at, expires_at, created_at
		FROM ns_invites
		WHERE invite_token = $1
	`, token).Scan(
		&item.ID,
		&item.NS,
		&item.Email,
		&item.Role,
		&item.InviteToken,
		&item.InvitedByPrincipalID,
		&item.AcceptedByPrincipalID,
		&item.AcceptedAt,
		&item.ExpiresAt,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantInviteRecord{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) AcceptTenantInvite(ctx context.Context, inviteID int64, principalID string) (TenantInviteRecord, error) {
	var item TenantInviteRecord
	err := r.pool.QueryRow(ctx, `
		UPDATE ns_invites
		SET accepted_by_principal_id = $1, accepted_at = NOW()
		WHERE id = $2 AND accepted_at IS NULL
		RETURNING id, ns, email, role, invite_token, invited_by_principal_id, accepted_by_principal_id, accepted_at, expires_at, created_at
	`, principalID, inviteID).Scan(
		&item.ID,
		&item.NS,
		&item.Email,
		&item.Role,
		&item.InviteToken,
		&item.InvitedByPrincipalID,
		&item.AcceptedByPrincipalID,
		&item.AcceptedAt,
		&item.ExpiresAt,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantInviteRecord{}, ErrConflict
	}
	return item, err
}
