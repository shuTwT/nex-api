package oauth

import (
	"context"
	"io"
	"net/http"
	"time"
)

// HTTPClient is the minimal outbound client contract used by OAuth providers.
// Keeping the concrete net/http types inside infra lets service orchestration
// remain transport-neutral.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func NewHTTPClient(timeout time.Duration) HTTPClient {
	return &http.Client{Timeout: timeout}
}

func NewRequestWithContext(ctx context.Context, method, rawURL string, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, rawURL, body)
}
