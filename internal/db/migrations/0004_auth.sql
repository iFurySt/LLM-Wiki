CREATE TABLE IF NOT EXISTS principals (
    id TEXT PRIMARY KEY,
    ns TEXT NOT NULL,
    principal_type TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ns, principal_type, display_name)
);

CREATE TABLE IF NOT EXISTS api_tokens (
    id BIGSERIAL PRIMARY KEY,
    token_type TEXT NOT NULL,
    principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
    ns TEXT NOT NULL,
    display_name TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    token_prefix TEXT NOT NULL,
    scopes_json TEXT NOT NULL DEFAULT '[]',
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_by_principal_id TEXT REFERENCES principals(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_requests (
    id TEXT PRIMARY KEY,
    flow_type TEXT NOT NULL,
    ns TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    scopes_json TEXT NOT NULL DEFAULT '[]',
    state TEXT NOT NULL DEFAULT '',
    redirect_uri TEXT NOT NULL DEFAULT '',
    code_challenge TEXT NOT NULL DEFAULT '',
    code_challenge_method TEXT NOT NULL DEFAULT '',
    auth_code TEXT NOT NULL DEFAULT '',
    device_code TEXT NOT NULL DEFAULT '',
    user_code TEXT NOT NULL DEFAULT '',
    principal_id TEXT REFERENCES principals(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    denied_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_tokens_prefix ON api_tokens (token_prefix);
CREATE INDEX IF NOT EXISTS idx_api_tokens_principal ON api_tokens (principal_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_requests_auth_code ON auth_requests (auth_code) WHERE auth_code <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_requests_device_code ON auth_requests (device_code) WHERE device_code <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_auth_requests_user_code ON auth_requests (user_code) WHERE user_code <> '';
