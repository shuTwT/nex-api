package stats

import (
	"fmt"
	"strings"
	"time"
)

const (
	legacyGlobalRequests = "global:request:count"
	legacyAPIRequests    = "api:request:count:"
	legacyUserRequests   = "user:api:request:count:"

	statsPrefix = "v1:stats:"
	usagePrefix = "v1:usage:"
	trendPrefix = "v1:trend:"
)

const (
	MCPAliasPrefix = "mcp:"
)

type RequestEvent struct {
	UserID  string
	Alias   string
	Credits float64
	At      time.Time
}

type UserAPIKey struct {
	UserID string
	Alias  string
}

type UserMCPKey struct {
	UserID     string
	Identifier string
}

type KeyMatrix struct{}

func NewKeyMatrix() KeyMatrix { return KeyMatrix{} }

func (KeyMatrix) GlobalRequests() string { return statsPrefix + "global:requests" }

func (KeyMatrix) APIRequests(alias string) string {
	return statsPrefix + "api:" + alias + ":requests"
}

func (KeyMatrix) MCPRequests(identifier string) string {
	return statsPrefix + "mcp:" + identifier + ":requests"
}

func (KeyMatrix) UserAPIRequests(userID, alias string) string {
	return statsPrefix + "user:" + userID + ":api:" + alias + ":requests"
}

func (KeyMatrix) UserMCPRequests(userID, identifier string) string {
	return statsPrefix + "user:" + userID + ":mcp:" + identifier + ":requests"
}

func (KeyMatrix) GlobalCredits(hour time.Time) string {
	return usagePrefix + "global:credits:" + hourKey(hour)
}

func (KeyMatrix) APICredits(alias string, hour time.Time) string {
	return usagePrefix + "api:" + alias + ":credits:" + hourKey(hour)
}

func (KeyMatrix) MCPCredits(identifier string, hour time.Time) string {
	return usagePrefix + "mcp:" + identifier + ":credits:" + hourKey(hour)
}

func (KeyMatrix) UserCredits(userID string, hour time.Time) string {
	return usagePrefix + "user:" + userID + ":credits:" + hourKey(hour)
}

func (KeyMatrix) UserAPICredits(userID, alias string, hour time.Time) string {
	return usagePrefix + "user:" + userID + ":api:" + alias + ":credits:" + hourKey(hour)
}

func (KeyMatrix) UserMCPCredits(userID, identifier string, hour time.Time) string {
	return usagePrefix + "user:" + userID + ":mcp:" + identifier + ":credits:" + hourKey(hour)
}

func (KeyMatrix) GlobalRequestTrend(hour time.Time) string {
	return trendPrefix + "global:requests:" + hourKey(hour)
}

func (KeyMatrix) APIRequestTrend(alias string, hour time.Time) string {
	return trendPrefix + "api:" + alias + ":requests:" + hourKey(hour)
}

func (KeyMatrix) MCPRequestTrend(identifier string, hour time.Time) string {
	return trendPrefix + "mcp:" + identifier + ":requests:" + hourKey(hour)
}

func (KeyMatrix) UserAPIRequestTrend(userID, alias string, hour time.Time) string {
	return trendPrefix + "user:" + userID + ":api:" + alias + ":requests:" + hourKey(hour)
}

func (KeyMatrix) UserMCPRequestTrend(userID, identifier string, hour time.Time) string {
	return trendPrefix + "user:" + userID + ":mcp:" + identifier + ":requests:" + hourKey(hour)
}

func legacyAPIKey(alias string) string { return legacyAPIRequests + alias }

func legacyUserAPIKey(userID, alias string) string {
	return legacyUserRequests + userID + ":" + alias
}

func isMCPAlias(alias string) bool { return strings.HasPrefix(alias, MCPAliasPrefix) }

func mcpIdentifier(alias string) string { return strings.TrimPrefix(alias, MCPAliasPrefix) }

func hourKey(value time.Time) string {
	return fmt.Sprintf("%d", value.UTC().Truncate(time.Hour).Unix())
}
