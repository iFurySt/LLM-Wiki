package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/jackc/pgx/v5"
)

var ErrUnauthorized = errors.New("unauthorized")
var ErrForbidden = errors.New("forbidden")
var ErrExpired = errors.New("expired")
var ErrPending = errors.New("pending")

type PrincipalRecord struct {
	ID            string
	TenantID      string
	PrincipalType string
	DisplayName   string
	CreatedAt     time.Time
}

type TokenRecord struct {
	ID                   int64
	TokenType            string
	PrincipalID          string
	TenantID             string
	DisplayName          string
	TokenPrefix          string
	Scopes               []string
	ExpiresAt            *time.Time
	LastUsedAt           *time.Time
	RevokedAt            *time.Time
	CreatedByPrincipalID *string
	CreatedAt            time.Time
}

type AuthRequest struct {
	ID                  string
	FlowType            string
	TenantID            string
	OAuthProvider       string
	DisplayName         string
	Scopes              []string
	State               string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	AuthCode            string
	DeviceCode          string
	UserCode            string
	PrincipalID         *string
	ApprovedAt          *time.Time
	DeniedAt            *time.Time
	ExpiresAt           time.Time
	CreatedAt           time.Time
}

type AuthenticatedPrincipal struct {
	Principal PrincipalRecord
	Token     TokenRecord
}

type IssueTokenParams struct {
	TokenType            string
	PrincipalID          string
	TenantID             string
	DisplayName          string
	Scopes               []string
	ExpiresAt            *time.Time
	CreatedByPrincipalID *string
}

type CreateAuthRequestParams struct {
	FlowType            string
	TenantID            string
	OAuthProvider       string
	DisplayName         string
	Scopes              []string
	State               string
	RedirectURI         string
	CodeChallenge       string
	CodeChallengeMethod string
	DeviceCode          string
	UserCode            string
	ExpiresAt           time.Time
}

