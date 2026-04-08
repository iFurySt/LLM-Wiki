package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
	"github.com/ifuryst/llm-wiki/internal/auth"
	"github.com/ifuryst/llm-wiki/internal/service"
	"github.com/ifuryst/llm-wiki/internal/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Manager struct {
	svc     *service.Service
	mu      sync.Mutex
	servers map[string]*mcp.Server
}

func NewManager(svc *service.Service) *Manager {
	return &Manager{
		svc:     svc,
		servers: make(map[string]*mcp.Server),
	}
}

func (m *Manager) ServerForPrincipal(principal auth.Principal) *mcp.Server {
	key := principal.NS + "|" + principal.PrincipalID + "|" + strings.Join(principal.Scopes, ",")
	m.mu.Lock()
	defer m.mu.Unlock()

	if server, ok := m.servers[key]; ok {
		return server
	}

	server := m.newServer(principal)
	m.servers[key] = server
	return server
}

func StreamableHTTPHandler(manager *Manager) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		principal, _ := auth.PrincipalFromContext(r.Context())
		return manager.ServerForPrincipal(principal)
	}, &mcp.StreamableHTTPOptions{
		SessionTimeout: 30 * time.Minute,
	})
}

func SSEHandler(manager *Manager) http.Handler {
	return mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		principal, _ := auth.PrincipalFromContext(r.Context())
		return manager.ServerForPrincipal(principal)
	}, &mcp.SSEOptions{})
}

func (m *Manager) newServer(principal auth.Principal) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "LLM-Wiki",
		Version: version.Version,
	}, nil)

	m.registerTools(server, principal)
	m.registerResources(server, principal)
	return server
}

type emptyInput struct{}

type listSpacesOutput struct {
	NS    string           `json:"ns"`
	Items []api.NSResponse `json:"items"`
}

type listFoldersOutput struct {
	NS    string               `json:"ns"`
	Items []api.FolderResponse `json:"items"`
}

type createFolderInput struct {
	Key         string `json:"key" jsonschema:"stable folder key"`
	DisplayName string `json:"display_name" jsonschema:"human-readable folder name"`
	Description string `json:"description,omitempty" jsonschema:"folder description"`
	Visibility  string `json:"visibility,omitempty" jsonschema:"private, ns, team, or restricted"`
}

type listDocumentsInput struct {
	FolderID *int64  `json:"folder_id,omitempty" jsonschema:"optional folder filter"`
	Status   *string `json:"status,omitempty" jsonschema:"optional status filter such as active or archived"`
}

type listDocumentsOutput struct {
	NS    string                 `json:"ns"`
	Items []api.DocumentResponse `json:"items"`
}

type getDocumentInput struct {
	ID int64 `json:"id" jsonschema:"document ID"`
}

type getDocumentBySlugInput struct {
	FolderID int64  `json:"folder_id" jsonschema:"folder ID"`
	Slug     string `json:"slug" jsonschema:"document slug"`
}

type createDocumentInput struct {
	FolderID      int64               `json:"folder_id" jsonschema:"folder ID"`
	Slug          string              `json:"slug" jsonschema:"document slug"`
	Title         string              `json:"title" jsonschema:"document title"`
	Content       string              `json:"content,omitempty" jsonschema:"markdown content"`
	Source        *api.DocumentSource `json:"source,omitempty" jsonschema:"optional document source metadata"`
	AuthorType    string              `json:"author_type,omitempty" jsonschema:"user, agent, or system"`
	AuthorID      string              `json:"author_id,omitempty" jsonschema:"author identifier"`
	ChangeSummary string              `json:"change_summary,omitempty" jsonschema:"short reason for the change"`
}

type updateDocumentInput struct {
	ID            int64  `json:"id" jsonschema:"document ID"`
	Title         string `json:"title" jsonschema:"document title"`
	Content       string `json:"content,omitempty" jsonschema:"markdown content"`
	AuthorType    string `json:"author_type,omitempty" jsonschema:"user, agent, or system"`
	AuthorID      string `json:"author_id,omitempty" jsonschema:"author identifier"`
	ChangeSummary string `json:"change_summary,omitempty" jsonschema:"short reason for the change"`
}

