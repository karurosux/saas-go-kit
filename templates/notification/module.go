package notification

import (
	"fmt"
	"os"
	"strconv"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/notification/constants"
	"{{.Project.GoModule}}/internal/notification/controller"
	"{{.Project.GoModule}}/internal/notification/interface"
	"{{.Project.GoModule}}/internal/notification/provider/dev"
	"{{.Project.GoModule}}/internal/notification/provider/smtp"
	"{{.Project.GoModule}}/internal/notification/service"
	"github.com/labstack/echo/v4"
)

// RegisterModule registers the notification module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	// Initialize email provider based on configuration
	emailProvider, err := createEmailProvider()
	if err != nil {
		return fmt.Errorf("failed to create email provider: %w", err)
	}
	
	// For now, SMS and push providers are optional (nil)
	// In a full implementation, you would create them similarly to email provider
	var smsProvider notificationinterface.SMSProvider = nil
	var pushProvider notificationinterface.PushProvider = nil
	
	// Create notification service
	notificationService := notificationservice.NewNotificationService(
		emailProvider,
		smsProvider,
		pushProvider,
	)
	
	// Create common notification service if configuration is available
	var commonService notificationinterface.CommonNotificationService = nil
	if config := getCommonNotificationConfig(); config.AppName != "" {
		commonService = notificationservice.NewCommonNotificationService(
			notificationService,
			config,
		)
	}
	
	// Create controller
	notificationController := notificationcontroller.NewNotificationController(
		notificationService,
		commonService,
	)
	
	// Register routes
	notificationController.RegisterRoutes(e, "/notifications")
	
	// Register components in container for other modules to use
	c.Set("notification.service", notificationService)
	c.Set("notification.commonService", commonService)
	c.Set("notification.emailProvider", emailProvider)
	
	return nil
}

// createEmailProvider creates an email provider based on environment configuration
func createEmailProvider() (notificationinterface.EmailProvider, error) {
	provider := os.Getenv(notificationconstants.EnvEmailProvider)
	if provider == "" {
		provider = notificationconstants.ProviderSMTP
	}
	
	config := notificationinterface.EmailConfig{
		Provider:  provider,
		Host:      os.Getenv(notificationconstants.EnvSMTPHost),
		Port:      getIntEnv(notificationconstants.EnvSMTPPort, notificationconstants.DefaultSMTPPort),
		Username:  os.Getenv(notificationconstants.EnvSMTPUsername),
		Password:  os.Getenv(notificationconstants.EnvSMTPPassword),
		FromEmail: os.Getenv(notificationconstants.EnvFromEmail),
		FromName:  os.Getenv(notificationconstants.EnvFromName),
		UseTLS:    true, // Default to true for security
	}
	
	// Use development provider if in development mode or if SMTP host is not configured
	if os.Getenv("ENV") == "development" || config.Host == "" {
		devProvider := devprovider.NewDevEmailProvider()
		if err := devProvider.Initialize(config); err != nil {
			return nil, fmt.Errorf("failed to initialize dev email provider: %w", err)
		}
		return devProvider, nil
	}
	
	// Use SMTP provider for production
	switch provider {
	case notificationconstants.ProviderSMTP:
		smtpProvider := smtpprovider.NewSMTPProvider()
		if err := smtpProvider.Initialize(config); err != nil {
			return nil, fmt.Errorf("failed to initialize SMTP provider: %w", err)
		}
		return smtpProvider, nil
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", provider)
	}
}

// getCommonNotificationConfig creates common notification configuration from environment
func getCommonNotificationConfig() notificationinterface.CommonNotificationConfig {
	return notificationinterface.CommonNotificationConfig{
		AppName:      os.Getenv(notificationconstants.EnvAppName),
		AppURL:       os.Getenv(notificationconstants.EnvAppURL),
		FromEmail:    os.Getenv(notificationconstants.EnvFromEmail),
		FromName:     os.Getenv(notificationconstants.EnvFromName),
		SupportEmail: os.Getenv(notificationconstants.EnvSupportEmail),
	}
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}