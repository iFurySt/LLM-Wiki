package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

var keyPattern = regexp.MustCompile(`^[a-z0-9]+(?:[a-z0-9-_]*[a-z0-9])?$`)

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Repo() *repository.Repository {
	return s.repo
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) CreateFolder(ctx context.Context, ns string, req api.CreateFolderRequest) (api.FolderResponse, error) {
	key := normalizeKey(req.Key)
	if err := validateKeyLike("folder key", key); err != nil {
		return api.FolderResponse{}, err
	}
	item, err := s.repo.CreateFolder(ctx, repository.CreateFolderParams{
		NS:          ns,
		Key:         key,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Description: strings.TrimSpace(req.Description),
		Visibility:  normalizeVisibility(req.Visibility),
	})
	if err != nil {
		return api.FolderResponse{}, err
	}
	return folderResponse(item), nil
}

func (s *Service) GetFolder(ctx context.Context, ns string, folderID int64) (api.FolderResponse, error) {
	item, err := s.repo.GetFolder(ctx, ns, folderID)
	if err != nil {
		return api.FolderResponse{}, err
	}
	return folderResponse(item), nil
}

func (s *Service) ListFolders(ctx context.Context, ns string) (api.ListFoldersResponse, error) {
	items, err := s.repo.ListFolders(ctx, ns)
	if err != nil {
		return api.ListFoldersResponse{}, err
	}
	resp := make([]api.FolderResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, folderResponse(item))
	}
	return api.ListFoldersResponse{Items: resp}, nil
}

func (s *Service) ArchiveFolder(ctx context.Context, ns string, folderID int64) (api.FolderResponse, error) {
	item, err := s.repo.ArchiveFolder(ctx, ns, folderID)
	if err != nil {
		return api.FolderResponse{}, err
	}
	return folderResponse(item), nil
}

func (s *Service) CreateDocument(ctx context.Context, ns string, req api.CreateDocumentRequest) (api.DocumentResponse, error) {
	slug := normalizeKey(req.Slug)
	if err := validateKeyLike("document slug", slug); err != nil {
		return api.DocumentResponse{}, err
	}
	authorType, authorID := authorFromContext(ctx, req.AuthorType, req.AuthorID)
	doc, rev, err := s.repo.CreateDocument(ctx, repository.CreateDocumentParams{
		NS:            ns,
		FolderID:      req.FolderID,
		Slug:          slug,
		Title:         strings.TrimSpace(req.Title),
		Content:       req.Content,
		AuthorType:    authorType,
		AuthorID:      authorID,
		ChangeSummary: strings.TrimSpace(req.ChangeSummary),
	})
	if err != nil {
		return api.DocumentResponse{}, err
	}
	return toDocumentResponse(doc, []repository.Revision{rev}), nil
}

func (s *Service) UpdateDocument(ctx context.Context, ns string, documentID int64, req api.UpdateDocumentRequest) (api.DocumentResponse, error) {
	authorType, authorID := authorFromContext(ctx, req.AuthorType, req.AuthorID)
	doc, rev, err := s.repo.UpdateDocument(ctx, repository.UpdateDocumentParams{
		NS:            ns,
		DocumentID:    documentID,
		Title:         strings.TrimSpace(req.Title),
		Content:       req.Content,
		AuthorType:    authorType,
		AuthorID:      authorID,
		ChangeSummary: strings.TrimSpace(req.ChangeSummary),
	})
	if err != nil {
		return api.DocumentResponse{}, err
	}
	return toDocumentResponse(doc, []repository.Revision{rev}), nil
}

func (s *Service) GetDocument(ctx context.Context, ns string, documentID int64) (api.DocumentResponse, error) {
	doc, revisions, err := s.repo.GetDocument(ctx, ns, documentID)
	if err != nil {
		return api.DocumentResponse{}, err
	}
	return toDocumentResponse(doc, revisions), nil
}

