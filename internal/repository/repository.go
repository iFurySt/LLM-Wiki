package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type Repository struct {
	pool *pgxpool.Pool
}

type Space struct {
	ID          int64
	TenantID    string
	Key         string
	DisplayName string
	CreatedAt   time.Time
}

type Namespace struct {
	ID          int64
	TenantID    string
	SpaceID     int64
	Key         string
	DisplayName string
	Description string
	Visibility  string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Document struct {
	ID                int64
	TenantID          string
	NamespaceID       int64
	Slug              string
	Title             string
	Content           string
	Status            string
	CurrentRevisionID int64
	CurrentRevisionNo int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Revision struct {
	ID            int64
	DocumentID    int64
	RevisionNo    int32
	Title         string
	Content       string
	AuthorType    string
	AuthorID      string
	ChangeSummary string
	CreatedAt     time.Time
}

type CreateNamespaceParams struct {
	TenantID    string
	Key         string
	DisplayName string
	Description string
	Visibility  string
}

type CreateDocumentParams struct {
	TenantID      string
	NamespaceID   int64
	Slug          string
	Title         string
	Content       string
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

type UpdateDocumentParams struct {
	TenantID      string
	DocumentID    int64
	Title         string
	Content       string
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

type ArchiveDocumentParams struct {
	TenantID      string
	DocumentID    int64
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *Repository) EnsureTenantSpace(ctx context.Context, tenantID string) (Space, error) {
	var space Space
	err := r.pool.QueryRow(ctx, `
		INSERT INTO spaces (tenant_id, key, display_name)
		VALUES ($1, 'default', 'Default Space')
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING id, tenant_id, key, display_name, created_at
	`, tenantID).Scan(
		&space.ID,
		&space.TenantID,
		&space.Key,
		&space.DisplayName,
		&space.CreatedAt,
	)
	return space, err
}

func (r *Repository) ListSpaces(ctx context.Context, tenantID string) ([]Space, error) {
	space, err := r.EnsureTenantSpace(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return []Space{space}, nil
}

func (r *Repository) CreateNamespace(ctx context.Context, params CreateNamespaceParams) (Namespace, error) {
	space, err := r.EnsureTenantSpace(ctx, params.TenantID)
	if err != nil {
		return Namespace{}, err
	}

	var namespace Namespace
	err = r.pool.QueryRow(ctx, `
		INSERT INTO namespaces (
			tenant_id, space_id, key, display_name, description, visibility, status
		) VALUES ($1, $2, $3, $4, $5, $6, 'active')
		RETURNING id, tenant_id, space_id, key, display_name, description, visibility, status, created_at, updated_at
	`,
		params.TenantID,
		space.ID,
		params.Key,
		params.DisplayName,
		params.Description,
		params.Visibility,
	).Scan(
		&namespace.ID,
		&namespace.TenantID,
		&namespace.SpaceID,
		&namespace.Key,
		&namespace.DisplayName,
		&namespace.Description,
		&namespace.Visibility,
		&namespace.Status,
		&namespace.CreatedAt,
		&namespace.UpdatedAt,
	)
	return namespace, err
}

func (r *Repository) GetNamespace(ctx context.Context, tenantID string, namespaceID int64) (Namespace, error) {
	var namespace Namespace
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, space_id, key, display_name, description, visibility, status, created_at, updated_at
		FROM namespaces
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, namespaceID).Scan(
		&namespace.ID,
		&namespace.TenantID,
		&namespace.SpaceID,
		&namespace.Key,
		&namespace.DisplayName,
		&namespace.Description,
		&namespace.Visibility,
		&namespace.Status,
		&namespace.CreatedAt,
		&namespace.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Namespace{}, ErrNotFound
	}
	return namespace, err
}

func (r *Repository) ListNamespaces(ctx context.Context, tenantID string) ([]Namespace, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, space_id, key, display_name, description, visibility, status, created_at, updated_at
		FROM namespaces
		WHERE tenant_id = $1
		ORDER BY key ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Namespace, 0)
	for rows.Next() {
		var item Namespace
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.SpaceID,
			&item.Key,
			&item.DisplayName,
			&item.Description,
			&item.Visibility,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ArchiveNamespace(ctx context.Context, tenantID string, namespaceID int64) (Namespace, error) {
	var activeDocs int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM documents
		WHERE tenant_id = $1 AND namespace_id = $2 AND status = 'active'
	`, tenantID, namespaceID).Scan(&activeDocs); err != nil {
		return Namespace{}, err
	}
	if activeDocs > 0 {
		return Namespace{}, ErrConflict
	}

	var namespace Namespace
	err := r.pool.QueryRow(ctx, `
		UPDATE namespaces
		SET status = 'archived', updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
		RETURNING id, tenant_id, space_id, key, display_name, description, visibility, status, created_at, updated_at
	`, tenantID, namespaceID).Scan(
		&namespace.ID,
		&namespace.TenantID,
		&namespace.SpaceID,
		&namespace.Key,
		&namespace.DisplayName,
		&namespace.Description,
		&namespace.Visibility,
		&namespace.Status,
		&namespace.CreatedAt,
		&namespace.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Namespace{}, ErrNotFound
	}
	return namespace, err
}

func (r *Repository) CreateDocument(ctx context.Context, params CreateDocumentParams) (Document, Revision, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Document{}, Revision{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var namespaceExists bool
	if err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM namespaces WHERE tenant_id = $1 AND id = $2
		)
	`, params.TenantID, params.NamespaceID).Scan(&namespaceExists); err != nil {
		return Document{}, Revision{}, err
	}
	if !namespaceExists {
		return Document{}, Revision{}, ErrNotFound
	}

	var documentID int64
	var createdAt time.Time
	if err = tx.QueryRow(ctx, `
		INSERT INTO documents (
			tenant_id, namespace_id, slug, title, content_md, status
		) VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, created_at
	`,
		params.TenantID,
		params.NamespaceID,
		params.Slug,
		params.Title,
		params.Content,
	).Scan(&documentID, &createdAt); err != nil {
		return Document{}, Revision{}, err
	}

	revision, err := insertRevision(ctx, tx, documentID, params.Title, params.Content, params.AuthorType, params.AuthorID, params.ChangeSummary, 1)
	if err != nil {
		return Document{}, Revision{}, err
	}

	var document Document
	if err = tx.QueryRow(ctx, `
		UPDATE documents
		SET current_revision_id = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, tenant_id, namespace_id, slug, title, content_md, status, current_revision_id, created_at, updated_at
	`, revision.ID, documentID).Scan(
		&document.ID,
		&document.TenantID,
		&document.NamespaceID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Status,
		&document.CurrentRevisionID,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return Document{}, Revision{}, err
	}
	document.CurrentRevisionNo = revision.RevisionNo

	if err = tx.Commit(ctx); err != nil {
		return Document{}, Revision{}, err
	}
	return document, revision, nil
}

func (r *Repository) UpdateDocument(ctx context.Context, params UpdateDocumentParams) (Document, Revision, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Document{}, Revision{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var existing Document
	err = tx.QueryRow(ctx, `
		SELECT id, tenant_id, namespace_id, slug, title, content_md, status, current_revision_id, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND id = $2
	`, params.TenantID, params.DocumentID).Scan(
		&existing.ID,
		&existing.TenantID,
		&existing.NamespaceID,
		&existing.Slug,
		&existing.Title,
		&existing.Content,
		&existing.Status,
		&existing.CurrentRevisionID,
		&existing.CreatedAt,
		&existing.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, Revision{}, ErrNotFound
	}
	if err != nil {
		return Document{}, Revision{}, err
	}

	var nextRevisionNo int32
	if err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_no), 0) + 1
		FROM revisions
		WHERE document_id = $1
	`, params.DocumentID).Scan(&nextRevisionNo); err != nil {
		return Document{}, Revision{}, err
	}

	revision, err := insertRevision(
		ctx,
		tx,
		params.DocumentID,
		params.Title,
		params.Content,
		params.AuthorType,
		params.AuthorID,
		params.ChangeSummary,
		nextRevisionNo,
	)
	if err != nil {
		return Document{}, Revision{}, err
	}

	var document Document
	if err = tx.QueryRow(ctx, `
		UPDATE documents
		SET title = $1, content_md = $2, current_revision_id = $3, updated_at = NOW()
		WHERE tenant_id = $4 AND id = $5
		RETURNING id, tenant_id, namespace_id, slug, title, content_md, status, current_revision_id, created_at, updated_at
	`, params.Title, params.Content, revision.ID, params.TenantID, params.DocumentID).Scan(
		&document.ID,
		&document.TenantID,
		&document.NamespaceID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Status,
		&document.CurrentRevisionID,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return Document{}, Revision{}, err
	}
	document.CurrentRevisionNo = revision.RevisionNo

	if err = tx.Commit(ctx); err != nil {
		return Document{}, Revision{}, err
	}
	return document, revision, nil
}

func (r *Repository) GetDocument(ctx context.Context, tenantID string, documentID int64) (Document, []Revision, error) {
	var document Document
	err := r.pool.QueryRow(ctx, `
		SELECT
			d.id,
			d.tenant_id,
			d.namespace_id,
			d.slug,
			d.title,
			d.content_md,
			d.status,
			d.current_revision_id,
			COALESCE(r.revision_no, 0),
			d.created_at,
			d.updated_at
		FROM documents d
		LEFT JOIN revisions r ON r.id = d.current_revision_id
		WHERE d.tenant_id = $1 AND d.id = $2
	`, tenantID, documentID).Scan(
		&document.ID,
		&document.TenantID,
		&document.NamespaceID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Status,
		&document.CurrentRevisionID,
		&document.CurrentRevisionNo,
		&document.CreatedAt,
		&document.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, nil, ErrNotFound
	}
	if err != nil {
		return Document{}, nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, document_id, revision_no, title, content_md, author_type, author_id, change_summary, created_at
		FROM revisions
		WHERE document_id = $1
		ORDER BY revision_no DESC
	`, documentID)
	if err != nil {
		return Document{}, nil, err
	}
	defer rows.Close()

	revisions := make([]Revision, 0)
	for rows.Next() {
		var revision Revision
		if err := rows.Scan(
			&revision.ID,
			&revision.DocumentID,
			&revision.RevisionNo,
			&revision.Title,
			&revision.Content,
			&revision.AuthorType,
			&revision.AuthorID,
			&revision.ChangeSummary,
			&revision.CreatedAt,
		); err != nil {
			return Document{}, nil, err
		}
		revisions = append(revisions, revision)
	}
	if err := rows.Err(); err != nil {
		return Document{}, nil, err
	}

	return document, revisions, nil
}

func (r *Repository) GetDocumentBySlug(ctx context.Context, tenantID string, namespaceID int64, slug string) (Document, []Revision, error) {
	var documentID int64
	err := r.pool.QueryRow(ctx, `
		SELECT id
		FROM documents
		WHERE tenant_id = $1 AND namespace_id = $2 AND slug = $3
	`, tenantID, namespaceID, slug).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, nil, ErrNotFound
	}
	if err != nil {
		return Document{}, nil, err
	}
	return r.GetDocument(ctx, tenantID, documentID)
}

func (r *Repository) ListDocuments(ctx context.Context, tenantID string, namespaceID *int64, status *string) ([]Document, error) {
	query := `
		SELECT d.id, d.tenant_id, d.namespace_id, d.slug, d.title, d.content_md, d.status, d.current_revision_id,
		       COALESCE(r.revision_no, 0), d.created_at, d.updated_at
		FROM documents d
		LEFT JOIN revisions r ON r.id = d.current_revision_id
		WHERE d.tenant_id = $1
	`
	args := []any{tenantID}
	nextArg := 2
	if namespaceID != nil {
		query += fmt.Sprintf(` AND d.namespace_id = $%d`, nextArg)
		args = append(args, *namespaceID)
		nextArg++
	}
	if status != nil && *status != "" {
		query += fmt.Sprintf(` AND d.status = $%d`, nextArg)
		args = append(args, *status)
	}
	query += ` ORDER BY d.updated_at DESC, d.id DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Document, 0)
	for rows.Next() {
		var item Document
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.NamespaceID,
			&item.Slug,
			&item.Title,
			&item.Content,
			&item.Status,
			&item.CurrentRevisionID,
			&item.CurrentRevisionNo,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) ArchiveDocument(ctx context.Context, params ArchiveDocumentParams) (Document, Revision, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Document{}, Revision{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var current Document
	err = tx.QueryRow(ctx, `
		SELECT id, tenant_id, namespace_id, slug, title, content_md, status, current_revision_id, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND id = $2
	`, params.TenantID, params.DocumentID).Scan(
		&current.ID,
		&current.TenantID,
		&current.NamespaceID,
		&current.Slug,
		&current.Title,
		&current.Content,
		&current.Status,
		&current.CurrentRevisionID,
		&current.CreatedAt,
		&current.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Document{}, Revision{}, ErrNotFound
	}
	if err != nil {
		return Document{}, Revision{}, err
	}

	var nextRevisionNo int32
	if err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(revision_no), 0) + 1
		FROM revisions
		WHERE document_id = $1
	`, params.DocumentID).Scan(&nextRevisionNo); err != nil {
		return Document{}, Revision{}, err
	}

	revision, err := insertRevision(
		ctx,
		tx,
		params.DocumentID,
		current.Title,
		current.Content,
		params.AuthorType,
		params.AuthorID,
		params.ChangeSummary,
		nextRevisionNo,
	)
	if err != nil {
		return Document{}, Revision{}, err
	}

	var document Document
	if err = tx.QueryRow(ctx, `
		UPDATE documents
		SET status = 'archived', current_revision_id = $1, updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3
		RETURNING id, tenant_id, namespace_id, slug, title, content_md, status, current_revision_id, created_at, updated_at
	`, revision.ID, params.TenantID, params.DocumentID).Scan(
		&document.ID,
		&document.TenantID,
		&document.NamespaceID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Status,
		&document.CurrentRevisionID,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return Document{}, Revision{}, err
	}
	document.CurrentRevisionNo = revision.RevisionNo

	if err = tx.Commit(ctx); err != nil {
		return Document{}, Revision{}, err
	}
	return document, revision, nil
}

func insertRevision(
	ctx context.Context,
	tx pgx.Tx,
	documentID int64,
	title string,
	content string,
	authorType string,
	authorID string,
	changeSummary string,
	revisionNo int32,
) (Revision, error) {
	var revision Revision
	err := tx.QueryRow(ctx, `
		INSERT INTO revisions (
			document_id, revision_no, title, content_md, author_type, author_id, change_summary
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, document_id, revision_no, title, content_md, author_type, author_id, change_summary, created_at
	`, documentID, revisionNo, title, content, authorType, authorID, changeSummary).Scan(
		&revision.ID,
		&revision.DocumentID,
		&revision.RevisionNo,
		&revision.Title,
		&revision.Content,
		&revision.AuthorType,
		&revision.AuthorID,
		&revision.ChangeSummary,
		&revision.CreatedAt,
	)
	return revision, err
}

func (r *Repository) String() string {
	return fmt.Sprintf("repository(%p)", r.pool)
}
