package proxy

import (
	"encoding/json"
	"net/http"
	"strings"
)

// TransformPayload encodes a non-JSON body as a JSON string so scripts always
// receive a stable JSON representation.
func TransformPayload(body []byte, contentType string) []byte {
	if len(body) == 0 {
		return nil
	}
	if IsJSONContentType(contentType) {
		return append([]byte(nil), body...)
	}
	encoded, err := json.Marshal(string(body))
	if err != nil {
		return nil
	}
	return encoded
}

// TransformedBytes decodes a script-produced JSON string back to raw bytes.
func TransformedBytes(body []byte, contentType string) []byte {
	if IsJSONContentType(contentType) || len(body) == 0 {
		return append([]byte(nil), body...)
	}
	var text string
	if err := json.Unmarshal(body, &text); err == nil {
		return []byte(text)
	}
	return append([]byte(nil), body...)
}

// IsJSONContentType reports whether the media type is JSON.
func IsJSONContentType(contentType string) bool {
	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	return strings.HasPrefix(mediaType, "application/json") || strings.HasSuffix(mediaType, "+json")
}

// FlattenHeaders collapses multi-value headers to their last value.
func FlattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[len(values)-1]
		}
	}
	return result
}
