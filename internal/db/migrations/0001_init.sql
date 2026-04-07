CREATE TABLE IF NOT EXISTS spaces (
    id BIGSERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL UNIQUE,
    key TEXT NOT NULL,
    display_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS namespaces (
    id BIGSERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    space_id BIGINT NOT NULL REFERENCES spaces(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    visibility TEXT NOT NULL DEFAULT 'private',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, key)
);

CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    namespace_id BIGINT NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    content_md TEXT NOT NULL DEFAULT '',
    current_revision_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (namespace_id, slug)
);

CREATE TABLE IF NOT EXISTS revisions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    revision_no INTEGER NOT NULL,
    title TEXT NOT NULL,
    content_md TEXT NOT NULL DEFAULT '',
    author_type TEXT NOT NULL DEFAULT 'agent',
    author_id TEXT NOT NULL DEFAULT '',
    change_summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, revision_no)
);

CREATE INDEX IF NOT EXISTS idx_namespaces_tenant_space ON namespaces (tenant_id, space_id);
CREATE INDEX IF NOT EXISTS idx_documents_tenant_namespace ON documents (tenant_id, namespace_id);
CREATE INDEX IF NOT EXISTS idx_revisions_document_created ON revisions (document_id, created_at DESC);
