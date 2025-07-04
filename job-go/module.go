package job

import (
	"context"
	"time"

	"github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Module struct {
	*core.BaseModule
	Service     JobService
	handlers    *handlers
	routePrefix string
	middleware  []echo.MiddlewareFunc
}

type ModuleConfig struct {
	DB                 *gorm.DB
	Service            JobService    // Optional: provide your own service
	RoutePrefix        string        // Default: "/api/jobs"
	Workers            int           // Default: 5
	PollInterval       time.Duration // Default: 5s
	MaxRetries         int           // Default: 3
	Middleware         []echo.MiddlewareFunc
	AutoStartWorkers   bool // Default: true
	RegisterJobHandlers bool // Default: true - whether to register HTTP endpoints
}

func NewModule(config ModuleConfig) (*Module, error) {
	if config.RoutePrefix == "" {
		config.RoutePrefix = "/api/jobs"
	}

	if config.Service == nil {
		panic("Service is required - use NewDefaultModule to create with default repositories")
	}

	module := &Module{
		BaseModule:  core.NewBaseModule("job"),
		Service:     config.Service,
		handlers:    NewHandlers(config.Service),
		routePrefix: config.RoutePrefix,
		middleware:  config.Middleware,
	}
	
	// Register routes
	module.registerRoutes()

	// Auto-start workers if configured
	if config.AutoStartWorkers {
		if err := config.Service.Start(context.Background()); err != nil {
			return nil, err
		}
	}

	return module, nil
}

func (m *Module) registerRoutes() {
	routes := []core.Route{
		{
			Method:      "POST",
			Path:        m.routePrefix,
			Handler:     m.handlers.CreateJob,
			Name:        "jobs.create",
			Description: "Create a new job",
			Middlewares: m.middleware,
		},
		{
			Method:      "GET",
			Path:        m.routePrefix + "/:id",
			Handler:     m.handlers.GetJob,
			Name:        "jobs.get",
			Description: "Get job details",
			Middlewares: m.middleware,
		},
		{
			Method:      "GET",
			Path:        m.routePrefix + "/:id/result",
			Handler:     m.handlers.GetJobResult,
			Name:        "jobs.result",
			Description: "Get job result",
			Middlewares: m.middleware,
		},
		{
			Method:      "GET",
			Path:        m.routePrefix,
			Handler:     m.handlers.ListJobs,
			Name:        "jobs.list",
			Description: "List jobs",
			Middlewares: m.middleware,
		},
		{
			Method:      "POST",
			Path:        m.routePrefix + "/:id/cancel",
			Handler:     m.handlers.CancelJob,
			Name:        "jobs.cancel",
			Description: "Cancel a job",
			Middlewares: m.middleware,
		},
		{
			Method:      "POST",
			Path:        m.routePrefix + "/:id/retry",
			Handler:     m.handlers.RetryJob,
			Name:        "jobs.retry",
			Description: "Retry a job",
			Middlewares: m.middleware,
		},
		{
			Method:      "DELETE",
			Path:        m.routePrefix + "/:id",
			Handler:     m.handlers.DeleteJob,
			Name:        "jobs.delete",
			Description: "Delete a job",
			Middlewares: m.middleware,
		},
	}
	
	m.AddRoutes(routes)
}

func (m *Module) Stop(ctx context.Context) error {
	return m.Service.Stop(ctx)
}


// Ensure Module implements core.Module interface
var _ core.Module = (*Module)(nil)