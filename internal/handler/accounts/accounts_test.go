package accounts

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestRequestMetadata_isValidJSONAndUsesRequestContext(t *testing.T) {
	request := httptest.NewRequest("PUT", "/api/tokens/id", nil)
	request.RemoteAddr = "192.0.2.10:443"
	request.Header.Set("User-Agent", "accounts-test")
	metadata := requestMetadata(request)
	if metadata.IP != "192.0.2.10" || metadata.UserAgent != "accounts-test" {
		t.Fatalf("metadata = %+v", metadata)
	}
	if !json.Valid([]byte(metadata.Metadata)) {
		t.Fatalf("metadata is not JSON: %s", metadata.Metadata)
	}
}
