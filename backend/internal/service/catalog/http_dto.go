package catalog

import (
	"github.com/shuTwT/nex-api/backend/ent"
	"github.com/shuTwT/nex-api/backend/pkg/domain/model"
	"strings"
)

type CategoryDTO = model.CatalogCategoryDTO
type ParameterDTO = model.CatalogParameterDTO
type ResponseDTO = model.CatalogResponseDTO
type APIDTO = model.CatalogAPIDTO
type MCPDTO = model.CatalogMCPDTO

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
			data.Parameters = append(data.Parameters, ParameterDTO{ID: p.ID, APIID: p.ApiId, Name: p.Name, Type: p.Type, Required: p.Required, Description: p.Description, DefaultValue: p.DefaultValue})
		}
		for _, r := range item.Edges.Responses {
			data.Responses = append(data.Responses, ResponseDTO{ID: r.ID, APIID: r.ApiId, Name: r.Name, Type: r.Type, Description: r.Description})
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