type archiveDocumentInput struct {
	ID            int64  `json:"id" jsonschema:"document ID"`
	AuthorType    string `json:"author_type,omitempty" jsonschema:"user, agent, or system"`
	AuthorID      string `json:"author_id,omitempty" jsonschema:"author identifier"`
	ChangeSummary string `json:"change_summary,omitempty" jsonschema:"archive reason"`
}

type archiveFolderInput struct {
	ID int64 `json:"id" jsonschema:"folder ID"`
}

func (m *Manager) registerTools(server *mcp.Server, principal auth.Principal) {
	ns := principal.NS
	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_ns",
		Title:       "List NS",
		Description: "List ns visible to the current principal.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listSpacesOutput, error) {
		if !auth.HasScopes(principal, auth.ScopeNSRead) {
			return nil, listSpacesOutput{}, fmt.Errorf("missing %s scope", auth.ScopeNSRead)
		}
		resp, err := m.svc.ListNS(ctx, ns)
		if err != nil {
			return nil, listSpacesOutput{}, err
		}
		return nil, listSpacesOutput{NS: ns, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_folders",
		Title:       "List Folders",
		Description: "List folders in the current ns.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listFoldersOutput, error) {
		if !auth.HasScopes(principal, auth.ScopeFoldersRead) {
			return nil, listFoldersOutput{}, fmt.Errorf("missing %s scope", auth.ScopeFoldersRead)
		}
		resp, err := m.svc.ListFolders(ctx, ns)
		if err != nil {
			return nil, listFoldersOutput{}, err
		}
		return nil, listFoldersOutput{NS: ns, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_create_folder",
		Title:       "Create Folder",
		Description: "Create a folder in the current ns.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createFolderInput) (*mcp.CallToolResult, api.FolderResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeFoldersWrite) {
			return nil, api.FolderResponse{}, fmt.Errorf("missing %s scope", auth.ScopeFoldersWrite)
		}
		resp, err := m.svc.CreateFolder(ctx, ns, api.CreateFolderRequest{
			Key:         in.Key,
			DisplayName: in.DisplayName,
			Description: in.Description,
			Visibility:  in.Visibility,
		})
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_archive_folder",
		Title:       "Archive Folder",
		Description: "Archive a folder in the current ns.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in archiveFolderInput) (*mcp.CallToolResult, api.FolderResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeFoldersWrite) {
			return nil, api.FolderResponse{}, fmt.Errorf("missing %s scope", auth.ScopeFoldersWrite)
		}
		resp, err := m.svc.ArchiveFolder(ctx, ns, in.ID)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_documents",
		Title:       "List Documents",
		Description: "List documents in the current ns, optionally filtered by folder or status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listDocumentsInput) (*mcp.CallToolResult, listDocumentsOutput, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsRead) {
			return nil, listDocumentsOutput{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsRead)
		}
		resp, err := m.svc.ListDocuments(ctx, ns, in.FolderID, in.Status)
		if err != nil {
			return nil, listDocumentsOutput{}, err
		}
		return nil, listDocumentsOutput{NS: ns, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_get_document",
		Title:       "Get Document",
		Description: "Fetch a document and its revisions by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsRead) {
			return nil, api.DocumentResponse{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsRead)
		}
		resp, err := m.svc.GetDocument(ctx, ns, in.ID)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_get_document_by_slug",
		Title:       "Get Document By Slug",
		Description: "Fetch a document and its revisions by folder ID and slug.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getDocumentBySlugInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsRead) {
			return nil, api.DocumentResponse{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsRead)
		}
		resp, err := m.svc.GetDocumentBySlug(ctx, ns, in.FolderID, in.Slug)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_create_document",
		Title:       "Create Document",
		Description: "Create a new LLM-Wiki document and its first revision.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsWrite) {
			return nil, api.DocumentResponse{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsWrite)
		}
		resp, err := m.svc.CreateDocument(ctx, ns, api.CreateDocumentRequest{
			FolderID:      in.FolderID,
			Slug:          in.Slug,
			Title:         in.Title,
			Content:       in.Content,
			Source:        in.Source,
			AuthorType:    in.AuthorType,
			AuthorID:      in.AuthorID,
			ChangeSummary: in.ChangeSummary,
		})
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_update_document",
		Title:       "Update Document",
		Description: "Update an existing LLM-Wiki document and create a new revision.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in updateDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsWrite) {
			return nil, api.DocumentResponse{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsWrite)
		}
		resp, err := m.svc.UpdateDocument(ctx, ns, in.ID, api.UpdateDocumentRequest{
			Title:         in.Title,
			Content:       in.Content,
			AuthorType:    in.AuthorType,
			AuthorID:      in.AuthorID,
			ChangeSummary: in.ChangeSummary,
		})
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_archive_document",
		Title:       "Archive Document",
		Description: "Archive a LLM-Wiki document while preserving revision history.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in archiveDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		if !auth.HasScopes(principal, auth.ScopeDocumentsArchive) {
			return nil, api.DocumentResponse{}, fmt.Errorf("missing %s scope", auth.ScopeDocumentsArchive)
		}
		resp, err := m.svc.ArchiveDocument(ctx, ns, in.ID, api.ArchiveDocumentRequest{
			AuthorType:    in.AuthorType,
			AuthorID:      in.AuthorID,
			ChangeSummary: in.ChangeSummary,
		})
		return nil, resp, err
	})
}

