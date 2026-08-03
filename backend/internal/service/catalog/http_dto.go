package catalog

import (
	"github.com/shuTwT/nex-api/backend/ent"
	"strings"
)

type CategoryDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}
type ParameterDTO struct {
	ID           string `json:"id,omitempty"`
	ApiID        string `json:"apiId,omitempty"`
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}
type ResponseDTO struct {
	ID          string `json:"id,omitempty"`
	ApiID       string `json:"apiId,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}
type APIDTO struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Alias         string         `json:"alias"`
	Description   string         `json:"description"`
	Endpoint      string         `json:"endpoint"`
	Method        string         `json:"method"`
	CategoryID    string         `json:"categoryId"`
	Category      *CategoryDTO   `json:"category,omitempty"`
	Pricing       int            `json:"pricing"`
	Documentation *string        `json:"documentation"`
	PreScript     *string        `json:"preScript"`
	PostScript    *string        `json:"postScript"`
	IsActive      bool           `json:"isActive"`
	CallCount     int            `json:"callCount"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
	Parameters    []ParameterDTO `json:"parameters,omitempty"`
	Responses     []ResponseDTO  `json:"responses,omitempty"`
}
type MCPDTO struct {
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

func CategoryDTOFromEntity(value any) CategoryDTO {
	item := value.(*ent.ApiCategory)
	return CategoryDTO{ID: item.ID, Name: item.Name, Description: item.Description, Icon: optionalDTO(item.Icon)}
}
func APIDTOFromEntity(value any, relations bool) APIDTO {
	item := value.(*ent.Api)
	data := APIDTO{ID: item.ID, Name: item.Name, Alias: item.Alias, Description: item.Description, Endpoint: item.Endpoint, Method: item.Method, CategoryID: item.CategoryId, Pricing: item.Pricing, Documentation: optionalDTO(item.Documentation), PreScript: optionalDTO(item.PreScript), PostScript: optionalDTO(item.PostScript), IsActive: item.IsActive, CallCount: item.CallCount, CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"), UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z")}
	if item.Edges.Category != nil {
		category := CategoryDTOFromEntity(item.Edges.Category)
		data.Category = &category
	}
	if relations {
		for _, p := range item.Edges.Parameters {
			data.Parameters = append(data.Parameters, ParameterDTO{ID: p.ID, ApiID: p.ApiId, Name: p.Name, Type: p.Type, Required: p.Required, Description: p.Description, DefaultValue: p.DefaultValue})
		}
		for _, r := range item.Edges.Responses {
			data.Responses = append(data.Responses, ResponseDTO{ID: r.ID, ApiID: r.ApiId, Name: r.Name, Type: r.Type, Description: r.Description})
		}
	}
	return data
}
func MCPDTOFromEntity(value any) MCPDTO {
	item := value.(*ent.McpService)
	return MCPDTO{ID: item.ID, Name: item.Name, Identifier: item.Identifier, Type: item.Type, Command: optionalDTO(item.Command), Endpoint: optionalDTO(item.Endpoint), EnvVars: optionalDTO(item.EnvVars), Pricing: item.Pricing, IsActive: item.IsActive, CallCount: item.CallCount, CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"), UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z")}
}
func optionalDTO(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	copied := value
	return &copied
}
