package api

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type WhoAmIResponse struct {
	PrincipalID   string   `json:"principal_id"`
	PrincipalType string   `json:"principal_type"`
	DisplayName   string   `json:"display_name"`
	TenantID      string   `json:"tenant_id"`
	Scopes        []string `json:"scopes"`
	TokenID       int64    `json:"token_id"`
	TokenType     string   `json:"token_type"`
}

type StartBrowserLoginRequest struct {
	TenantID            string   `json:"tenant_id" binding:"required"`
	DisplayName         string   `json:"display_name"`
	Scopes              []string `json:"scopes"`
	State               string   `json:"state" binding:"required"`
	RedirectURI         string   `json:"redirect_uri" binding:"required"`
	CodeChallenge       string   `json:"code_challenge" binding:"required"`
	CodeChallengeMethod string   `json:"code_challenge_method" binding:"required"`
}

type StartBrowserLoginResponse struct {
	RequestID    string `json:"request_id"`
	AuthorizeURL string `json:"authorize_url"`
	ExpiresAt    string `json:"expires_at"`
}

type StartDeviceLoginRequest struct {
	TenantID    string   `json:"tenant_id" binding:"required"`
	DisplayName string   `json:"display_name"`
	Scopes      []string `json:"scopes"`
}

type StartDeviceLoginResponse struct {
	RequestID       string `json:"request_id"`
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	IntervalSeconds int    `json:"interval_seconds"`
	ExpiresAt       string `json:"expires_at"`
}

type TokenExchangeRequest struct {
	GrantType    string `json:"grant_type" binding:"required"`
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	DeviceCode   string `json:"device_code"`
	RefreshToken string `json:"refresh_token"`
}

type TokenExchangeResponse struct {
	AccessToken  string   `json:"access_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	Scopes       []string `json:"scopes"`
	TenantID     string   `json:"tenant_id"`
	PrincipalID  string   `json:"principal_id"`
}

type SetupStatusResponse struct {
	Initialized   bool   `json:"initialized"`
	DefaultTenant string `json:"default_tenant"`
}

type InitializeRequest struct {
	TenantID    string `json:"tenant_id" binding:"required"`
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID          int64  `json:"id"`
	PrincipalID string `json:"principal_id"`
	TenantID    string `json:"tenant_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
	CreatedAt   string `json:"created_at"`
}

type ListUsersResponse struct {
	Items []UserResponse `json:"items"`
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required"`
	IsAdmin     bool   `json:"is_admin"`
}

type ServicePrincipalResponse struct {
	ID            string `json:"id"`
	TenantID      string `json:"tenant_id"`
	PrincipalType string `json:"principal_type"`
	DisplayName   string `json:"display_name"`
	CreatedAt     string `json:"created_at"`
}

type ListServicePrincipalsResponse struct {
	Items []ServicePrincipalResponse `json:"items"`
}

type CreateServicePrincipalRequest struct {
	DisplayName string `json:"display_name" binding:"required"`
}

type IssueTokenRequest struct {
	PrincipalID       string   `json:"principal_id" binding:"required"`
	DisplayName       string   `json:"display_name" binding:"required"`
	Scopes            []string `json:"scopes"`
	ExpiresInSeconds  int      `json:"expires_in_seconds"`
}

type TokenResponse struct {
	ID                   int64    `json:"id"`
	TokenType            string   `json:"token_type"`
	PrincipalID          string   `json:"principal_id"`
	TenantID             string   `json:"tenant_id"`
	DisplayName          string   `json:"display_name"`
	TokenPrefix          string   `json:"token_prefix"`
	Scopes               []string `json:"scopes"`
	ExpiresAt            string   `json:"expires_at,omitempty"`
	LastUsedAt           string   `json:"last_used_at,omitempty"`
	RevokedAt            string   `json:"revoked_at,omitempty"`
	CreatedByPrincipalID string   `json:"created_by_principal_id,omitempty"`
	CreatedAt            string   `json:"created_at"`
	PlaintextToken       string   `json:"plaintext_token,omitempty"`
}

type ListTokensResponse struct {
	Items []TokenResponse `json:"items"`
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
