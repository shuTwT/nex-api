package proxy

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const legacyMaxResponseBodyBytes = 64 << 20

func TestForwardReadsResponseBeyondLegacyLimit(t *testing.T) {
	responseSize := int64(legacyMaxResponseBodyBytes + 1)
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"application/json"}},
			Body:       io.NopCloser(&repeatedByteReader{remaining: responseSize, value: 'x'}),
		}, nil
	})}
	incoming := httptest.NewRequest(http.MethodGet, "http://gateway.example/request", nil)

	response, err := Forward(context.Background(), client, incoming, "http://upstream.example/response", RequestData{})
	if err != nil {
		t.Fatalf("Forward() error = %v", err)
	}
	if got := int64(len(response.Body)); got != responseSize {
		t.Errorf("response body length = %d, want %d", got, responseSize)
	}
	if got := int64(len(response.TransformBody)); got != responseSize {
		t.Errorf("transform body length = %d, want %d", got, responseSize)
	}
	if got := response.Body[len(response.Body)-1]; got != 'x' {
		t.Errorf("last response byte = %q, want %q", got, 'x')
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type repeatedByteReader struct {
	remaining int64
	value     byte
}

func (r *repeatedByteReader) Read(payload []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	count := min(int64(len(payload)), r.remaining)
	for index := range payload[:count] {
		payload[index] = r.value
	}
	r.remaining -= count
	return int(count), nil
}
