CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    principal_id TEXT NOT NULL UNIQUE REFERENCES principals(id) ON DELETE CASCADE,
    ns TEXT NOT NULL,
    username TEXT NOT NULL,
    display_name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ns, username)
);

CREATE TABLE IF NOT EXISTS web_sessions (
    id TEXT PRIMARY KEY,
    principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
    ns TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_admin ON users (ns, is_admin);
CREATE INDEX IF NOT EXISTS idx_web_sessions_principal ON web_sessions (principal_id, expires_at DESC);
