ALTER TABLE auth_requests
ADD COLUMN IF NOT EXISTS oauth_provider TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS oauth_providers (
    name TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    auth_url TEXT NOT NULL,
    token_url TEXT NOT NULL,
    userinfo_url TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    scopes_json TEXT NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    auto_create_users BOOLEAN NOT NULL DEFAULT TRUE,
    auto_create_tenants BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS oauth_accounts (
    id BIGSERIAL PRIMARY KEY,
    provider_name TEXT NOT NULL REFERENCES oauth_providers(name) ON DELETE CASCADE,
    external_subject TEXT NOT NULL,
    principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
    ns TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_name, external_subject)
);

CREATE INDEX IF NOT EXISTS idx_oauth_providers_enabled ON oauth_providers (enabled);
CREATE INDEX IF NOT EXISTS idx_oauth_accounts_principal ON oauth_accounts (principal_id);