func (m *Manager) registerResources(server *mcp.Server, principal auth.Principal) {
	tenantID := principal.NS
	server.AddResource(&mcp.Resource{
		Name:        "llm-wiki-ns",
		Title:       "LLM-Wiki Spaces",
		Description: "All ns visible to the current principal.",
		MIMEType:    "application/json",
		URI:         "llm-wiki://ns",
	}, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := m.svc.ListNS(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		return jsonResource("llm-wiki://ns", resp)
	})

	server.AddResource(&mcp.Resource{
		Name:        "llm-wiki-folders",
		Title:       "LLM-Wiki Folders",
		Description: "All folders visible in the current ns.",
		MIMEType:    "application/json",
		URI:         "llm-wiki://folders",
	}, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := m.svc.ListFolders(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		return jsonResource("llm-wiki://folders", resp)
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "llm-wiki-document",
		Title:       "LLM-Wiki Document",
		Description: "Read a document and its revision history by document ID.",
		MIMEType:    "application/json",
		URITemplate: "llm-wiki://documents/{id}",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		documentID, err := documentIDFromURI(req.Params.URI)
		if err != nil {
			return nil, err
		}
		resp, err := m.svc.GetDocument(ctx, tenantID, documentID)
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, resp)
	})

	server.AddResourceTemplate(&mcp.ResourceTemplate{
		Name:        "llm-wiki-document-by-slug",
		Title:       "LLM-Wiki Document By Slug",
		Description: "Read a document by folder ID and slug.",
		MIMEType:    "application/json",
		URITemplate: "llm-wiki://documents/by-slug/{folder_id}/{slug}",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		namespaceID, slug, err := documentSlugFromURI(req.Params.URI)
		if err != nil {
			return nil, err
		}
		resp, err := m.svc.GetDocumentBySlug(ctx, tenantID, namespaceID, slug)
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, resp)
	})
}

func jsonResource(uri string, value any) (*mcp.ReadResourceResult, error) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(payload),
		}},
	}, nil
}

func documentIDFromURI(raw string) (int64, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return 0, err
	}
	if parsed.Scheme != "llm-wiki" || parsed.Host != "documents" {
		return 0, fmt.Errorf("invalid document resource URI: %s", raw)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 1 || parts[0] == "" {
		return 0, fmt.Errorf("invalid document resource URI: %s", raw)
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func documentSlugFromURI(raw string) (int64, string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return 0, "", err
	}
	if parsed.Scheme != "llm-wiki" || parsed.Host != "documents" {
		return 0, "", fmt.Errorf("invalid document-by-slug resource URI: %s", raw)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "by-slug" || parts[1] == "" || parts[2] == "" {
		return 0, "", fmt.Errorf("invalid document-by-slug resource URI: %s", raw)
	}
	namespaceID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, "", err
	}
	return namespaceID, parts[2], nil
}
