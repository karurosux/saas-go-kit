package email

import (
	"context"
	"embed"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/samber/do"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"{{.Project.GoModule}}/internal/core"
	emailcontroller "{{.Project.GoModule}}/internal/email/controller"
	emailinterface "{{.Project.GoModule}}/internal/email/interface"
	emailgorm "{{.Project.GoModule}}/internal/email/repository/gorm"
	emailservice "{{.Project.GoModule}}/internal/email/service"
)

var templateFS embed.FS

func ProvideEmailQueue(i *do.Injector) (emailinterface.EmailQueue, error) {
	db := do.MustInvoke[*gorm.DB](i)
	
	if err := emailgorm.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run email migrations: %w", err)
	}
	
	return emailgorm.NewEmailQueueRepository(db), nil
}

func ProvideTemplateRepository(i *do.Injector) (emailinterface.TemplateRepository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return emailgorm.NewTemplateRepository(db), nil
}

func ProvideTemplateManager(i *do.Injector) (emailinterface.TemplateManager, error) {
	templateRepo := do.MustInvoke[emailinterface.TemplateRepository](i)
	return emailservice.NewTemplateManager(templateFS, "templates", templateRepo), nil
}

func ProvideEmailSender(i *do.Injector) (emailinterface.EmailSender, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return emailservice.NewMockSender(true), nil
	}
	
	smtpPort, err := strconv.Atoi(getEnvWithDefault("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}
	
	useTLS, _ := strconv.ParseBool(getEnvWithDefault("SMTP_TLS", "true"))
	skipVerify, _ := strconv.ParseBool(getEnvWithDefault("SMTP_SKIP_VERIFY", "false"))
	
	config := emailservice.SMTPConfig{
		Host:       smtpHost,
		Port:       smtpPort,
		Username:   os.Getenv("SMTP_USERNAME"),
		Password:   os.Getenv("SMTP_PASSWORD"),
		From:       getEnvWithDefault("SMTP_FROM", "noreply@example.com"),
		FromName:   getEnvWithDefault("SMTP_FROM_NAME", ""),
		TLS:        useTLS,
		SkipVerify: skipVerify,
	}
	
	return emailservice.NewSMTPSender(config), nil
}

func ProvideEmailService(i *do.Injector) (emailinterface.EmailService, error) {
	sender := do.MustInvoke[emailinterface.EmailSender](i)
	queue := do.MustInvoke[emailinterface.EmailQueue](i)
	templates := do.MustInvoke[emailinterface.TemplateManager](i)
	
	defaultFrom := getEnvWithDefault("EMAIL_FROM", "noreply@example.com")
	
	service := emailservice.NewEmailService(sender, queue, templates, defaultFrom)
	
	if getEnvWithDefault("EMAIL_QUEUE_ENABLED", "true") == "true" {
		intervalStr := getEnvWithDefault("EMAIL_QUEUE_INTERVAL", "30s")
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			interval = 30 * time.Second
		}
		
		go service.StartQueueProcessor(context.Background(), interval)
	}
	
	return service, nil
}

func ProvideEmailController(i *do.Injector) (*emailcontroller.EmailController, error) {
	emailService := do.MustInvoke[emailinterface.EmailService](i)
	templates := do.MustInvoke[emailinterface.TemplateManager](i)
	return emailcontroller.NewEmailController(emailService, templates), nil
}

func RegisterModule(container *core.Container) error {
	do.Provide(container, ProvideEmailQueue)
	do.Provide(container, ProvideTemplateRepository)
	do.Provide(container, ProvideTemplateManager)
	do.Provide(container, ProvideEmailSender)
	do.Provide(container, ProvideEmailService)
	do.Provide(container, ProvideEmailController)
	
	e := do.MustInvoke[*echo.Echo](container)
	emailController := do.MustInvoke[*emailcontroller.EmailController](container)
	
	emailController.RegisterRoutes(e, "/api/v1/email")
	
	return nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}