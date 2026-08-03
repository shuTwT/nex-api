package catalog

import (
	"strings"

	servicecatalog "github.com/shuTwT/nex-api/backend/internal/service/catalog"
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

type categoryData = servicecatalog.CategoryDTO
type apiData = servicecatalog.APIDTO
type mcpData = servicecatalog.MCPDTO
type categoryListData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	APICount    int     `json:"apiCount"`
}

func toCategoryData(item any) categoryData { return servicecatalog.CategoryDTOFromEntity(item) }

func toCategoryListData(item servicecatalog.CategoryListItem) categoryListData {
	return categoryListData{ID: item.Category.ID, Name: item.Category.Name, Description: item.Category.Description, Icon: optionalString(item.Category.Icon), APICount: item.APICount}
}

func toAPIData(item any, includeRelations bool) apiData {
	return servicecatalog.APIDTOFromEntity(item, includeRelations)
}
func toMCPData(item any) mcpData { return servicecatalog.MCPDTOFromEntity(item) }

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	copy := value
	return &copy
}
