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

func (c *Client) ListSpaces(ctx context.Context) (api.ListSpacesResponse, error) {
	var resp api.ListSpacesResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/spaces", nil, &resp)
	return resp, err
}

func (c *Client) CreateNamespace(ctx context.Context, req api.CreateNamespaceRequest) (api.NamespaceResponse, error) {
	var resp api.NamespaceResponse
	err := c.doJSON(ctx, http.MethodPost, "/v1/namespaces", req, &resp)
	return resp, err
}

func (c *Client) GetNamespace(ctx context.Context, namespaceID int64) (api.NamespaceResponse, error) {
	var resp api.NamespaceResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/namespaces/%d", namespaceID), nil, &resp)
	return resp, err
}

func (c *Client) ListNamespaces(ctx context.Context) (api.ListNamespacesResponse, error) {
	var resp api.ListNamespacesResponse
	err := c.doJSON(ctx, http.MethodGet, "/v1/namespaces", nil, &resp)
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

func (c *Client) GetDocumentBySlug(ctx context.Context, namespaceID int64, slug string) (api.DocumentResponse, error) {
	var resp api.DocumentResponse
	err := c.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/document-by-slug?namespace_id=%d&slug=%s", namespaceID, slug), nil, &resp)
	return resp, err
}

func (c *Client) ListDocuments(ctx context.Context, namespaceID *int64, status *string) (api.ListDocumentsResponse, error) {
	path := "/v1/documents"
	query := make([]string, 0, 2)
	if namespaceID != nil {
		query = append(query, fmt.Sprintf("namespace_id=%d", *namespaceID))
	}
	if status != nil && *status != "" {
		query = append(query, fmt.Sprintf("status=%s", *status))
	}
	if len(query) > 0 {
		path = path + "?" + strings.Join(query, "&")
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

func (c *Client) ArchiveNamespace(ctx context.Context, namespaceID int64) (api.NamespaceResponse, error) {
	var resp api.NamespaceResponse
	err := c.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/namespaces/%d/archive", namespaceID), map[string]any{}, &resp)
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
		if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil {
			if payload.Error.Message != "" {
				return fmt.Errorf("%s: %s", payload.Error.Code, payload.Error.Message)
			}
		}
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	if respBody == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(respBody)
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
}
