package api

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SpaceResponse struct {
	ID          int64  `json:"id"`
	TenantID    string `json:"tenant_id"`
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
}

type ListSpacesResponse struct {
	Items []SpaceResponse `json:"items"`
}

type CreateNamespaceRequest struct {
	Key         string `json:"key" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type ArchiveNamespaceRequest struct {
	Reason string `json:"reason"`
}

type NamespaceResponse struct {
	ID          int64  `json:"id"`
	TenantID    string `json:"tenant_id"`
	SpaceID     int64  `json:"space_id"`
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type ListNamespacesResponse struct {
	Items []NamespaceResponse `json:"items"`
}

type CreateDocumentRequest struct {
	NamespaceID   int64  `json:"namespace_id" binding:"required"`
	Slug          string `json:"slug" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Content       string `json:"content"`
	AuthorType    string `json:"author_type"`
	AuthorID      string `json:"author_id"`
	ChangeSummary string `json:"change_summary"`
}

type UpdateDocumentRequest struct {
	Title         string `json:"title" binding:"required"`
	Content       string `json:"content"`
	AuthorType    string `json:"author_type"`
	AuthorID      string `json:"author_id"`
	ChangeSummary string `json:"change_summary"`
}

type DocumentResponse struct {
	ID                int64              `json:"id"`
	TenantID          string             `json:"tenant_id"`
	NamespaceID       int64              `json:"namespace_id"`
	Slug              string             `json:"slug"`
	Title             string             `json:"title"`
	Content           string             `json:"content"`
	Status            string             `json:"status"`
	CurrentRevisionID int64              `json:"current_revision_id"`
	CurrentRevisionNo int32              `json:"current_revision_no"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
	Revisions         []RevisionResponse `json:"revisions,omitempty"`
}

type ListDocumentsResponse struct {
	Items []DocumentResponse `json:"items"`
}

type ArchiveDocumentRequest struct {
	AuthorType    string `json:"author_type"`
	AuthorID      string `json:"author_id"`
	ChangeSummary string `json:"change_summary"`
}

type RevisionResponse struct {
	ID            int64  `json:"id"`
	DocumentID    int64  `json:"document_id"`
	RevisionNo    int32  `json:"revision_no"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	AuthorType    string `json:"author_type"`
	AuthorID      string `json:"author_id"`
	ChangeSummary string `json:"change_summary"`
	CreatedAt     string `json:"created_at"`
}
