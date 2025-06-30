package core

import (
	"fmt"
	"log"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Kit manages modules and their lifecycle
type Kit struct {
	echo       *echo.Echo
	modules    map[string]Module
	registered []string
	mounted    bool
	config     KitConfig
}

// KitConfig holds configuration for the kit
type KitConfig struct {
	// Debug enables debug mode
	Debug bool
	
	// RoutePrefix adds a prefix to all routes
	RoutePrefix string
	
	// DisableStartupBanner disables the startup banner
	DisableStartupBanner bool
	
	// ErrorHandler is the global error handler
	ErrorHandler echo.HTTPErrorHandler
}

// NewKit creates a new kit instance
func NewKit(e *echo.Echo, config ...KitConfig) *Kit {
	cfg := KitConfig{}
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.ErrorHandler != nil {
		e.HTTPErrorHandler = cfg.ErrorHandler
	}

	return &Kit{
		echo:    e,
		modules: make(map[string]Module),
		config:  cfg,
	}
}

// Register registers a module
func (k *Kit) Register(module Module) error {
	if k.mounted {
		return fmt.Errorf("cannot register module after mounting")
	}

	name := module.Name()
	if _, exists := k.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	// Check dependencies
	for _, dep := range module.Dependencies() {
		if _, exists := k.modules[dep]; !exists {
			return fmt.Errorf("module %s depends on %s which is not registered", name, dep)
		}
	}

	k.modules[name] = module
	k.registered = append(k.registered, name)

	if k.config.Debug {
		log.Printf("[SaaS Kit] Registered module: %s", name)
	}

	return nil
}

// Get returns a registered module
func (k *Kit) Get(name string) Module {
	return k.modules[name]
}

// Mount mounts all registered modules
func (k *Kit) Mount() error {
	if k.mounted {
		return fmt.Errorf("kit already mounted")
	}

	// Sort modules by dependencies
	sorted, err := k.sortModules()
	if err != nil {
		return err
	}

	// Initialize modules in order
	for _, name := range sorted {
		module := k.modules[name]
		
		// Collect dependencies
		deps := make(map[string]Module)
		for _, depName := range module.Dependencies() {
			deps[depName] = k.modules[depName]
		}

		// Initialize module
		if err := module.Init(deps); err != nil {
			return fmt.Errorf("failed to initialize module %s: %w", name, err)
		}

		// Apply module middleware
		for _, mw := range module.Middleware() {
			k.echo.Use(mw)
		}

		// Register routes
		for _, route := range module.Routes() {
			path := k.config.RoutePrefix + route.Path
			
			// Build middleware chain
			handlers := append([]echo.MiddlewareFunc{}, route.Middlewares...)
			
			// Register route
			switch route.Method {
			case "GET":
				k.echo.GET(path, route.Handler, handlers...)
			case "POST":
				k.echo.POST(path, route.Handler, handlers...)
			case "PUT":
				k.echo.PUT(path, route.Handler, handlers...)
			case "DELETE":
				k.echo.DELETE(path, route.Handler, handlers...)
			case "PATCH":
				k.echo.PATCH(path, route.Handler, handlers...)
			case "HEAD":
				k.echo.HEAD(path, route.Handler, handlers...)
			case "OPTIONS":
				k.echo.OPTIONS(path, route.Handler, handlers...)
			default:
				return fmt.Errorf("unsupported method %s for route %s", route.Method, path)
			}

			if k.config.Debug {
				log.Printf("[SaaS Kit] Mounted route: %s %s (%s)", route.Method, path, route.Name)
			}
		}

		if k.config.Debug {
			log.Printf("[SaaS Kit] Mounted module: %s", name)
		}
	}

	k.mounted = true

	if !k.config.DisableStartupBanner {
		k.printBanner()
	}

	return nil
}

// sortModules performs topological sort on modules based on dependencies
func (k *Kit) sortModules() ([]string, error) {
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	result := []string{}

	var visit func(string) error
	visit = func(name string) error {
		if temp[name] {
			return fmt.Errorf("circular dependency detected involving module %s", name)
		}
		if visited[name] {
			return nil
		}

		temp[name] = true
		module := k.modules[name]
		
		for _, dep := range module.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}

		temp[name] = false
		visited[name] = true
		result = append(result, name)
		
		return nil
	}

	// Visit all modules
	for name := range k.modules {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// printBanner prints the startup banner
func (k *Kit) printBanner() {
	fmt.Println(`
╔═══════════════════════════════════════╗
║         SaaS Go Kit v1.0.0            ║
╠═══════════════════════════════════════╣`)
	
	// Sort module names for consistent output
	names := make([]string, 0, len(k.modules))
	for name := range k.modules {
		names = append(names, name)
	}
	sort.Strings(names)
	
	fmt.Println("║ Loaded Modules:                       ║")
	for _, name := range names {
		fmt.Printf("║   • %-33s ║\n", name)
	}
	
	fmt.Println("╚═══════════════════════════════════════╝")
}

// Builder provides a fluent interface for building a kit
type Builder struct {
	echo    *echo.Echo
	config  KitConfig
	modules []Module
}

// NewBuilder creates a new kit builder
func NewBuilder() *Builder {
	e := echo.New()
	e.HideBanner = true
	
	// Default middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	
	return &Builder{
		echo:   e,
		config: KitConfig{},
	}
}

// WithDebug enables debug mode
func (b *Builder) WithDebug(debug bool) *Builder {
	b.config.Debug = debug
	b.echo.Debug = debug
	return b
}

// WithRoutePrefix sets the route prefix
func (b *Builder) WithRoutePrefix(prefix string) *Builder {
	b.config.RoutePrefix = prefix
	return b
}

// WithModule adds a module
func (b *Builder) WithModule(module Module) *Builder {
	b.modules = append(b.modules, module)
	return b
}

// WithModules adds multiple modules
func (b *Builder) WithModules(modules ...Module) *Builder {
	b.modules = append(b.modules, modules...)
	return b
}

// WithMiddleware adds global middleware
func (b *Builder) WithMiddleware(middleware ...echo.MiddlewareFunc) *Builder {
	b.echo.Use(middleware...)
	return b
}

// WithErrorHandler sets the error handler
func (b *Builder) WithErrorHandler(handler echo.HTTPErrorHandler) *Builder {
	b.config.ErrorHandler = handler
	return b
}

// Build builds the kit and returns the echo instance
func (b *Builder) Build() (*echo.Echo, error) {
	kit := NewKit(b.echo, b.config)
	
	// Register all modules
	for _, module := range b.modules {
		if err := kit.Register(module); err != nil {
			return nil, err
		}
	}
	
	// Mount modules
	if err := kit.Mount(); err != nil {
		return nil, err
	}
	
	return b.echo, nil
}