func (s *Service) GetDocumentBySlug(ctx context.Context, ns string, folderID int64, slug string) (api.DocumentResponse, error) {
	normalizedSlug := normalizeKey(slug)
	if err := validateKeyLike("document slug", normalizedSlug); err != nil {
		return api.DocumentResponse{}, err
	}
	doc, revisions, err := s.repo.GetDocumentBySlug(ctx, ns, folderID, normalizedSlug)
	if err != nil {
		return api.DocumentResponse{}, err
	}
	return toDocumentResponse(doc, revisions), nil
}

func (s *Service) ListDocuments(ctx context.Context, ns string, folderID *int64, status *string) (api.ListDocumentsResponse, error) {
	items, err := s.repo.ListDocuments(ctx, ns, folderID, status)
	if err != nil {
		return api.ListDocumentsResponse{}, err
	}
	resp := make([]api.DocumentResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toDocumentResponse(item, nil))
	}
	return api.ListDocumentsResponse{Items: resp}, nil
}

func (s *Service) ArchiveDocument(ctx context.Context, ns string, documentID int64, req api.ArchiveDocumentRequest) (api.DocumentResponse, error) {
	authorType, authorID := authorFromContext(ctx, req.AuthorType, req.AuthorID)
	doc, rev, err := s.repo.ArchiveDocument(ctx, repository.ArchiveDocumentParams{
		NS:            ns,
		DocumentID:    documentID,
		AuthorType:    authorType,
		AuthorID:      authorID,
		ChangeSummary: strings.TrimSpace(req.ChangeSummary),
	})
	if err != nil {
		return api.DocumentResponse{}, err
	}
	return toDocumentResponse(doc, []repository.Revision{rev}), nil
}

func folderResponse(item repository.Folder) api.FolderResponse {
	return api.FolderResponse{
		ID:          item.ID,
		NS:          item.NS,
		Key:         item.Key,
		DisplayName: item.DisplayName,
		Description: item.Description,
		Visibility:  item.Visibility,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}

func toDocumentResponse(doc repository.Document, revisions []repository.Revision) api.DocumentResponse {
	result := api.DocumentResponse{
		ID:                doc.ID,
		NS:                doc.NS,
		FolderID:          doc.FolderID,
		Slug:              doc.Slug,
		Title:             doc.Title,
		Content:           doc.Content,
		Status:            doc.Status,
		CurrentRevisionID: doc.CurrentRevisionID,
		CurrentRevisionNo: doc.CurrentRevisionNo,
		CreatedAt:         doc.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         doc.UpdatedAt.Format(time.RFC3339),
	}
	if len(revisions) == 0 {
		return result
	}
	result.Revisions = make([]api.RevisionResponse, 0, len(revisions))
	for _, item := range revisions {
		result.Revisions = append(result.Revisions, api.RevisionResponse{
			ID:            item.ID,
			DocumentID:    item.DocumentID,
			RevisionNo:    item.RevisionNo,
			Title:         item.Title,
			Content:       item.Content,
			AuthorType:    item.AuthorType,
			AuthorID:      item.AuthorID,
			ChangeSummary: item.ChangeSummary,
			CreatedAt:     item.CreatedAt.Format(time.RFC3339),
		})
	}
	return result
}

func normalizeVisibility(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "ns", "team", "restricted", "private":
		return normalized
	default:
		return "private"
	}
}

func normalizeAuthorType(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "user", "agent", "system":
		return normalized
	default:
		return "agent"
	}
}

func authorFromContext(ctx context.Context, explicitType string, explicitID string) (string, string) {
	authorType := normalizeAuthorType(explicitType)
	authorID := strings.TrimSpace(explicitID)
	if authorID != "" {
		return authorType, authorID
	}
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		return authorType, authorID
	}
	switch principal.PrincipalType {
	case "user":
		authorType = "user"
	case "service", "admin":
		authorType = "system"
	default:
		authorType = "agent"
	}
	return authorType, principal.PrincipalID
}

func normalizeKey(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	normalized = strings.ReplaceAll(normalized, " ", "-")
	return normalized
}

func validateKeyLike(label string, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", label)
	}
	if !keyPattern.MatchString(value) {
		return fmt.Errorf("%s must match %s", label, keyPattern.String())
	}
	return nil
}
