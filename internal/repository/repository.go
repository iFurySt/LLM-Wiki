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

type NS struct {
	ID          int64
	NS          string
	Key         string
	DisplayName string
	CreatedAt   time.Time
	Role        string
}

type Folder struct {
	ID          int64
	NS          string
	Key         string
	DisplayName string
	Description string
	Visibility  string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DocumentSource struct {
	ID          string
	Label       string
	Category    string
	InputMode   string
	OriginalRef string
	CapturedAt  *time.Time
	ContentType string
	Adapter     string
}

type Document struct {
	ID                int64
	NS                string
	FolderID          int64
	Slug              string
	Title             string
	Content           string
	Source            DocumentSource
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
	Source        DocumentSource
	AuthorType    string
	AuthorID      string
	ChangeSummary string
	CreatedAt     time.Time
}

type CreateFolderParams struct {
	NS          string
	Key         string
	DisplayName string
	Description string
	Visibility  string
}

type CreateNamespaceParams = CreateFolderParams

type CreateDocumentParams struct {
	NS            string
	FolderID      int64
	Slug          string
	Title         string
	Content       string
	Source        DocumentSource
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

type UpdateDocumentParams struct {
	NS            string
	DocumentID    int64
	Title         string
	Content       string
	AuthorType    string
	AuthorID      string
	ChangeSummary string
}

type ArchiveDocumentParams struct {
	NS            string
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

func (r *Repository) EnsureNS(ctx context.Context, ns string) (NS, error) {
	return r.EnsureNSWithDisplayName(ctx, ns, "Default NS")
}

func (r *Repository) EnsureTenantSpace(ctx context.Context, tenantID string) (NS, error) {
	return r.EnsureNS(ctx, tenantID)
}

func (r *Repository) EnsureNSWithDisplayName(ctx context.Context, ns string, displayName string) (NS, error) {
	var item NS
	err := r.pool.QueryRow(ctx, `
		INSERT INTO ns (ns, key, display_name)
		VALUES ($1, 'default', $2)
		ON CONFLICT (ns) DO UPDATE SET
			ns = EXCLUDED.ns,
			display_name = CASE
				WHEN ns.display_name = '' OR ns.display_name = 'Default NS' THEN EXCLUDED.display_name
				ELSE ns.display_name
			END
		RETURNING id, ns, key, display_name, created_at
	`, ns, displayName).Scan(
		&item.ID,
		&item.NS,
		&item.Key,
		&item.DisplayName,
		&item.CreatedAt,
	)
	return item, err
}

func (r *Repository) EnsureTenantSpaceWithDisplayName(ctx context.Context, tenantID string, displayName string) (NS, error) {
	return r.EnsureNSWithDisplayName(ctx, tenantID, displayName)
}

func (r *Repository) ListNS(ctx context.Context, ns string) ([]NS, error) {
	item, err := r.EnsureNS(ctx, ns)
	if err != nil {
		return nil, err
	}
	return []NS{item}, nil
}

func (r *Repository) CreateFolder(ctx context.Context, params CreateFolderParams) (Folder, error) {
	if _, err := r.EnsureNS(ctx, params.NS); err != nil {
		return Folder{}, err
	}

	var folder Folder
	err := r.pool.QueryRow(ctx, `
		INSERT INTO folders (
			ns, key, display_name, description, visibility, status
		) VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, ns, key, display_name, description, visibility, status, created_at, updated_at
	`,
		params.NS,
		params.Key,
		params.DisplayName,
		params.Description,
		params.Visibility,
	).Scan(
		&folder.ID,
		&folder.NS,
		&folder.Key,
		&folder.DisplayName,
		&folder.Description,
		&folder.Visibility,
		&folder.Status,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)
	return folder, err
}

func (r *Repository) GetFolder(ctx context.Context, ns string, folderID int64) (Folder, error) {
	var folder Folder
	err := r.pool.QueryRow(ctx, `
		SELECT id, ns, key, display_name, description, visibility, status, created_at, updated_at
		FROM folders
		WHERE ns = $1 AND id = $2
	`, ns, folderID).Scan(
		&folder.ID,
		&folder.NS,
		&folder.Key,
		&folder.DisplayName,
		&folder.Description,
		&folder.Visibility,
		&folder.Status,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Folder{}, ErrNotFound
	}
	return folder, err
}

func (r *Repository) ListFolders(ctx context.Context, ns string) ([]Folder, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ns, key, display_name, description, visibility, status, created_at, updated_at
		FROM folders
		WHERE ns = $1
		ORDER BY key ASC
	`, ns)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Folder, 0)
	for rows.Next() {
		var item Folder
		if err := rows.Scan(
			&item.ID,
			&item.NS,
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

func (r *Repository) ArchiveFolder(ctx context.Context, ns string, folderID int64) (Folder, error) {
	var activeDocs int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(1)
		FROM documents
		WHERE ns = $1 AND folder_id = $2 AND status = 'active'
	`, ns, folderID).Scan(&activeDocs); err != nil {
		return Folder{}, err
	}
	if activeDocs > 0 {
		return Folder{}, ErrConflict
	}

	var folder Folder
	err := r.pool.QueryRow(ctx, `
		UPDATE folders
		SET status = 'archived', updated_at = NOW()
		WHERE ns = $1 AND id = $2
		RETURNING id, ns, key, display_name, description, visibility, status, created_at, updated_at
	`, ns, folderID).Scan(
		&folder.ID,
		&folder.NS,
		&folder.Key,
		&folder.DisplayName,
		&folder.Description,
		&folder.Visibility,
		&folder.Status,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Folder{}, ErrNotFound
	}
	return folder, err
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
			SELECT 1 FROM folders WHERE ns = $1 AND id = $2
		)
	`, params.NS, params.FolderID).Scan(&namespaceExists); err != nil {
		return Document{}, Revision{}, err
	}
	if !namespaceExists {
		return Document{}, Revision{}, ErrNotFound
	}

	var documentID int64
	var createdAt time.Time
	if err = tx.QueryRow(ctx, `
		INSERT INTO documents (
			ns, folder_id, slug, title, content_md, source_id, source_label, source_category,
			source_input_mode, source_original_ref, source_captured_at, source_content_type, source_adapter, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'active')
		RETURNING id, created_at
	`,
		params.NS,
		params.FolderID,
		params.Slug,
		params.Title,
		params.Content,
		params.Source.ID,
		params.Source.Label,
		params.Source.Category,
		params.Source.InputMode,
		params.Source.OriginalRef,
		params.Source.CapturedAt,
		params.Source.ContentType,
		params.Source.Adapter,
	).Scan(&documentID, &createdAt); err != nil {
		return Document{}, Revision{}, err
	}

	revision, err := insertRevision(ctx, tx, documentID, params.Title, params.Content, params.Source, params.AuthorType, params.AuthorID, params.ChangeSummary, 1)
	if err != nil {
		return Document{}, Revision{}, err
	}

	var document Document
	if err = tx.QueryRow(ctx, `
		UPDATE documents
		SET current_revision_id = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, ns, folder_id, slug, title, content_md,
		          source_id, source_label, source_category, source_input_mode, source_original_ref,
		          source_captured_at, source_content_type, source_adapter,
		          status, current_revision_id, created_at, updated_at
	`, revision.ID, documentID).Scan(
		&document.ID,
		&document.NS,
		&document.FolderID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Source.ID,
		&document.Source.Label,
		&document.Source.Category,
		&document.Source.InputMode,
		&document.Source.OriginalRef,
		&document.Source.CapturedAt,
		&document.Source.ContentType,
		&document.Source.Adapter,
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
		SELECT id, ns, folder_id, slug, title, content_md,
		       source_id, source_label, source_category, source_input_mode, source_original_ref,
		       source_captured_at, source_content_type, source_adapter,
		       status, current_revision_id, created_at, updated_at
		FROM documents
		WHERE ns = $1 AND id = $2
	`, params.NS, params.DocumentID).Scan(
		&existing.ID,
		&existing.NS,
		&existing.FolderID,
		&existing.Slug,
		&existing.Title,
		&existing.Content,
		&existing.Source.ID,
		&existing.Source.Label,
		&existing.Source.Category,
		&existing.Source.InputMode,
		&existing.Source.OriginalRef,
		&existing.Source.CapturedAt,
		&existing.Source.ContentType,
		&existing.Source.Adapter,
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
		existing.Source,
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
		WHERE ns = $4 AND id = $5
		RETURNING id, ns, folder_id, slug, title, content_md,
		          source_id, source_label, source_category, source_input_mode, source_original_ref,
		          source_captured_at, source_content_type, source_adapter,
		          status, current_revision_id, created_at, updated_at
	`, params.Title, params.Content, revision.ID, params.NS, params.DocumentID).Scan(
		&document.ID,
		&document.NS,
		&document.FolderID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Source.ID,
		&document.Source.Label,
		&document.Source.Category,
		&document.Source.InputMode,
		&document.Source.OriginalRef,
		&document.Source.CapturedAt,
		&document.Source.ContentType,
		&document.Source.Adapter,
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
			d.ns,
			d.folder_id,
			d.slug,
			d.title,
			d.content_md,
			d.source_id,
			d.source_label,
			d.source_category,
			d.source_input_mode,
			d.source_original_ref,
			d.source_captured_at,
			d.source_content_type,
			d.source_adapter,
			d.status,
			d.current_revision_id,
			COALESCE(r.revision_no, 0),
			d.created_at,
			d.updated_at
		FROM documents d
		LEFT JOIN revisions r ON r.id = d.current_revision_id
		WHERE d.ns = $1 AND d.id = $2
	`, tenantID, documentID).Scan(
		&document.ID,
		&document.NS,
		&document.FolderID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Source.ID,
		&document.Source.Label,
		&document.Source.Category,
		&document.Source.InputMode,
		&document.Source.OriginalRef,
		&document.Source.CapturedAt,
		&document.Source.ContentType,
		&document.Source.Adapter,
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
		SELECT id, document_id, revision_no, title, content_md,
		       source_id, source_label, source_category, source_input_mode, source_original_ref,
		       source_captured_at, source_content_type, source_adapter,
		       author_type, author_id, change_summary, created_at
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
			&revision.Source.ID,
			&revision.Source.Label,
			&revision.Source.Category,
			&revision.Source.InputMode,
			&revision.Source.OriginalRef,
			&revision.Source.CapturedAt,
			&revision.Source.ContentType,
			&revision.Source.Adapter,
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
		WHERE ns = $1 AND folder_id = $2 AND slug = $3
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
		SELECT d.id, d.ns, d.folder_id, d.slug, d.title, d.content_md,
		       d.source_id, d.source_label, d.source_category, d.source_input_mode, d.source_original_ref,
		       d.source_captured_at, d.source_content_type, d.source_adapter,
		       d.status, d.current_revision_id,
		       COALESCE(r.revision_no, 0), d.created_at, d.updated_at
		FROM documents d
		LEFT JOIN revisions r ON r.id = d.current_revision_id
		WHERE d.ns = $1
	`
	args := []any{tenantID}
	nextArg := 2
	if namespaceID != nil {
		query += fmt.Sprintf(` AND d.folder_id = $%d`, nextArg)
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
			&item.NS,
			&item.FolderID,
			&item.Slug,
			&item.Title,
			&item.Content,
			&item.Source.ID,
			&item.Source.Label,
			&item.Source.Category,
			&item.Source.InputMode,
			&item.Source.OriginalRef,
			&item.Source.CapturedAt,
			&item.Source.ContentType,
			&item.Source.Adapter,
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
		SELECT id, ns, folder_id, slug, title, content_md,
		       source_id, source_label, source_category, source_input_mode, source_original_ref,
		       source_captured_at, source_content_type, source_adapter,
		       status, current_revision_id, created_at, updated_at
		FROM documents
		WHERE ns = $1 AND id = $2
	`, params.NS, params.DocumentID).Scan(
		&current.ID,
		&current.NS,
		&current.FolderID,
		&current.Slug,
		&current.Title,
		&current.Content,
		&current.Source.ID,
		&current.Source.Label,
		&current.Source.Category,
		&current.Source.InputMode,
		&current.Source.OriginalRef,
		&current.Source.CapturedAt,
		&current.Source.ContentType,
		&current.Source.Adapter,
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
		current.Source,
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
		WHERE ns = $2 AND id = $3
		RETURNING id, ns, folder_id, slug, title, content_md,
		          source_id, source_label, source_category, source_input_mode, source_original_ref,
		          source_captured_at, source_content_type, source_adapter,
		          status, current_revision_id, created_at, updated_at
	`, revision.ID, params.NS, params.DocumentID).Scan(
		&document.ID,
		&document.NS,
		&document.FolderID,
		&document.Slug,
		&document.Title,
		&document.Content,
		&document.Source.ID,
		&document.Source.Label,
		&document.Source.Category,
		&document.Source.InputMode,
		&document.Source.OriginalRef,
		&document.Source.CapturedAt,
		&document.Source.ContentType,
		&document.Source.Adapter,
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
	source DocumentSource,
	authorType string,
	authorID string,
	changeSummary string,
	revisionNo int32,
) (Revision, error) {
	var revision Revision
	err := tx.QueryRow(ctx, `
		INSERT INTO revisions (
			document_id, revision_no, title, content_md,
			source_id, source_label, source_category, source_input_mode, source_original_ref,
			source_captured_at, source_content_type, source_adapter,
			author_type, author_id, change_summary
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, document_id, revision_no, title, content_md,
		          source_id, source_label, source_category, source_input_mode, source_original_ref,
		          source_captured_at, source_content_type, source_adapter,
		          author_type, author_id, change_summary, created_at
	`, documentID, revisionNo, title, content, source.ID, source.Label, source.Category, source.InputMode, source.OriginalRef, source.CapturedAt, source.ContentType, source.Adapter, authorType, authorID, changeSummary).Scan(
		&revision.ID,
		&revision.DocumentID,
		&revision.RevisionNo,
		&revision.Title,
		&revision.Content,
		&revision.Source.ID,
		&revision.Source.Label,
		&revision.Source.Category,
		&revision.Source.InputMode,
		&revision.Source.OriginalRef,
		&revision.Source.CapturedAt,
		&revision.Source.ContentType,
		&revision.Source.Adapter,
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
