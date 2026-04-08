ALTER TABLE oauth_accounts
DROP CONSTRAINT IF EXISTS oauth_accounts_provider_name_external_subject_key;

ALTER TABLE oauth_accounts
ADD CONSTRAINT oauth_accounts_provider_subject_tenant_key UNIQUE (provider_name, external_subject, ns);

CREATE TABLE IF NOT EXISTS ns_invites (
    id BIGSERIAL PRIMARY KEY,
    ns TEXT NOT NULL,
    email TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    invite_token TEXT NOT NULL UNIQUE,
    invited_by_principal_id TEXT REFERENCES principals(id) ON DELETE SET NULL,
    accepted_by_principal_id TEXT REFERENCES principals(id) ON DELETE SET NULL,
    accepted_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenant_invites_tenant_email ON ns_invites (ns, email);
