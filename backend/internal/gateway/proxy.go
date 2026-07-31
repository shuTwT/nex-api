package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type requestData struct {
	Headers      map[string]string
	Query        map[string]string
	QueryChanged bool
	Body         []byte
	ContentType  string
}

type proxyResponse struct {
	StatusCode       int
	Headers          http.Header
	Body             []byte
	ContentType      string
	TransformBody    []byte
	TransformHeaders map[string]string
}

func newHTTPClient() *http.Client {
	return &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableCompression: true}}
}

func (h *Handler) forward(ctx context.Context, incoming *http.Request, endpoint string, input requestData) (proxyResponse, error) {
	upstream, err := url.Parse(endpoint)
	if err != nil || upstream.Scheme == "" || upstream.Host == "" {
		return proxyResponse{}, &url.Error{Op: "parse", URL: endpoint, Err: err}
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
		return proxyResponse{}, err
	}
	copyRequestHeaders(req.Header, incoming.Header, input.Headers)
	req.ContentLength = int64(len(input.Body))
	resp, err := h.client.Do(req)
	if err != nil {
		return proxyResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return proxyResponse{}, err
	}
	transformHeaders := flattenHeaders(resp.Header)
	return proxyResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: body, ContentType: resp.Header.Get("Content-Type"), TransformBody: transformPayload(body, resp.Header.Get("Content-Type")), TransformHeaders: transformHeaders}, nil
}

func writeResponse(w http.ResponseWriter, response proxyResponse) {
	copyResponseHeaders(w.Header(), response.Headers)
	w.WriteHeader(response.StatusCode)
	if _, err := w.Write(response.Body); err != nil {
		return
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	payload, err := json.Marshal(errorEnvelope{Success: false, Error: message, Code: code})
	if err != nil {
		h.logger.Error("gateway error response encoding failed", "err", err)
		return
	}
	if _, err := w.Write(append(payload, '\n')); err != nil {
		h.logger.Error("gateway error response write failed", "err", err)
	}
}

type errorEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

func methodAllowed(configured, method string) bool {
	configured = strings.TrimSpace(strings.ToUpper(configured))
	method = strings.TrimSpace(strings.ToUpper(method))
	return configured == "ALL" || configured == method
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	return r.RemoteAddr
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

var hopByHopHeaders = map[string]struct{}{
	"connection": {}, "keep-alive": {}, "proxy-authenticate": {}, "proxy-authorization": {},
	"proxy-connection": {}, "te": {}, "trailer": {}, "transfer-encoding": {}, "upgrade": {},
	"content-length": {},
}

func copyRequestHeaders(dst, src http.Header, transformed map[string]string) {
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

func copyResponseHeaders(dst, src http.Header) {
	blocked := connectionHeaders(src)
	for key, values := range src {
		if isHopByHop(strings.ToLower(key)) || blocked[strings.ToLower(key)] {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
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

func isHopByHop(key string) bool {
	if _, ok := hopByHopHeaders[key]; ok {
		return true
	}
	return false
}
