package model

type MarketplaceAPIResp struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Alias          string `json:"alias"`
	Endpoint       string `json:"endpoint"`
	Method         string `json:"method"`
	Pricing        int    `json:"pricing"`
	Category       string `json:"category"`
	IsFree         bool   `json:"isFree"`
	IsActive       bool   `json:"isActive,omitempty"`
	TodayCallCount int64  `json:"todayCallCount"`
	UserCount      int    `json:"userCount"`
	TotalCallCount int64  `json:"totalCallCount"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}
type MarketplaceMCPResp struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Identifier     string `json:"identifier"`
	Category       string `json:"category"`
	Description    string `json:"description"`
	Documentation  string `json:"documentation,omitempty"`
	Type           string `json:"type"`
	Pricing        int    `json:"pricing"`
	IsFree         bool   `json:"isFree"`
	IsActive       bool   `json:"isActive,omitempty"`
	TodayCallCount int64  `json:"todayCallCount"`
	UserCount      int    `json:"userCount"`
	TotalCallCount int64  `json:"totalCallCount"`
	CreatedAt      string `json:"createdAt,omitempty"`
	UpdatedAt      string `json:"updatedAt,omitempty"`
}

type MarketplaceMCPTool struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
}
