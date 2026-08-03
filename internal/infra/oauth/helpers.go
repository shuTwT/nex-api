package oauth

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

// AccountTokensFromOAuth converts provider token exchange results into the
// persisted account token shape.
func AccountTokensFromOAuth(tokens OAuthTokens) AccountTokens {
	return AccountTokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		Scope:        tokens.Scope,
		IDToken:      tokens.IDToken,
	}
}

// NormalizedScopes normalizes comma/space separated scope values.
func NormalizedScopes(raw string) string {
	parts := strings.FieldsFunc(raw, func(character rune) bool {
		return character == ',' || unicode.IsSpace(character)
	})
	return strings.Join(parts, " ")
}

// StringAttribute reads a string attribute from a nested map.
func StringAttribute(attributes map[string]any, preferred string, fallbacks ...string) string {
	for _, field := range append([]string{preferred}, fallbacks...) {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if value, ok := NestedAttribute(attributes, field); ok {
			switch typed := value.(type) {
			case string:
				if value := strings.TrimSpace(typed); value != "" {
					return value
				}
			case json.Number:
				return typed.String()
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			}
		}
	}
	return ""
}

// BoolAttribute reads a boolean attribute from a nested map.
func BoolAttribute(attributes map[string]any, fields ...string) (bool, bool) {
	for _, field := range fields {
		value, ok := NestedAttribute(attributes, field)
		if !ok {
			continue
		}
		if typed, ok := value.(bool); ok {
			return typed, true
		}
	}
	return false, false
}

// NestedAttribute resolves a dotted path inside a JSON object map.
func NestedAttribute(attributes map[string]any, path string) (any, bool) {
	var current any = attributes
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}
