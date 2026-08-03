package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GatewayRequest and GatewayResponse keep the proxy adapter boundary free of
// net/http values. Services map their application DTOs to these values.
type GatewayRequest struct {
	Method  string
	Headers map[string][]string
	Query   map[string][]string
	Body    []byte
	Data    RequestData
}

type GatewayResponse struct {
	StatusCode       int
	Headers          map[string][]string
	Body             []byte
	ContentType      string
	TransformBody    []byte
	TransformHeaders map[string]string
}

type GatewayForwarder struct{ client *http.Client }

func NewGatewayForwarder(client *http.Client) *GatewayForwarder {
	if client == nil {
		client = NewHTTPClient()
	}
	return &GatewayForwarder{client: client}
}

func (f *GatewayForwarder) Forward(ctx context.Context, endpoint string, incoming GatewayRequest) (GatewayResponse, error) {
	upstream, err := url.Parse(endpoint)
	if err != nil || upstream.Scheme == "" || upstream.Host == "" {
		return GatewayResponse{}, &url.Error{Op: "parse", URL: endpoint, Err: err}
	}
	query := upstream.Query()
	for key, values := range incoming.Query {
		query.Del(key)
		for _, value := range values {
			query.Add(key, value)
		}
	}
	if incoming.Data.QueryChanged {
		for key, value := range incoming.Data.Query {
			query.Set(key, value)
		}
	}
	upstream.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, incoming.Method, upstream.String(), bytes.NewReader(incoming.Data.Body))
	if err != nil {
		return GatewayResponse{}, err
	}
	CopyRequestHeadersWithTransforms(request.Header, incoming.Headers, incoming.Data.Headers)
	request.ContentLength = int64(len(incoming.Data.Body))
	response, err := f.client.Do(request)
	if err != nil {
		return GatewayResponse{}, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return GatewayResponse{}, fmt.Errorf("read upstream response: %w", err)
	}
	headers := make(map[string][]string, len(response.Header))
	for key, values := range response.Header {
		headers[key] = append([]string(nil), values...)
	}
	return GatewayResponse{StatusCode: response.StatusCode, Headers: headers, Body: body, ContentType: response.Header.Get("Content-Type"), TransformBody: TransformPayload(body, response.Header.Get("Content-Type")), TransformHeaders: FlattenHeaders(response.Header)}, nil
}
