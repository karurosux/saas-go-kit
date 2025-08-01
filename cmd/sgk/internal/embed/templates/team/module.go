package team

import (
	"fmt"
	"os"
	
	"{{.Project.GoModule}}/internal/core"
	"{{.Project.GoModule}}/internal/team/controller"
	"{{.Project.GoModule}}/internal/team/middleware"
	"{{.Project.GoModule}}/internal/team/repository/gorm"
	"{{.Project.GoModule}}/internal/team/service"
	"github.com/labstack/echo/v4"
	gormdb "gorm.io/gorm"
)

// RegisterModule registers the team module with the container
func RegisterModule(c core.Container) error {
	// Get dependencies from container
	e, ok := c.Get("echo").(*echo.Echo)
	if !ok {
		return fmt.Errorf("echo instance not found in container")
	}
	
	db, ok := c.Get("db").(*gormdb.DB)
	if !ok {
		return fmt.Errorf("database instance not found in container")
	}
	
	// Run migrations
	if err := gorm.AutoMigrate(db); err != nil {
		return fmt.Errorf("failed to run team migrations: %w", err)
	}
	
	// Create repositories
	userRepo := gorm.NewUserRepository(db)
	memberRepo := gorm.NewTeamMemberRepository(db)
	tokenRepo := gorm.NewInvitationTokenRepository(db)
	
	// Create email sender
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	emailSender := teamservice.NewMockEmailSender()
	
	// Create notification service
	notificationSvc := teamservice.NewNotificationService(emailSender, baseURL)
	
	// Get max team size from environment
	maxTeamSize := 100
	if maxSizeStr := os.Getenv("MAX_TEAM_SIZE"); maxSizeStr != "" {
		if size, err := fmt.Sscanf(maxSizeStr, "%d", &maxTeamSize); err == nil && size > 0 {
			// Use the parsed size
		}
	}
	
	// Create team service
	teamService := teamservice.NewTeamService(
		userRepo,
		memberRepo,
		tokenRepo,
		notificationSvc,
		maxTeamSize,
	)
	
	// Create middleware
	teamMiddleware := teammiddleware.NewTeamMiddleware(teamService)
	
	// Create controller
	teamController := teamcontroller.NewTeamController(teamService)
	
	// Register routes
	teamController.RegisterRoutes(e, "/teams", teamMiddleware)
	
	// Register components in container for other modules to use
	c.Set("team.service", teamService)
	c.Set("team.middleware", teamMiddleware)
	c.Set("team.userRepository", userRepo)
	c.Set("team.memberRepository", memberRepo)
	c.Set("team.tokenRepository", tokenRepo)
	
	// Create initial owner for new accounts if configured
	c.OnAccountCreated(func(accountID, userID string) error {
		// This is a hook that can be called when a new account is created
		// to automatically add the creator as the owner
		return nil
	})
	
	return nil
}