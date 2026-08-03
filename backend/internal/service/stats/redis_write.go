package stats

import "time"

func (s *Store) requestKeys(event RequestEvent) []string {
	keys := []string{s.matrix.GlobalRequests(), legacyGlobalRequests}
	if isMCPAlias(event.Alias) {
		identifier := mcpIdentifier(event.Alias)
		keys = append(keys, s.matrix.MCPRequests(identifier))
		if event.UserID != "" {
			keys = append(keys, s.matrix.UserMCPRequests(event.UserID, identifier))
		}
		return keys
	}
	keys = append(keys, s.matrix.APIRequests(event.Alias), legacyAPIKey(event.Alias))
	if event.UserID != "" {
		keys = append(keys,
			s.matrix.UserAPIRequests(event.UserID, event.Alias),
			legacyUserAPIKey(event.UserID, event.Alias),
		)
	}
	return keys
}

func (s *Store) usageKeys(event RequestEvent, hour time.Time) []string {
	keys := []string{s.matrix.GlobalCredits(hour)}
	if isMCPAlias(event.Alias) {
		keys = append(keys, s.matrix.MCPCredits(mcpIdentifier(event.Alias), hour))
		if event.UserID != "" {
			keys = append(keys, s.matrix.UserCredits(event.UserID, hour), s.matrix.UserMCPCredits(event.UserID, mcpIdentifier(event.Alias), hour))
		}
		return keys
	}
	keys = append(keys, s.matrix.APICredits(event.Alias, hour))
	if event.UserID != "" {
		keys = append(keys, s.matrix.UserCredits(event.UserID, hour), s.matrix.UserAPICredits(event.UserID, event.Alias, hour))
	}
	return keys
}

func (s *Store) trendKeys(event RequestEvent, hour time.Time) []string {
	keys := []string{s.matrix.GlobalRequestTrend(hour)}
	if isMCPAlias(event.Alias) {
		identifier := mcpIdentifier(event.Alias)
		keys = append(keys, s.matrix.MCPRequestTrend(identifier, hour))
		if event.UserID != "" {
			keys = append(keys, s.matrix.UserMCPRequestTrend(event.UserID, identifier, hour))
		}
		return keys
	}
	keys = append(keys, s.matrix.APIRequestTrend(event.Alias, hour))
	if event.UserID != "" {
		keys = append(keys, s.matrix.UserAPIRequestTrend(event.UserID, event.Alias, hour))
	}
	return keys
}
