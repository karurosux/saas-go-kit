package {{.ModuleName}}model

type Create{{.ModuleNameCap}}Request struct {
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}

type Update{{.ModuleNameCap}}Request struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}

type {{.ModuleNameCap}}Query struct {
	Name     *string `query:"name"`
	IsActive *bool   `query:"is_active"`
	Page     int     `query:"page" validate:"min=1"`
	Limit    int     `query:"limit" validate:"min=1,max=100"`
}