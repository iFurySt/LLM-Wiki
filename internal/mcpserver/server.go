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

func (m *Manager) ServerForTenant(tenantID string) *mcp.Server {
	if strings.TrimSpace(tenantID) == "" {
		tenantID = "default"
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if server, ok := m.servers[tenantID]; ok {
		return server
	}

	server := m.newServer(tenantID)
	m.servers[tenantID] = server
	return server
}

func StreamableHTTPHandler(manager *Manager) http.Handler {
	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return manager.ServerForTenant(r.Header.Get("X-LLM-Wiki-Tenant-ID"))
	}, &mcp.StreamableHTTPOptions{
		SessionTimeout: 30 * time.Minute,
	})
}

func SSEHandler(manager *Manager) http.Handler {
	return mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return manager.ServerForTenant(r.Header.Get("X-LLM-Wiki-Tenant-ID"))
	}, &mcp.SSEOptions{})
}

func (m *Manager) newServer(tenantID string) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "LLM-Wiki",
		Version: version.Version,
	}, nil)

	m.registerTools(server, tenantID)
	m.registerResources(server, tenantID)
	return server
}

type emptyInput struct{}

type listSpacesOutput struct {
	TenantID string              `json:"tenant_id"`
	Items    []api.SpaceResponse `json:"items"`
}

type listNamespacesOutput struct {
	TenantID string                  `json:"tenant_id"`
	Items    []api.NamespaceResponse `json:"items"`
}

type createNamespaceInput struct {
	Key         string `json:"key" jsonschema:"stable namespace key"`
	DisplayName string `json:"display_name" jsonschema:"human-readable namespace name"`
	Description string `json:"description,omitempty" jsonschema:"namespace description"`
	Visibility  string `json:"visibility,omitempty" jsonschema:"private, tenant, team, or restricted"`
}

type listDocumentsInput struct {
	NamespaceID *int64  `json:"namespace_id,omitempty" jsonschema:"optional namespace filter"`
	Status      *string `json:"status,omitempty" jsonschema:"optional status filter such as active or archived"`
}

type listDocumentsOutput struct {
	TenantID string                 `json:"tenant_id"`
	Items    []api.DocumentResponse `json:"items"`
}

type getDocumentInput struct {
	ID int64 `json:"id" jsonschema:"document ID"`
}

type getDocumentBySlugInput struct {
	NamespaceID int64  `json:"namespace_id" jsonschema:"namespace ID"`
	Slug        string `json:"slug" jsonschema:"document slug"`
}

type createDocumentInput struct {
	NamespaceID   int64  `json:"namespace_id" jsonschema:"namespace ID"`
	Slug          string `json:"slug" jsonschema:"document slug"`
	Title         string `json:"title" jsonschema:"document title"`
	Content       string `json:"content,omitempty" jsonschema:"markdown content"`
	AuthorType    string `json:"author_type,omitempty" jsonschema:"user, agent, or system"`
	AuthorID      string `json:"author_id,omitempty" jsonschema:"author identifier"`
	ChangeSummary string `json:"change_summary,omitempty" jsonschema:"short reason for the change"`
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

type archiveNamespaceInput struct {
	ID int64 `json:"id" jsonschema:"namespace ID"`
}

func (m *Manager) registerTools(server *mcp.Server, tenantID string) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_spaces",
		Title:       "List Spaces",
		Description: "List spaces for the current LLM-Wiki tenant.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listSpacesOutput, error) {
		resp, err := m.svc.ListSpaces(ctx, tenantID)
		if err != nil {
			return nil, listSpacesOutput{}, err
		}
		return nil, listSpacesOutput{TenantID: tenantID, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_namespaces",
		Title:       "List Namespaces",
		Description: "List namespaces for the current LLM-Wiki tenant.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, listNamespacesOutput, error) {
		resp, err := m.svc.ListNamespaces(ctx, tenantID)
		if err != nil {
			return nil, listNamespacesOutput{}, err
		}
		return nil, listNamespacesOutput{TenantID: tenantID, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_create_namespace",
		Title:       "Create Namespace",
		Description: "Create a namespace inside the current tenant's default space.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createNamespaceInput) (*mcp.CallToolResult, api.NamespaceResponse, error) {
		resp, err := m.svc.CreateNamespace(ctx, tenantID, api.CreateNamespaceRequest{
			Key:         in.Key,
			DisplayName: in.DisplayName,
			Description: in.Description,
			Visibility:  in.Visibility,
		})
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_archive_namespace",
		Title:       "Archive Namespace",
		Description: "Archive a namespace in the current tenant.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in archiveNamespaceInput) (*mcp.CallToolResult, api.NamespaceResponse, error) {
		resp, err := m.svc.ArchiveNamespace(ctx, tenantID, in.ID)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_list_documents",
		Title:       "List Documents",
		Description: "List documents in the current tenant, optionally filtered by namespace or status.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listDocumentsInput) (*mcp.CallToolResult, listDocumentsOutput, error) {
		resp, err := m.svc.ListDocuments(ctx, tenantID, in.NamespaceID, in.Status)
		if err != nil {
			return nil, listDocumentsOutput{}, err
		}
		return nil, listDocumentsOutput{TenantID: tenantID, Items: resp.Items}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_get_document",
		Title:       "Get Document",
		Description: "Fetch a document and its revisions by ID.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		resp, err := m.svc.GetDocument(ctx, tenantID, in.ID)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_get_document_by_slug",
		Title:       "Get Document By Slug",
		Description: "Fetch a document and its revisions by namespace ID and slug.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getDocumentBySlugInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		resp, err := m.svc.GetDocumentBySlug(ctx, tenantID, in.NamespaceID, in.Slug)
		return nil, resp, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "llm_wiki_create_document",
		Title:       "Create Document",
		Description: "Create a new LLM-Wiki document and its first revision.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in createDocumentInput) (*mcp.CallToolResult, api.DocumentResponse, error) {
		resp, err := m.svc.CreateDocument(ctx, tenantID, api.CreateDocumentRequest{
			NamespaceID:   in.NamespaceID,
			Slug:          in.Slug,
			Title:         in.Title,
			Content:       in.Content,
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
		resp, err := m.svc.UpdateDocument(ctx, tenantID, in.ID, api.UpdateDocumentRequest{
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
		resp, err := m.svc.ArchiveDocument(ctx, tenantID, in.ID, api.ArchiveDocumentRequest{
			AuthorType:    in.AuthorType,
			AuthorID:      in.AuthorID,
			ChangeSummary: in.ChangeSummary,
		})
		return nil, resp, err
	})
}

func (m *Manager) registerResources(server *mcp.Server, tenantID string) {
	server.AddResource(&mcp.Resource{
		Name:        "llm-wiki-spaces",
		Title:       "LLM-Wiki Spaces",
		Description: "All spaces visible in the current LLM-Wiki tenant.",
		MIMEType:    "application/json",
		URI:         "llm-wiki://spaces",
	}, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := m.svc.ListSpaces(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		return jsonResource("llm-wiki://spaces", resp)
	})

	server.AddResource(&mcp.Resource{
		Name:        "llm-wiki-namespaces",
		Title:       "LLM-Wiki Namespaces",
		Description: "All namespaces visible in the current LLM-Wiki tenant.",
		MIMEType:    "application/json",
		URI:         "llm-wiki://namespaces",
	}, func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		resp, err := m.svc.ListNamespaces(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		return jsonResource("llm-wiki://namespaces", resp)
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
		Description: "Read a document by namespace ID and slug.",
		MIMEType:    "application/json",
		URITemplate: "llm-wiki://documents/by-slug/{namespace_id}/{slug}",
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
