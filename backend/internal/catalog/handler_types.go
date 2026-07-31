package catalog

import (
	"strings"

	"github.com/shuTwT/nex-api/backend/internal/database/ent"
)

type apiRequest struct {
	Name          string  `json:"name"`
	Alias         string  `json:"alias"`
	Description   string  `json:"description"`
	Endpoint      string  `json:"endpoint"`
	Method        string  `json:"method"`
	CategoryID    string  `json:"categoryId"`
	Pricing       *int    `json:"pricing"`
	Documentation *string `json:"documentation"`
	PreScript     *string `json:"preScript"`
	PostScript    *string `json:"postScript"`
	IsActive      *bool   `json:"isActive"`
}

type apiUpdateRequest struct {
	Name          *string `json:"name"`
	Alias         *string `json:"alias"`
	Description   *string `json:"description"`
	Endpoint      *string `json:"endpoint"`
	Method        *string `json:"method"`
	CategoryID    *string `json:"categoryId"`
	Pricing       *int    `json:"pricing"`
	Documentation *string `json:"documentation"`
	PreScript     *string `json:"preScript"`
	PostScript    *string `json:"postScript"`
	IsActive      *bool   `json:"isActive"`
}

type categoryRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}

type categoryUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}

type mcpRequest struct {
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Type       string  `json:"type"`
	Command    *string `json:"command"`
	Endpoint   *string `json:"endpoint"`
	EnvVars    *string `json:"envVars"`
	Pricing    *int    `json:"pricing"`
	IsActive   *bool   `json:"isActive"`
}

type mcpUpdateRequest struct {
	Name       *string `json:"name"`
	Identifier *string `json:"identifier"`
	Type       *string `json:"type"`
	Command    *string `json:"command"`
	Endpoint   *string `json:"endpoint"`
	EnvVars    *string `json:"envVars"`
	Pricing    *int    `json:"pricing"`
	IsActive   *bool   `json:"isActive"`
}

type categoryData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}

type apiData struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Alias         string              `json:"alias"`
	Description   string              `json:"description"`
	Endpoint      string              `json:"endpoint"`
	Method        string              `json:"method"`
	CategoryID    string              `json:"categoryId"`
	Category      *categoryData       `json:"category,omitempty"`
	Pricing       int                 `json:"pricing"`
	Documentation *string             `json:"documentation"`
	PreScript     *string             `json:"preScript"`
	PostScript    *string             `json:"postScript"`
	IsActive      bool                `json:"isActive"`
	CallCount     int                 `json:"callCount"`
	CreatedAt     string              `json:"createdAt"`
	UpdatedAt     string              `json:"updatedAt"`
	Parameters    []*ent.ApiParameter `json:"parameters,omitempty"`
	Responses     []*ent.ApiResponse  `json:"responses,omitempty"`
}

type mcpData struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Type       string  `json:"type"`
	Command    *string `json:"command"`
	Endpoint   *string `json:"endpoint"`
	EnvVars    *string `json:"envVars"`
	Pricing    int     `json:"pricing"`
	IsActive   bool    `json:"isActive"`
	CallCount  int     `json:"callCount"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

type categoryListData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	APICount    int     `json:"apiCount"`
}

func toCategoryData(item *ent.ApiCategory) categoryData {
	return categoryData{ID: item.ID, Name: item.Name, Description: item.Description, Icon: optionalString(item.Icon)}
}

func toCategoryListData(item CategoryListItem) categoryListData {
	return categoryListData{ID: item.Category.ID, Name: item.Category.Name, Description: item.Category.Description, Icon: optionalString(item.Category.Icon), APICount: item.APICount}
}

func toAPIData(item *ent.Api, includeRelations bool) apiData {
	data := apiData{ID: item.ID, Name: item.Name, Alias: item.Alias, Description: item.Description, Endpoint: item.Endpoint, Method: item.Method, CategoryID: item.CategoryId, Pricing: item.Pricing, Documentation: optionalString(item.Documentation), PreScript: optionalString(item.PreScript), PostScript: optionalString(item.PostScript), IsActive: item.IsActive, CallCount: item.CallCount, CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"), UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z")}
	if item.Edges.Category != nil {
		data.Category = &categoryData{ID: item.Edges.Category.ID, Name: item.Edges.Category.Name, Description: item.Edges.Category.Description, Icon: optionalString(item.Edges.Category.Icon)}
	}
	if includeRelations {
		data.Parameters = item.Edges.Parameters
		data.Responses = item.Edges.Responses
	}
	return data
}

func toMCPData(item *ent.McpService) mcpData {
	return mcpData{ID: item.ID, Name: item.Name, Identifier: item.Identifier, Type: item.Type, Command: optionalString(item.Command), Endpoint: optionalString(item.Endpoint), EnvVars: optionalString(item.EnvVars), Pricing: item.Pricing, IsActive: item.IsActive, CallCount: item.CallCount, CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"), UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z")}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	copy := value
	return &copy
}
