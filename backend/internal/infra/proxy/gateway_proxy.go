package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RequestData is the normalized upstream request payload used by the gateway
// orchestration and script transforms.
type RequestData struct {
	Headers      map[string]string
	Query        map[string]string
	QueryChanged bool
	Body         []byte
	ContentType  string
}

// ProxyResponse captures the upstream response for gateway handling.
type ProxyResponse struct {
	StatusCode       int
	Headers          http.Header
	Body             []byte
	ContentType      string
	TransformBody    []byte
	TransformHeaders map[string]string
}

// NewHTTPClient builds the upstream HTTP client used by the API gateway.
func NewHTTPClient() *http.Client {
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableCompression: true}}
}

// Forward relays a request to the upstream endpoint, merging incoming query
// parameters and transformed headers/query.
func Forward(ctx context.Context, client *http.Client, incoming *http.Request, endpoint string, input RequestData) (ProxyResponse, error) {
	upstream, err := url.Parse(endpoint)
	if err != nil || upstream.Scheme == "" || upstream.Host == "" {
		return ProxyResponse{}, &url.Error{Op: "parse", URL: endpoint, Err: err}
	}
	query := upstream.Query()
	for key, values := range incoming.URL.Query() {
		query.Del(key)
		for _, value := range values {
			query.Add(key, value)
		}
	}
	if input.QueryChanged {
		for key, value := range input.Query {
			query.Set(key, value)
		}
	}
	upstream.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, incoming.Method, upstream.String(), bytes.NewReader(input.Body))
	if err != nil {
		return ProxyResponse{}, err
	}
	CopyRequestHeadersWithTransforms(req.Header, incoming.Header, input.Headers)
	req.ContentLength = int64(len(input.Body))
	resp, err := client.Do(req)
	if err != nil {
		return ProxyResponse{}, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ProxyResponse{}, err
	}
	headers := make(http.Header, len(resp.Header))
	CopyResponseHeaders(headers, resp.Header)
	return ProxyResponse{
		StatusCode:       resp.StatusCode,
		Headers:          headers,
		Body:             body,
		ContentType:      resp.Header.Get("Content-Type"),
		TransformBody:    TransformPayload(body, resp.Header.Get("Content-Type")),
		TransformHeaders: FlattenHeaders(resp.Header),
	}, nil
}

// CopyRequestHeadersWithTransforms copies request headers excluding
// hop-by-hop/authorization/host, then applies transformed header values.
func CopyRequestHeadersWithTransforms(dst, src http.Header, transformed map[string]string) {
	blocked := connectionHeaders(src)
	for key, values := range src {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "host" || isHopByHop(lower) || blocked[lower] {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
	for key, value := range transformed {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "host" || isHopByHop(lower) || blocked[lower] {
			continue
		}
		dst.Del(key)
		dst.Set(key, value)
	}
}

func connectionHeaders(headers http.Header) map[string]bool {
	blocked := make(map[string]bool)
	for _, value := range headers.Values("Connection") {
		for _, token := range strings.Split(value, ",") {
			blocked[strings.ToLower(strings.TrimSpace(token))] = true
		}
	}
	return blocked
}

var hopByHopHeaders = map[string]struct{}{
	"connection": {}, "keep-alive": {}, "proxy-authenticate": {}, "proxy-authorization": {},
	"proxy-connection": {}, "te": {}, "trailer": {}, "transfer-encoding": {}, "upgrade": {},
	"content-length": {},
}

func isHopByHop(key string) bool {
	if _, ok := hopByHopHeaders[key]; ok {
		return true
	}
	return false
}

// IsHopByHopHeader reports whether a header must not be forwarded.
func IsHopByHopHeader(key string) bool { return isHopByHop(key) }
