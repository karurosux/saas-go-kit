package emailcontroller

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"{{.Project.GoModule}}/internal/core"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
)

type EmailController struct {
	emailService emailinterface.EmailService
	templates    emailinterface.TemplateManager
	validator    *core.Validator
}

func NewEmailController(emailService emailinterface.EmailService, templates emailinterface.TemplateManager) *EmailController {
	return &EmailController{
		emailService: emailService,
		templates:    templates,
		validator:    core.NewValidator(),
	}
}

func (c *EmailController) RegisterRoutes(e *echo.Echo, prefix string) {
	g := e.Group(prefix)
	
	g.POST("/send", c.SendEmail)
	g.POST("/send-template", c.SendTemplateEmail)
	g.POST("/queue", c.QueueEmail)
	g.GET("/status/:id", c.GetEmailStatus)
	
	g.GET("/templates", c.ListTemplates)
	g.GET("/templates/:name", c.GetTemplate)
	g.POST("/templates", c.CreateTemplate)
	g.PUT("/templates/:name", c.UpdateTemplate)
	g.DELETE("/templates/:name", c.DeleteTemplate)
}

type SendEmailRequest struct {
	To      []string `json:"to" validate:"required,min=1,dive,email"`
	CC      []string `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC     []string `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Subject string   `json:"subject" validate:"required"`
	Body    string   `json:"body" validate:"required"`
	HTML    string   `json:"html,omitempty"`
}

func (c *EmailController) SendEmail(ctx echo.Context) error {
	var req SendEmailRequest
	if err := ctx.Bind(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	if err := c.validator.Validate(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	var err error
	if req.HTML != "" {
		err = c.emailService.SendHTML(ctx.Request().Context(), req.To, req.Subject, req.Body, req.HTML)
	} else {
		err = c.emailService.Send(ctx.Request().Context(), req.To, req.Subject, req.Body)
	}
	
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, map[string]string{"status": "sent"}, "Email sent successfully")
}

type SendTemplateEmailRequest struct {
	To           []string               `json:"to" validate:"required,min=1,dive,email"`
	CC           []string               `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC          []string               `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Template     string                 `json:"template" validate:"required"`
	TemplateData map[string]interface{} `json:"template_data"`
}

func (c *EmailController) SendTemplateEmail(ctx echo.Context) error {
	var req SendTemplateEmailRequest
	if err := ctx.Bind(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	if err := c.validator.Validate(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	err := c.emailService.SendTemplate(ctx.Request().Context(), req.To, req.Template, req.TemplateData)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, map[string]string{"status": "sent"}, "Template email sent successfully")
}

type QueueEmailRequest struct {
	To           []string                     `json:"to" validate:"required,min=1,dive,email"`
	CC           []string                     `json:"cc,omitempty" validate:"omitempty,dive,email"`
	BCC          []string                     `json:"bcc,omitempty" validate:"omitempty,dive,email"`
	Subject      string                       `json:"subject"`
	Body         string                       `json:"body"`
	HTML         string                       `json:"html,omitempty"`
	Template     string                       `json:"template,omitempty"`
	TemplateData map[string]interface{}       `json:"template_data,omitempty"`
	Priority     emailinterface.EmailPriority `json:"priority,omitempty"`
}

func (c *EmailController) QueueEmail(ctx echo.Context) error {
	var req QueueEmailRequest
	if err := ctx.Bind(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	if err := c.validator.Validate(&req); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	message := &emailinterface.EmailMessage{
		To:           req.To,
		CC:           req.CC,
		BCC:          req.BCC,
		Subject:      req.Subject,
		Body:         req.Body,
		HTML:         req.HTML,
		Template:     req.Template,
		TemplateData: req.TemplateData,
		Priority:     req.Priority,
	}
	
	err := c.emailService.QueueEmail(ctx.Request().Context(), message)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, map[string]interface{}{"id": message.ID}, "Email queued successfully")
}

func (c *EmailController) GetEmailStatus(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return core.BadRequest(ctx, err)
	}
	
	message, err := c.emailService.GetEmailStatus(ctx.Request().Context(), uint(id))
	if err != nil {
		return core.NotFound(ctx, err)
	}
	
	return core.Success(ctx, message, "Email status retrieved")
}

func (c *EmailController) ListTemplates(ctx echo.Context) error {
	templates, err := c.templates.ListTemplates(ctx.Request().Context())
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, templates, "Templates retrieved")
}

func (c *EmailController) GetTemplate(ctx echo.Context) error {
	name := ctx.Param("name")
	
	template, err := c.templates.GetTemplate(ctx.Request().Context(), name)
	if err != nil {
		return core.NotFound(ctx, err)
	}
	
	return core.Success(ctx, template, "Template retrieved")
}

func (c *EmailController) CreateTemplate(ctx echo.Context) error {
	var template emailinterface.EmailTemplate
	if err := ctx.Bind(&template); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	if err := c.validator.Validate(&template); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	err := c.templates.CreateTemplate(ctx.Request().Context(), &template)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Created(ctx, template, "Template created successfully")
}

func (c *EmailController) UpdateTemplate(ctx echo.Context) error {
	name := ctx.Param("name")
	
	var template emailinterface.EmailTemplate
	if err := ctx.Bind(&template); err != nil {
		return core.BadRequest(ctx, err)
	}
	
	err := c.templates.UpdateTemplate(ctx.Request().Context(), name, &template)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, template, "Template updated successfully")
}

func (c *EmailController) DeleteTemplate(ctx echo.Context) error {
	name := ctx.Param("name")
	
	err := c.templates.DeleteTemplate(ctx.Request().Context(), name)
	if err != nil {
		return core.InternalServerError(ctx, err)
	}
	
	return core.Success(ctx, nil, "Template deleted successfully")
}