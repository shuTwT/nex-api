package model

type CatalogAPICreateReq struct {
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
type CatalogAPIUpdateReq struct {
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
type CatalogCategoryCreateReq struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}
type CatalogCategoryUpdateReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
}
type CatalogMCPCreateReq struct {
	Name       string  `json:"name"`
	Identifier string  `json:"identifier"`
	Type       string  `json:"type"`
	Command    *string `json:"command"`
	Endpoint   *string `json:"endpoint"`
	EnvVars    *string `json:"envVars"`
	Pricing    *int    `json:"pricing"`
	IsActive   *bool   `json:"isActive"`
}
type CatalogMCPUpdateReq struct {
	Name       *string `json:"name"`
	Identifier *string `json:"identifier"`
	Type       *string `json:"type"`
	Command    *string `json:"command"`
	Endpoint   *string `json:"endpoint"`
	EnvVars    *string `json:"envVars"`
	Pricing    *int    `json:"pricing"`
	IsActive   *bool   `json:"isActive"`
}

type CatalogCategoryDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}
type CatalogParameterDTO struct {
	ID           string `json:"id,omitempty"`
	APIID        string `json:"apiId,omitempty"`
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Required     bool   `json:"required,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultValue string `json:"defaultValue,omitempty"`
}
type CatalogResponseDTO struct {
	ID          string `json:"id,omitempty"`
	APIID       string `json:"apiId,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}
type CatalogAPIDTO struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Alias         string                `json:"alias"`
	Description   string                `json:"description"`
	Endpoint      string                `json:"endpoint"`
	Method        string                `json:"method"`
	CategoryID    string                `json:"categoryId"`
	Category      *CatalogCategoryDTO   `json:"category,omitempty"`
	Pricing       int                   `json:"pricing"`
	Documentation *string               `json:"documentation"`
	PreScript     *string               `json:"preScript"`
	PostScript    *string               `json:"postScript"`
	IsActive      bool                  `json:"isActive"`
	CallCount     int                   `json:"callCount"`
	CreatedAt     string                `json:"createdAt"`
	UpdatedAt     string                `json:"updatedAt"`
	Parameters    []CatalogParameterDTO `json:"parameters,omitempty"`
	Responses     []CatalogResponseDTO  `json:"responses,omitempty"`
}
type CatalogMCPDTO struct {
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
type CatalogCategoryListResp struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	APICount    int     `json:"apiCount"`
}
