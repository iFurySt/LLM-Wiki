package repository

import (
	"context"
	"errors"
	"time"

	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/jackc/pgx/v5"
)

type UserRecord struct {
	ID           int64
	PrincipalID  string
	TenantID     string
	Username     string
	DisplayName  string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
}

type WebSessionRecord struct {
	ID          string
	PrincipalID string
	TenantID    string
	ExpiresAt   time.Time
	CreatedAt   time.Time
}

func (r *Repository) AnyUsersExist(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users)`).Scan(&exists)
	return exists, err
}

func (r *Repository) AnyAdminUsersExist(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE is_admin = TRUE)`).Scan(&exists)
	return exists, err
}

func (r *Repository) CreateUser(ctx context.Context, tenantID string, username string, displayName string, passwordHash string, isAdmin bool) (UserRecord, error) {
	principal, err := r.EnsurePrincipal(ctx, tenantID, "user", displayName)
	if err != nil {
		return UserRecord{}, err
	}

	var item UserRecord
	err = r.pool.QueryRow(ctx, `
		INSERT INTO users (principal_id, tenant_id, username, display_name, password_hash, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, principal_id, tenant_id, username, display_name, password_hash, is_admin, created_at
	`, principal.ID, tenantID, username, displayName, passwordHash, isAdmin).Scan(
		&item.ID,
		&item.PrincipalID,
		&item.TenantID,
		&item.Username,
		&item.DisplayName,
		&item.PasswordHash,
		&item.IsAdmin,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) ListUsers(ctx context.Context, tenantID string) ([]UserRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, principal_id, tenant_id, username, display_name, password_hash, is_admin, created_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY username ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]UserRecord, 0)
	for rows.Next() {
		var item UserRecord
		if err := rows.Scan(&item.ID, &item.PrincipalID, &item.TenantID, &item.Username, &item.DisplayName, &item.PasswordHash, &item.IsAdmin, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetUserByUsername(ctx context.Context, tenantID string, username string) (UserRecord, error) {
	var item UserRecord
	err := r.pool.QueryRow(ctx, `
		SELECT id, principal_id, tenant_id, username, display_name, password_hash, is_admin, created_at
		FROM users
		WHERE tenant_id = $1 AND username = $2
	`, tenantID, username).Scan(
		&item.ID,
		&item.PrincipalID,
		&item.TenantID,
		&item.Username,
		&item.DisplayName,
		&item.PasswordHash,
		&item.IsAdmin,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) CreateWebSession(ctx context.Context, principalID string, tenantID string, ttl time.Duration) (WebSessionRecord, error) {
	sessionID, err := auth.RandomID("ws", 16)
	if err != nil {
		return WebSessionRecord{}, err
	}
	expiresAt := time.Now().Add(ttl)
	var item WebSessionRecord
	err = r.pool.QueryRow(ctx, `
		INSERT INTO web_sessions (id, principal_id, tenant_id, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, principal_id, tenant_id, expires_at, created_at
	`, sessionID, principalID, tenantID, expiresAt).Scan(
		&item.ID,
		&item.PrincipalID,
		&item.TenantID,
		&item.ExpiresAt,
		&item.CreatedAt,
	)
	return item, err
}

func (r *Repository) GetWebSession(ctx context.Context, sessionID string) (WebSessionRecord, UserRecord, error) {
	var session WebSessionRecord
	var user UserRecord
	err := r.pool.QueryRow(ctx, `
		SELECT
			s.id, s.principal_id, s.tenant_id, s.expires_at, s.created_at,
			u.id, u.principal_id, u.tenant_id, u.username, u.display_name, u.password_hash, u.is_admin, u.created_at
		FROM web_sessions s
		JOIN users u ON u.principal_id = s.principal_id
		WHERE s.id = $1
	`, sessionID).Scan(
		&session.ID,
		&session.PrincipalID,
		&session.TenantID,
		&session.ExpiresAt,
		&session.CreatedAt,
		&user.ID,
		&user.PrincipalID,
		&user.TenantID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return WebSessionRecord{}, UserRecord{}, ErrNotFound
	}
	return session, user, err
}

func (r *Repository) DeleteWebSession(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM web_sessions WHERE id = $1`, sessionID)
	return err
}