func (r *Repository) EnsurePrincipal(ctx context.Context, tenantID string, principalType string, displayName string) (PrincipalRecord, error) {
	var record PrincipalRecord
	id, err := auth.RandomID("prn", 16)
	if err != nil {
		return PrincipalRecord{}, err
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO principals (id, tenant_id, principal_type, display_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, principal_type, display_name)
		DO UPDATE SET display_name = EXCLUDED.display_name
		RETURNING id, tenant_id, principal_type, display_name, created_at
	`, id, tenantID, principalType, displayName).Scan(
		&record.ID,
		&record.TenantID,
		&record.PrincipalType,
		&record.DisplayName,
		&record.CreatedAt,
	)
	return record, err
}

func (r *Repository) ListPrincipalsByType(ctx context.Context, tenantID string, principalType string) ([]PrincipalRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, principal_type, display_name, created_at
		FROM principals
		WHERE tenant_id = $1 AND principal_type = $2
		ORDER BY display_name ASC
	`, tenantID, principalType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PrincipalRecord, 0)
	for rows.Next() {
		var item PrincipalRecord
		if err := rows.Scan(&item.ID, &item.TenantID, &item.PrincipalType, &item.DisplayName, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetPrincipal(ctx context.Context, principalID string) (PrincipalRecord, error) {
	var record PrincipalRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, principal_type, display_name, created_at
		FROM principals
		WHERE id = $1
	`, principalID).Scan(
		&record.ID,
		&record.TenantID,
		&record.PrincipalType,
		&record.DisplayName,
		&record.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return PrincipalRecord{}, ErrNotFound
	}
	return record, err
}

func (r *Repository) IssueToken(ctx context.Context, params IssueTokenParams) (TokenRecord, string, error) {
	rawToken, prefix, err := auth.NewOpaqueToken("lw")
	if err != nil {
		return TokenRecord{}, "", err
	}
	record, err := r.insertToken(ctx, params, rawToken, prefix)
	return record, rawToken, err
}

func (r *Repository) UpsertStaticToken(ctx context.Context, params IssueTokenParams, rawToken string) (TokenRecord, error) {
	prefix := rawToken
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	record, err := r.insertToken(ctx, params, rawToken, prefix)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			return TokenRecord{}, err
		}
	}
	if err == nil {
		return record, nil
	}
	var existing TokenRecord
	var scopesJSON string
	err = r.pool.QueryRow(ctx, `
		SELECT id, token_type, principal_id, tenant_id, display_name, token_prefix, scopes_json, expires_at, last_used_at, revoked_at, created_by_principal_id, created_at
		FROM api_tokens
		WHERE token_hash = $1
	`, auth.HashToken(rawToken)).Scan(
		&existing.ID,
		&existing.TokenType,
		&existing.PrincipalID,
		&existing.TenantID,
		&existing.DisplayName,
		&existing.TokenPrefix,
		&scopesJSON,
		&existing.ExpiresAt,
		&existing.LastUsedAt,
		&existing.RevokedAt,
		&existing.CreatedByPrincipalID,
		&existing.CreatedAt,
	)
	if err != nil {
		return TokenRecord{}, err
	}
	existing.Scopes, err = unmarshalScopes(scopesJSON)
	return existing, err
}

func (r *Repository) insertToken(ctx context.Context, params IssueTokenParams, rawToken string, prefix string) (TokenRecord, error) {
	hash := auth.HashToken(rawToken)
	scopesJSON, err := marshalScopes(params.Scopes)
	if err != nil {
		return TokenRecord{}, err
	}

	var record TokenRecord
	err = r.pool.QueryRow(ctx, `
		INSERT INTO api_tokens (
			token_type, principal_id, tenant_id, display_name, token_hash, token_prefix, scopes_json, expires_at, created_by_principal_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, token_type, principal_id, tenant_id, display_name, token_prefix, scopes_json, expires_at, last_used_at, revoked_at, created_by_principal_id, created_at
	`, params.TokenType, params.PrincipalID, params.TenantID, params.DisplayName, hash, prefix, scopesJSON, params.ExpiresAt, params.CreatedByPrincipalID).Scan(
		&record.ID,
		&record.TokenType,
		&record.PrincipalID,
		&record.TenantID,
		&record.DisplayName,
		&record.TokenPrefix,
		&scopesJSON,
		&record.ExpiresAt,
		&record.LastUsedAt,
		&record.RevokedAt,
		&record.CreatedByPrincipalID,
		&record.CreatedAt,
	)
	if err != nil {
		return TokenRecord{}, err
	}
	record.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return TokenRecord{}, err
	}
	return record, nil
}

func (r *Repository) ListTokens(ctx context.Context, tenantID string) ([]TokenRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, token_type, principal_id, tenant_id, display_name, token_prefix, scopes_json, expires_at, last_used_at, revoked_at, created_by_principal_id, created_at
		FROM api_tokens
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TokenRecord, 0)
	for rows.Next() {
		var item TokenRecord
		var scopesJSON string
		if err := rows.Scan(
			&item.ID,
			&item.TokenType,
			&item.PrincipalID,
			&item.TenantID,
			&item.DisplayName,
			&item.TokenPrefix,
			&scopesJSON,
			&item.ExpiresAt,
			&item.LastUsedAt,
			&item.RevokedAt,
			&item.CreatedByPrincipalID,
			&item.CreatedAt,
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

func (r *Repository) RevokeToken(ctx context.Context, tenantID string, tokenID int64) (TokenRecord, error) {
	var item TokenRecord
	var scopesJSON string
	err := r.pool.QueryRow(ctx, `
		UPDATE api_tokens
		SET revoked_at = NOW()
		WHERE tenant_id = $1 AND id = $2
		RETURNING id, token_type, principal_id, tenant_id, display_name, token_prefix, scopes_json, expires_at, last_used_at, revoked_at, created_by_principal_id, created_at
	`, tenantID, tokenID).Scan(
		&item.ID,
		&item.TokenType,
		&item.PrincipalID,
		&item.TenantID,
		&item.DisplayName,
		&item.TokenPrefix,
		&scopesJSON,
		&item.ExpiresAt,
		&item.LastUsedAt,
		&item.RevokedAt,
		&item.CreatedByPrincipalID,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return TokenRecord{}, ErrNotFound
	}
	if err != nil {
		return TokenRecord{}, err
	}
	item.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return TokenRecord{}, err
	}
	return item, nil
}

func (r *Repository) AuthenticateToken(ctx context.Context, rawToken string) (AuthenticatedPrincipal, error) {
	return r.authenticateTokenByType(ctx, rawToken, "access", "service", "personal")
}

func (r *Repository) AuthenticateRefreshToken(ctx context.Context, rawToken string) (AuthenticatedPrincipal, error) {
	return r.authenticateTokenByType(ctx, rawToken, "refresh")
}

func (r *Repository) authenticateTokenByType(ctx context.Context, rawToken string, allowedTypes ...string) (AuthenticatedPrincipal, error) {
	hash := auth.HashToken(rawToken)
	var principal PrincipalRecord
	var token TokenRecord
	var scopesJSON string
	err := r.pool.QueryRow(ctx, `
		SELECT
			p.id,
			p.tenant_id,
			p.principal_type,
			p.display_name,
			p.created_at,
			t.id,
			t.token_type,
			t.principal_id,
			t.tenant_id,
			t.display_name,
			t.token_prefix,
			t.scopes_json,
			t.expires_at,
			t.last_used_at,
			t.revoked_at,
			t.created_by_principal_id,
			t.created_at
		FROM api_tokens t
		JOIN principals p ON p.id = t.principal_id
		WHERE t.token_hash = $1
	`, hash).Scan(
		&principal.ID,
		&principal.TenantID,
		&principal.PrincipalType,
		&principal.DisplayName,
		&principal.CreatedAt,
		&token.ID,
		&token.TokenType,
		&token.PrincipalID,
		&token.TenantID,
		&token.DisplayName,
		&token.TokenPrefix,
		&scopesJSON,
		&token.ExpiresAt,
		&token.LastUsedAt,
		&token.RevokedAt,
		&token.CreatedByPrincipalID,
		&token.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthenticatedPrincipal{}, ErrUnauthorized
	}
	if err != nil {
		return AuthenticatedPrincipal{}, err
	}
	if token.RevokedAt != nil {
		return AuthenticatedPrincipal{}, ErrUnauthorized
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return AuthenticatedPrincipal{}, ErrExpired
	}
	allowed := false
	for _, item := range allowedTypes {
		if token.TokenType == item {
			allowed = true
			break
		}
	}
	if !allowed {
		return AuthenticatedPrincipal{}, ErrUnauthorized
	}
	token.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return AuthenticatedPrincipal{}, err
	}
	_, _ = r.pool.Exec(ctx, `UPDATE api_tokens SET last_used_at = NOW() WHERE id = $1`, token.ID)
	return AuthenticatedPrincipal{Principal: principal, Token: token}, nil
}

func (r *Repository) CreateAuthRequest(ctx context.Context, params CreateAuthRequestParams) (AuthRequest, error) {
	id, err := auth.RandomID("req", 16)
	if err != nil {
		return AuthRequest{}, err
	}
	scopesJSON, err := marshalScopes(params.Scopes)
	if err != nil {
		return AuthRequest{}, err
	}
	var item AuthRequest
	err = r.pool.QueryRow(ctx, `
		INSERT INTO auth_requests (
			id, flow_type, tenant_id, oauth_provider, display_name, scopes_json, state, redirect_uri, code_challenge, code_challenge_method, device_code, user_code, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, flow_type, tenant_id, oauth_provider, display_name, scopes_json, state, redirect_uri, code_challenge, code_challenge_method, auth_code, device_code, user_code, principal_id, approved_at, denied_at, expires_at, created_at
	`, id, params.FlowType, params.TenantID, params.OAuthProvider, params.DisplayName, scopesJSON, params.State, params.RedirectURI, params.CodeChallenge, params.CodeChallengeMethod, params.DeviceCode, params.UserCode, params.ExpiresAt).Scan(
		&item.ID,
		&item.FlowType,
		&item.TenantID,
		&item.OAuthProvider,
		&item.DisplayName,
		&scopesJSON,
		&item.State,
		&item.RedirectURI,
		&item.CodeChallenge,
		&item.CodeChallengeMethod,
		&item.AuthCode,
		&item.DeviceCode,
		&item.UserCode,
		&item.PrincipalID,
		&item.ApprovedAt,
		&item.DeniedAt,
		&item.ExpiresAt,
		&item.CreatedAt,
	)
	if err != nil {
		return AuthRequest{}, err
	}
	item.Scopes, err = unmarshalScopes(scopesJSON)
	return item, err
}

func (r *Repository) GetAuthRequest(ctx context.Context, requestID string) (AuthRequest, error) {
	return r.getAuthRequestBy(ctx, "id", requestID)
}

func (r *Repository) GetAuthRequestByUserCode(ctx context.Context, userCode string) (AuthRequest, error) {
	return r.getAuthRequestBy(ctx, "user_code", userCode)
}

func (r *Repository) GetAuthRequestByDeviceCode(ctx context.Context, deviceCode string) (AuthRequest, error) {
	return r.getAuthRequestBy(ctx, "device_code", deviceCode)
}

func (r *Repository) ExchangeAuthorizationCode(ctx context.Context, authCode string) (AuthRequest, error) {
	return r.getAuthRequestBy(ctx, "auth_code", authCode)
}

func (r *Repository) getAuthRequestBy(ctx context.Context, column string, value string) (AuthRequest, error) {
	query := fmt.Sprintf(`
		SELECT id, flow_type, tenant_id, oauth_provider, display_name, scopes_json, state, redirect_uri, code_challenge, code_challenge_method, auth_code, device_code, user_code, principal_id, approved_at, denied_at, expires_at, created_at
		FROM auth_requests
		WHERE %s = $1
	`, column)
	var item AuthRequest
	var scopesJSON string
	err := r.pool.QueryRow(ctx, query, value).Scan(
		&item.ID,
		&item.FlowType,
		&item.TenantID,
		&item.OAuthProvider,
		&item.DisplayName,
		&scopesJSON,
		&item.State,
		&item.RedirectURI,
		&item.CodeChallenge,
		&item.CodeChallengeMethod,
		&item.AuthCode,
		&item.DeviceCode,
		&item.UserCode,
		&item.PrincipalID,
		&item.ApprovedAt,
		&item.DeniedAt,
		&item.ExpiresAt,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthRequest{}, ErrNotFound
	}
	if err != nil {
		return AuthRequest{}, err
	}
	item.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return AuthRequest{}, err
	}
	return item, nil
}

func (r *Repository) ApproveAuthRequest(ctx context.Context, requestID string, displayName string) (AuthRequest, PrincipalRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var request AuthRequest
	var scopesJSON string
	err = tx.QueryRow(ctx, `
		SELECT id, flow_type, tenant_id, oauth_provider, display_name, scopes_json, state, redirect_uri, code_challenge, code_challenge_method, auth_code, device_code, user_code, principal_id, approved_at, denied_at, expires_at, created_at
		FROM auth_requests
		WHERE id = $1
		FOR UPDATE
	`, requestID).Scan(
		&request.ID,
		&request.FlowType,
		&request.TenantID,
		&request.OAuthProvider,
		&request.DisplayName,
		&scopesJSON,
		&request.State,
		&request.RedirectURI,
		&request.CodeChallenge,
		&request.CodeChallengeMethod,
		&request.AuthCode,
		&request.DeviceCode,
		&request.UserCode,
		&request.PrincipalID,
		&request.ApprovedAt,
		&request.DeniedAt,
		&request.ExpiresAt,
		&request.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthRequest{}, PrincipalRecord{}, ErrNotFound
	}
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	request.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	if request.ExpiresAt.Before(time.Now()) {
		return AuthRequest{}, PrincipalRecord{}, ErrExpired
	}

	principalID, err := auth.RandomID("prn", 16)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	var principal PrincipalRecord
	err = tx.QueryRow(ctx, `
		INSERT INTO principals (id, tenant_id, principal_type, display_name)
		VALUES ($1, $2, 'user', $3)
		ON CONFLICT (tenant_id, principal_type, display_name)
		DO UPDATE SET display_name = EXCLUDED.display_name
		RETURNING id, tenant_id, principal_type, display_name, created_at
	`, principalID, request.TenantID, displayName).Scan(
		&principal.ID,
		&principal.TenantID,
		&principal.PrincipalType,
		&principal.DisplayName,
		&principal.CreatedAt,
	)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}

	authCode, err := auth.RandomID("code", 16)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	err = tx.QueryRow(ctx, `
		UPDATE auth_requests
		SET principal_id = $1, display_name = $2, auth_code = $3, approved_at = NOW()
		WHERE id = $4
		RETURNING id, flow_type, tenant_id, oauth_provider, display_name, scopes_json, state, redirect_uri, code_challenge, code_challenge_method, auth_code, device_code, user_code, principal_id, approved_at, denied_at, expires_at, created_at
	`, principal.ID, displayName, authCode, request.ID).Scan(
		&request.ID,
		&request.FlowType,
		&request.TenantID,
		&request.OAuthProvider,
		&request.DisplayName,
		&scopesJSON,
		&request.State,
		&request.RedirectURI,
		&request.CodeChallenge,
		&request.CodeChallengeMethod,
		&request.AuthCode,
		&request.DeviceCode,
		&request.UserCode,
		&request.PrincipalID,
		&request.ApprovedAt,
		&request.DeniedAt,
		&request.ExpiresAt,
		&request.CreatedAt,
	)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	request.Scopes, err = unmarshalScopes(scopesJSON)
	if err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return AuthRequest{}, PrincipalRecord{}, err
	}
	return request, principal, nil
}

func (r *Repository) ApproveAuthRequestForPrincipal(ctx context.Context, requestID string, principalID string) (AuthRequest, error) {
	var request AuthRequest
	var scopesJSON string
	authCode, err := auth.RandomID("code", 16)
	if err != nil {
		return AuthRequest{}, err
	}
	err = r.pool.QueryRow(ctx, `
		UPDATE auth_requests
		SET principal_id = $1,
			auth_code = $2,
			approved_at = NOW(),
			tenant_id = p.tenant_id,
			display_name = p.display_name
		FROM principals p
		WHERE auth_requests.id = $3
		  AND p.id = $1
		RETURNING
			auth_requests.id,
			auth_requests.flow_type,
			auth_requests.tenant_id,
			auth_requests.oauth_provider,
			auth_requests.display_name,
			auth_requests.scopes_json,
			auth_requests.state,
			auth_requests.redirect_uri,
			auth_requests.code_challenge,
			auth_requests.code_challenge_method,
			auth_requests.auth_code,
			auth_requests.device_code,
			auth_requests.user_code,
			auth_requests.principal_id,
			auth_requests.approved_at,
			auth_requests.denied_at,
			auth_requests.expires_at,
			auth_requests.created_at
	`, principalID, authCode, requestID).Scan(
		&request.ID,
		&request.FlowType,
		&request.TenantID,
		&request.OAuthProvider,
		&request.DisplayName,
		&scopesJSON,
		&request.State,
		&request.RedirectURI,
		&request.CodeChallenge,
		&request.CodeChallengeMethod,
		&request.AuthCode,
		&request.DeviceCode,
		&request.UserCode,
		&request.PrincipalID,
		&request.ApprovedAt,
		&request.DeniedAt,
		&request.ExpiresAt,
		&request.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthRequest{}, ErrNotFound
	}
	if err != nil {
		return AuthRequest{}, err
	}
	request.Scopes, err = unmarshalScopes(scopesJSON)
	return request, err
}

func marshalScopes(scopes []string) (string, error) {
	if scopes == nil {
		scopes = []string{}
	}
	payload, err := json.Marshal(scopes)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func unmarshalScopes(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}
	var scopes []string
	if err := json.Unmarshal([]byte(raw), &scopes); err != nil {
		return nil, err
	}
	return scopes, nil
}
