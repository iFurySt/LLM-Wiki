package e2e

import (
	"context"
	"net/http"
	"time"

	"github.com/ifuryst/llm-wiki/internal/service"
)

func bootstrapTestToken(ctx context.Context, svc *service.Service, tenantID string, token string) error {
	return svc.BootstrapToken(ctx, tenantID, "bootstrap-admin", token)
}

func bearerHTTPClient(token string) *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &bearerRoundTripper{
			base:  http.DefaultTransport,
			token: token,
		},
	}
}

type bearerRoundTripper struct {
	base  http.RoundTripper
	token string
}

func (r *bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header = cloned.Header.Clone()
	cloned.Header.Set("Authorization", "Bearer "+r.token)
	return r.base.RoundTrip(cloned)
}
