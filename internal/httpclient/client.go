package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ifuryst/llm-wiki/internal/api"
)

type Client struct {
	baseURL     string
	httpClient  *http.Client
	accessToken string
}

type SystemInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Server      struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
}

func New(baseURL string, timeout time.Duration, accessToken string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
		accessToken: strings.TrimSpace(accessToken),
	}
}

func (c *Client) SetAccessToken(value string) {
	c.accessToken = strings.TrimSpace(value)
}

func (c *Client) GetSystemInfo(ctx context.Context) (SystemInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/system/info", nil)
	if err != nil {
		return SystemInfo{}, err
	}
	c.applyHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return SystemInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SystemInfo{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var payload SystemInfo
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return SystemInfo{}, err
	}

	return payload, nil
}

func (c *Client) WhoAmI(ctx context.Context) (api.WhoAmIResponse, error) {
	var resp api.WhoAmIResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/auth/whoami", nil, &resp)
	return resp, err
}

func (c *Client) StartBrowserLogin(ctx context.Context, req api.StartBrowserLoginRequest) (api.StartBrowserLoginResponse, error) {
	var resp api.StartBrowserLoginResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/browser/start", req, &resp)
	return resp, err
}

func (c *Client) ListOAuthProviders(ctx context.Context) (api.ListOAuthProvidersResponse, error) {
	var resp api.ListOAuthProvidersResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/auth/providers", nil, &resp)
	return resp, err
}

func (c *Client) StartDeviceLogin(ctx context.Context, req api.StartDeviceLoginRequest) (api.StartDeviceLoginResponse, error) {
	var resp api.StartDeviceLoginResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/device/start", req, &resp)
	return resp, err
}

func (c *Client) ExchangeToken(ctx context.Context, req api.TokenExchangeRequest) (api.TokenExchangeResponse, error) {
	var resp api.TokenExchangeResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/token", req, &resp)
	return resp, err
}

func (c *Client) SwitchNS(ctx context.Context, ns string) (api.TokenExchangeResponse, error) {
	var resp api.TokenExchangeResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/switch-ns", api.SwitchNSRequest{NS: ns}, &resp)
	return resp, err
}

func (c *Client) CreateServicePrincipal(ctx context.Context, req api.CreateServicePrincipalRequest) (api.ServicePrincipalResponse, error) {
	var resp api.ServicePrincipalResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/service-principals", req, &resp)
	return resp, err
}

func (c *Client) ListServicePrincipals(ctx context.Context) (api.ListServicePrincipalsResponse, error) {
	var resp api.ListServicePrincipalsResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/auth/service-principals", nil, &resp)
	return resp, err
}

func (c *Client) IssueToken(ctx context.Context, req api.IssueTokenRequest) (api.TokenResponse, error) {
	var resp api.TokenResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/auth/tokens", req, &resp)
	return resp, err
}

func (c *Client) ListTokens(ctx context.Context) (api.ListTokensResponse, error) {
	var resp api.ListTokensResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/auth/tokens", nil, &resp)
	return resp, err
}

func (c *Client) RevokeToken(ctx context.Context, tokenID int64) (api.TokenResponse, error) {
	var resp api.TokenResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/auth/tokens/%d/revoke", tokenID), map[string]any{}, &resp)
	return resp, err
}

func (c *Client) ListNS(ctx context.Context) (api.ListNSResponse, error) {
	var resp api.ListNSResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/ns", nil, &resp)
	return resp, err
}

func (c *Client) CreateNS(ctx context.Context, req api.CreateNSRequest) (api.NSResponse, error) {
	var resp api.NSResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/ns", req, &resp)
	return resp, err
}

func (c *Client) ListInvites(ctx context.Context) (api.ListInvitesResponse, error) {
	var resp api.ListInvitesResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/ns/invites", nil, &resp)
	return resp, err
}

func (c *Client) CreateInvite(ctx context.Context, req api.CreateInviteRequest) (api.InviteResponse, error) {
	var resp api.InviteResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/ns/invites", req, &resp)
	return resp, err
}

func (c *Client) AcceptInvite(ctx context.Context, token string) (api.NSResponse, error) {
	var resp api.NSResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/ns/invites/accept", api.AcceptInviteRequest{InviteToken: token}, &resp)
	return resp, err
}

func (c *Client) CreateFolder(ctx context.Context, req api.CreateFolderRequest) (api.FolderResponse, error) {
	var resp api.FolderResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/folders", req, &resp)
	return resp, err
}

func (c *Client) GetFolder(ctx context.Context, folderID int64) (api.FolderResponse, error) {
	var resp api.FolderResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/folders/%d", folderID), nil, &resp)
	return resp, err
}

func (c *Client) ListFolders(ctx context.Context) (api.ListFoldersResponse, error) {
	var resp api.ListFoldersResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/folders", nil, &resp)
	return resp, err
}

func (c *Client) ArchiveFolder(ctx context.Context, folderID int64) (api.FolderResponse, error) {
	var resp api.FolderResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/folders/%d/archive", folderID), map[string]any{}, &resp)
	return resp, err
}

func (c *Client) CreateDocument(ctx context.Context, req api.CreateDocumentRequest) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/documents", req, &resp)
	return resp, err
}

func (c *Client) GetDocument(ctx context.Context, documentID int64) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/documents/%d", documentID), nil, &resp)
	return resp, err
}

func (c *Client) GetDocumentBySlug(ctx context.Context, folderID int64, slug string) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/document-by-slug?folder_id=%d&slug=%s", folderID, slug), nil, &resp)
	return resp, err
}

func (c *Client) ListDocuments(ctx context.Context, folderID *int64, status *string) (api.ListDocumentsResponse, error) {
	path := "/v1/documents"
	query := make([]string, 0, 2)
	if folderID != nil {
		query = append(query, fmt.Sprintf("folder_id=%d", *folderID))
	}
	if status != nil && *status != "" {
		query = append(query, fmt.Sprintf("status=%s", *status))
	}
	if len(query) > 0 {
		path += "?" + strings.Join(query, "&")
	}
	var resp api.ListDocumentsResponse
	err := c.doJSON(ctx, http.MethodGet, path, nil, &resp)
	return resp, err
}

func (c *Client) UpdateDocument(ctx context.Context, documentID int64, req api.UpdateDocumentRequest) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/documents/%d", documentID), req, &resp)
	return resp, err
}

func (c *Client) ArchiveDocument(ctx context.Context, documentID int64, req api.ArchiveDocumentRequest) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/documents/%d/archive", documentID), req, &resp)
	return resp, err
}

func (c *Client) doJSON(ctx context.Context, method string, path string, reqBody any, respBody any) error {
	var bodyReader *bytes.Reader
	if reqBody == nil {
		bodyReader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var payload api.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && payload.Error.Message != "" {
			return fmt.Errorf("%s: %s", payload.Error.Code, payload.Error.Message)
		}
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if respBody == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(respBody)
}

func (c *Client) applyHeaders(req *http.Request) {
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Accept", "application/json")
}
