package gateway

import (
	"encoding/json"
	"net/http"
	"strings"
)

func transformRequest(r *http.Request, body []byte) requestData {
	query := make(map[string]string, len(r.URL.Query()))
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			query[key] = values[len(values)-1]
		}
	}
	return requestData{Headers: flattenHeaders(r.Header), Query: query, Body: body, ContentType: r.Header.Get("Content-Type")}
}

func transformPayload(body []byte, contentType string) []byte {
	if len(body) == 0 {
		return nil
	}
	if isJSONContentType(contentType) {
		return append([]byte(nil), body...)
	}
	encoded, err := json.Marshal(string(body))
	if err != nil {
		return nil
	}
	return encoded
}

func transformedBytes(body []byte, contentType string) []byte {
	if isJSONContentType(contentType) || len(body) == 0 {
		return append([]byte(nil), body...)
	}
	var text string
	if err := json.Unmarshal(body, &text); err == nil {
		return []byte(text)
	}
	return append([]byte(nil), body...)
}

func isJSONContentType(contentType string) bool {
	mediaType := strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0]))
	return strings.HasPrefix(mediaType, "application/json") || strings.HasSuffix(mediaType, "+json")
}

func flattenHeaders(headers http.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[len(values)-1]
		}
	}
	return result
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
