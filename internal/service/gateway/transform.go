package gateway

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/shuTwT/nex-api/internal/infra/proxy"
	"github.com/shuTwT/nex-api/internal/infra/worker"
)

// Transformer executes script transforms on the worker engine.
type Transformer interface {
	Execute(context.Context, worker.Job) (worker.TransformResult, error)
}

type RequestData struct {
	Headers      map[string]string
	Query        map[string]string
	QueryChanged bool
	Body         []byte
	ContentType  string
}

type ResponseData struct {
	StatusCode       int
	Headers          map[string][]string
	Body             []byte
	ContentType      string
	TransformBody    []byte
	TransformHeaders map[string]string
}

type ForwardRequest struct {
	Method  string
	Headers map[string][]string
	Query   map[string][]string
	Body    []byte
	Data    RequestData
}

type Forwarder interface {
	Forward(context.Context, string, ForwardRequest) (ResponseData, error)
}

// ProxyForwarder is the service-owned adapter for the infrastructure proxy.
type ProxyForwarder struct{ delegate *proxy.GatewayForwarder }

func NewProxyForwarder() *ProxyForwarder {
	return &ProxyForwarder{delegate: proxy.NewGatewayForwarder(nil)}
}

func (f *ProxyForwarder) Forward(ctx context.Context, endpoint string, request ForwardRequest) (ResponseData, error) {
	response, err := f.delegate.Forward(ctx, endpoint, proxy.GatewayRequest{
		Method: request.Method, Headers: request.Headers, Query: request.Query, Body: request.Body,
		Data: proxy.RequestData{Headers: request.Data.Headers, Query: request.Data.Query, QueryChanged: request.Data.QueryChanged, Body: request.Data.Body, ContentType: request.Data.ContentType},
	})
	if err != nil {
		return ResponseData{}, err
	}
	return ResponseData{StatusCode: response.StatusCode, Headers: response.Headers, Body: response.Body, ContentType: response.ContentType, TransformBody: response.TransformBody, TransformHeaders: response.TransformHeaders}, nil
}

// TransformService owns the pre/post script transform orchestration for API
// gateway requests.
type TransformService struct {
	executor Transformer
}

func NewTransformService(executor Transformer) *TransformService {
	return &TransformService{executor: executor}
}

// PreTransform runs the pre-request script over the normalized request data.
func (s *TransformService) PreTransform(ctx context.Context, script string, input RequestData) (RequestData, error) {
	if s == nil || s.executor == nil {
		return RequestData{}, fmt.Errorf("gateway: transform worker is unavailable")
	}
	result, err := s.executor.Execute(ctx, worker.Job{ID: uuid.NewString(), Kind: worker.ScriptKindPre, Script: script, Headers: input.Headers, Query: input.Query, Body: proxy.TransformPayload(input.Body, input.ContentType)})
	if err != nil {
		return RequestData{}, fmt.Errorf("pre-transform: %w", err)
	}
	mergeStrings(input.Headers, result.Headers)
	mergeQuery(input.Query, result.Query)
	input.QueryChanged = true
	if result.BodySet {
		input.Body = proxy.TransformedBytes(result.Body, input.ContentType)
	}
	return input, nil
}

// PostTransform runs the response script over the upstream response.
func (s *TransformService) PostTransform(ctx context.Context, script string, response ResponseData) (ResponseData, error) {
	if s == nil || s.executor == nil {
		return ResponseData{}, fmt.Errorf("gateway: transform worker is unavailable")
	}
	result, err := s.executor.Execute(ctx, worker.Job{ID: uuid.NewString(), Kind: worker.ScriptKindPost, Script: script, ResponseBody: response.TransformBody, ResponseHeaders: response.TransformHeaders})
	if err != nil {
		return ResponseData{}, fmt.Errorf("post-transform: %w", err)
	}
	for key, value := range result.ResponseHeaders {
		response.Headers[key] = []string{value}
	}
	if result.ResponseBodySet {
		response.Body = proxy.TransformedBytes(result.ResponseBody, response.ContentType)
	}
	return response, nil
}

// TransformRequest builds the normalized request data for scripts.
func TransformRequest(headers, queryValues map[string][]string, body []byte) RequestData {
	query := make(map[string]string, len(queryValues))
	for key, values := range queryValues {
		if len(values) > 0 {
			query[key] = values[len(values)-1]
		}
	}
	flattenedHeaders := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			flattenedHeaders[key] = values[len(values)-1]
		}
	}
	return RequestData{Headers: flattenedHeaders, Query: query, Body: body, ContentType: flattenedHeaders["Content-Type"]}
}

func mergeStrings(base, updates map[string]string) {
	for key, value := range updates {
		base[key] = value
	}
}

func mergeQuery(base, updates map[string]string) {
	for key, value := range updates {
		base[key] = value
	}
}
