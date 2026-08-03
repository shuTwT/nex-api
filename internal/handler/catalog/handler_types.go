package catalog

import (
	"strings"

	servicecatalog "github.com/shuTwT/nex-api/internal/service/catalog"
	"github.com/shuTwT/nex-api/pkg/domain/model"
)

type apiRequest = model.CatalogAPICreateReq
type apiUpdateRequest = model.CatalogAPIUpdateReq
type categoryRequest = model.CatalogCategoryCreateReq
type categoryUpdateRequest = model.CatalogCategoryUpdateReq
type mcpRequest = model.CatalogMCPCreateReq
type mcpUpdateRequest = model.CatalogMCPUpdateReq
type categoryData = model.CatalogCategoryDTO
type apiData = model.CatalogAPIDTO
type mcpData = model.CatalogMCPDTO
type categoryListData = model.CatalogCategoryListResp

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
