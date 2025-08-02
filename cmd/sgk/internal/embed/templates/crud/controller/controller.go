package {{.ModuleName}}controller

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"{{.Project.GoModule}}/internal/core"
	{{.ModuleName}}interface "{{.Project.GoModule}}/internal/{{.ModuleName}}/interface"
	{{.ModuleName}}model "{{.Project.GoModule}}/internal/{{.ModuleName}}/model"
)

type {{.ModuleNameCap}}Controller struct {
	service   {{.ModuleName}}interface.{{.ModuleNameCap}}Service
	validator *core.Validator
}

func New{{.ModuleNameCap}}Controller(service {{.ModuleName}}interface.{{.ModuleNameCap}}Service) *{{.ModuleNameCap}}Controller {
	return &{{.ModuleNameCap}}Controller{
		service:   service,
		validator: core.NewValidator(),
	}
}

func (c *{{.ModuleNameCap}}Controller) RegisterRoutes(e *echo.Echo, prefix string) {
	g := e.Group(prefix)
	g.POST("", c.Create)
	g.GET("", c.List)
	g.GET("/:id", c.GetByID)
	g.PUT("/:id", c.Update)
	g.DELETE("/:id", c.Delete)
}

func (c *{{.ModuleNameCap}}Controller) Create(ctx echo.Context) error {
	var req {{.ModuleName}}model.Create{{.ModuleNameCap}}Request
	if err := ctx.Bind(&req); err != nil {
		return core.BadRequest(ctx, err)
	}

	if err := c.validator.Validate(&req); err != nil {
		return core.BadRequest(ctx, err)
	}

	{{.ModuleName}}, err := c.service.Create(ctx.Request().Context(), req)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}

	return core.Created(ctx, {{.ModuleName}}, "{{.ModuleNameCap}} created successfully")
}

func (c *{{.ModuleNameCap}}Controller) GetByID(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return core.BadRequest(ctx, err)
	}

	{{.ModuleName}}, err := c.service.GetByID(ctx.Request().Context(), uint(id))
	if err != nil {
		return core.NotFound(ctx, err)
	}

	return core.Success(ctx, {{.ModuleName}}, "{{.ModuleNameCap}} retrieved successfully")
}

func (c *{{.ModuleNameCap}}Controller) List(ctx echo.Context) error {
	var query {{.ModuleName}}model.{{.ModuleNameCap}}Query
	if err := ctx.Bind(&query); err != nil {
		return core.BadRequest(ctx, err)
	}

	if err := c.validator.Validate(&query); err != nil {
		return core.BadRequest(ctx, err)
	}

	{{.ModuleName}}s, total, err := c.service.List(ctx.Request().Context(), query)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}

	response := map[string]interface{}{
		"data":  {{.ModuleName}}s,
		"total": total,
		"page":  query.Page,
		"limit": query.Limit,
	}

	return core.Success(ctx, response, "{{.ModuleNameCap}}s retrieved successfully")
}

func (c *{{.ModuleNameCap}}Controller) Update(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return core.BadRequest(ctx, err)
	}

	var req {{.ModuleName}}model.Update{{.ModuleNameCap}}Request
	if err := ctx.Bind(&req); err != nil {
		return core.BadRequest(ctx, err)
	}

	if err := c.validator.Validate(&req); err != nil {
		return core.BadRequest(ctx, err)
	}

	{{.ModuleName}}, err := c.service.Update(ctx.Request().Context(), uint(id), req)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}

	return core.Success(ctx, {{.ModuleName}}, "{{.ModuleNameCap}} updated successfully")
}

func (c *{{.ModuleNameCap}}Controller) Delete(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return core.BadRequest(ctx, err)
	}

	if err := c.service.Delete(ctx.Request().Context(), uint(id)); err != nil {
		return core.InternalServerError(ctx, err)
	}

	return core.Success(ctx, nil, "{{.ModuleNameCap}} deleted successfully")
}