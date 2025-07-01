package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	core "github.com/karurosux/saas-go-kit/core-go"
	"github.com/labstack/echo/v4"
)

type TypeField struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	JSONName string      `json:"json_name,omitempty"`
	Required bool        `json:"required"`
	Example  interface{} `json:"example,omitempty"`
}

type TypeInfo struct {
	Name        string      `json:"name"`
	Fields      []TypeField `json:"fields"`
	Description string      `json:"description,omitempty"`
}

type RouteMetadata struct {
	Method        string      `json:"method"`
	Path          string      `json:"path"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	RequestType   *TypeInfo   `json:"request_type,omitempty"`
	ResponseType  *TypeInfo   `json:"response_type,omitempty"`
	QueryParams   []TypeField `json:"query_params,omitempty"`
	PathParams    []string    `json:"path_params,omitempty"`
}

type ModuleMetadata struct {
	Name   string          `json:"name"`
	Routes []RouteMetadata `json:"routes"`
	Types  []TypeInfo      `json:"types,omitempty"`
}

// ModuleProvider interface that projects must implement
type ModuleProvider interface {
	GetModules() map[string]core.Module
}

func main() {
	var outputPath = flag.String("output", "./generated/routes.json", "Output path for route metadata")
	var configPath = flag.String("config", "", "Path to module configuration file (optional)")
	flag.Parse()

	// Try to find and load modules dynamically
	modules := make(map[string]core.Module)
	
	// If config file is provided, use it
	if *configPath != "" {
		loadedModules, err := loadModulesFromConfig(*configPath)
		if err != nil {
			log.Printf("Warning: Failed to load modules from config: %v", err)
		} else {
			modules = loadedModules
		}
	}
	
	// If no modules loaded, try to auto-discover
	if len(modules) == 0 {
		log.Println("No config provided, attempting auto-discovery...")
		discoveredModules, err := autoDiscoverModules()
		if err != nil {
			log.Printf("Warning: Auto-discovery failed: %v", err)
		} else {
			modules = discoveredModules
		}
	}
	
	// If still no modules, show usage
	if len(modules) == 0 {
		log.Fatal(`
No modules found. Please use one of these approaches:

1. Create a module configuration file and use -config flag
2. Make sure your project has a cmd/route_extractor with module setup
3. For custom setup, copy this tool to your project and modify it

For projects with custom modules, it's recommended to create your own route extractor:
cmd/extract-routes/main.go in your project with your specific module setup.
`)
	}

	// Extract routes
	if err := ExtractRoutes(modules, *outputPath); err != nil {
		log.Fatal("Failed to extract routes:", err)
	}

	fmt.Printf("Route metadata exported to %s\n", *outputPath)
}

func loadModulesFromConfig(configPath string) (map[string]core.Module, error) {
	// TODO: Implement config file loading
	return nil, fmt.Errorf("config file loading not implemented yet")
}

func autoDiscoverModules() (map[string]core.Module, error) {
	// Auto-discover modules by scanning go.mod and source code
	return discoverModulesFromProject()
}

func discoverModulesFromProject() (map[string]core.Module, error) {
	modules := make(map[string]core.Module)
	
	// Scan for main.go to understand the project structure
	mainFile, err := findMainFile()
	if err != nil {
		return nil, fmt.Errorf("could not find main.go: %v", err)
	}
	
	// Parse the main.go file to extract module instantiations
	discoveredModules, err := parseMainFileForModules(mainFile)
	if err != nil {
		log.Printf("Warning: Could not parse main.go for modules: %v", err)
	} else {
		for name, module := range discoveredModules {
			modules[name] = module
		}
	}
	
	log.Printf("Discovered %d modules from project", len(modules))
	return modules, nil
}

func findMainFile() (string, error) {
	// Look for main.go in common locations
	possiblePaths := []string{
		"./main.go",
		"./cmd/server/main.go", 
		"./cmd/app/main.go",
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("no main.go found in common locations")
}

func parseMainFileForModules(mainPath string) (map[string]core.Module, error) {
	log.Printf("Parsing %s for module instantiations...", mainPath)
	
	// Parse the main.go file
	fset := token.NewFileSet()
	src, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read main.go: %v", err)
	}
	
	node, err := parser.ParseFile(fset, mainPath, src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main.go: %v", err)
	}
	
	// Look for module instantiation patterns
	modules := make(map[string]core.Module)
	
	// Find imports to understand available modules
	imports := extractImports(node)
	log.Printf("Found imports: %v", imports)
	
	// Find module instantiations in the AST
	ast.Inspect(node, func(n ast.Node) bool {
		if assignStmt, ok := n.(*ast.AssignStmt); ok {
			if len(assignStmt.Lhs) == 1 && len(assignStmt.Rhs) == 1 {
				if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
					if callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr); ok {
						log.Printf("Analyzing assignment: %s := %s", ident.Name, "call expression")
						moduleName, moduleType := analyzeModuleInstantiation(ident.Name, callExpr, imports)
						if moduleName != "" {
							log.Printf("Found module: %s (type: %s)", moduleName, moduleType)
							
							// Create dummy module instance
							if dummyModule := createDummyModule(moduleType, moduleName); dummyModule != nil {
								modules[moduleName] = dummyModule
							}
						}
					}
				}
			}
		}
		return true
	})
	
	log.Printf("Successfully parsed %d modules", len(modules))
	return modules, nil
}

func extractImports(node *ast.File) map[string]string {
	imports := make(map[string]string)
	
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		
		// Handle aliased imports
		if imp.Name != nil {
			imports[imp.Name.Name] = path
		} else {
			// Extract package name from path
			parts := strings.Split(path, "/")
			pkg := parts[len(parts)-1]
			imports[pkg] = path
			
			// Also add mapping without -go suffix for saas-go-kit modules
			if strings.Contains(path, "saas-go-kit") && strings.HasSuffix(pkg, "-go") {
				shortName := strings.TrimSuffix(pkg, "-go")
				imports[shortName] = path
			}
		}
	}
	
	return imports
}

func analyzeModuleInstantiation(varName string, callExpr *ast.CallExpr, imports map[string]string) (string, string) {
	// Look for module creation patterns like:
	// authModule := auth.NewModule(...)
	// healthModule := health.NewModule(...)
	
	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selExpr.X.(*ast.Ident); ok {
			pkgName := ident.Name
			funcName := selExpr.Sel.Name
			
			log.Printf("Found call: %s.%s", pkgName, funcName)
			
			if funcName == "NewModule" {
				// Map package names to module types
				if importPath, exists := imports[pkgName]; exists {
					log.Printf("Found NewModule call: %s -> %s", pkgName, importPath)
					moduleName := extractModuleNameFromVar(varName)
					log.Printf("Detected module: %s (path: %s)", moduleName, importPath)
					return moduleName, importPath
				} else {
					log.Printf("Package %s not found in imports", pkgName)
				}
			}
		}
	}
	
	return "", ""
}

func extractModuleNameFromVar(varName string) string {
	// Convert "authModule" -> "auth", "healthModule" -> "health"
	if strings.HasSuffix(varName, "Module") {
		return strings.TrimSuffix(varName, "Module")
	}
	return varName
}

func createDummyModule(moduleType, moduleName string) core.Module {
	// Create dummy modules based on the import path
	switch {
	case strings.Contains(moduleType, "auth-go"):
		return createDummyAuthModule()
	case strings.Contains(moduleType, "health-go"):
		return createDummyHealthModule()
	case strings.Contains(moduleType, "role-go"):
		return createDummyRoleModule()
	default:
		// Create a generic dummy module for custom modules
		log.Printf("Creating generic dummy module for: %s", moduleName)
		return createGenericDummyModule(moduleName)
	}
}

func createDummyAuthModule() core.Module {
	return &DummyModule{
		name: "auth",
		routes: []core.Route{
			{Method: "POST", Path: "/register", Name: "auth.register", Description: "Register new user"},
			{Method: "POST", Path: "/login", Name: "auth.login", Description: "User login"},
			{Method: "POST", Path: "/logout", Name: "auth.logout", Description: "User logout"},
			{Method: "GET", Path: "/profile", Name: "auth.profile", Description: "Get user profile"},
		},
	}
}

func createDummyHealthModule() core.Module {
	return &DummyModule{
		name: "health",
		routes: []core.Route{
			{Method: "GET", Path: "/health", Name: "health.check", Description: "Health check"},
			{Method: "GET", Path: "/health/detailed", Name: "health.detailed", Description: "Detailed health check"},
		},
	}
}

func createDummyRoleModule() core.Module {
	return &DummyModule{
		name: "role",
		routes: []core.Route{
			{Method: "GET", Path: "/roles", Name: "role.list", Description: "List roles"},
			{Method: "POST", Path: "/roles", Name: "role.create", Description: "Create role"},
			{Method: "GET", Path: "/roles/:id", Name: "role.get", Description: "Get role"},
		},
	}
}

func createGenericDummyModule(moduleName string) core.Module {
	// Create a generic dummy module with common CRUD routes
	return &DummyModule{
		name: moduleName,
		routes: []core.Route{
			{Method: "GET", Path: fmt.Sprintf("/%s", moduleName), Name: fmt.Sprintf("%s.list", moduleName), Description: fmt.Sprintf("List %s", moduleName)},
			{Method: "POST", Path: fmt.Sprintf("/%s", moduleName), Name: fmt.Sprintf("%s.create", moduleName), Description: fmt.Sprintf("Create %s", moduleName)},
			{Method: "GET", Path: fmt.Sprintf("/%s/:id", moduleName), Name: fmt.Sprintf("%s.get", moduleName), Description: fmt.Sprintf("Get %s", moduleName)},
			{Method: "PUT", Path: fmt.Sprintf("/%s/:id", moduleName), Name: fmt.Sprintf("%s.update", moduleName), Description: fmt.Sprintf("Update %s", moduleName)},
			{Method: "DELETE", Path: fmt.Sprintf("/%s/:id", moduleName), Name: fmt.Sprintf("%s.delete", moduleName), Description: fmt.Sprintf("Delete %s", moduleName)},
		},
	}
}

// DummyModule implements core.Module interface for route extraction
type DummyModule struct {
	name   string
	routes []core.Route
}

func (d *DummyModule) Name() string {
	return d.name
}

func (d *DummyModule) Routes() []core.Route {
	return d.routes
}

func (d *DummyModule) Middleware() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{}
}

func (d *DummyModule) Dependencies() []string {
	return []string{}
}

func (d *DummyModule) Init(deps map[string]core.Module) error {
	return nil
}

// ExtractRoutes extracts route metadata from modules and saves to file
func ExtractRoutes(modules map[string]core.Module, outputPath string) error {
	var moduleMetadata []ModuleMetadata

	for name, module := range modules {
		metadata := extractModuleMetadata(name, module)
		moduleMetadata = append(moduleMetadata, metadata)
	}

	// Marshal to JSON
	output, err := json.MarshalIndent(moduleMetadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Create output directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Route metadata exported to %s\n", outputPath)
	return nil
}

func extractModuleMetadata(name string, module core.Module) ModuleMetadata {
	metadata := ModuleMetadata{
		Name:   name,
		Routes: []RouteMetadata{},
		Types:  []TypeInfo{},
	}

	// Extract types from module path if we can find it
	typeRegistry := make(map[string]*TypeInfo)
	if modulePath := getModulePath(name); modulePath != "" {
		extractTypesFromSource(modulePath, typeRegistry)
	}

	// Convert type registry to slice
	for _, typeInfo := range typeRegistry {
		metadata.Types = append(metadata.Types, *typeInfo)
	}

	// Extract routes
	routes := module.Routes()
	for _, route := range routes {
		routeMetadata := RouteMetadata{
			Method:      route.Method,
			Path:        route.Path,
			Name:        route.Name,
			Description: route.Description,
			PathParams:  extractPathParams(route.Path),
		}

		// Try to extract request/response types
		if route.Handler != nil {
			analyzeHandlerForTypes(route, &routeMetadata, typeRegistry)
		}

		metadata.Routes = append(metadata.Routes, routeMetadata)
	}

	return metadata
}

func getModulePath(moduleName string) string {
	// Common paths where modules might be found
	possiblePaths := []string{
		fmt.Sprintf("./vendor/github.com/karurosux/saas-go-kit/%s-go", moduleName),
		fmt.Sprintf("./vendor/github.com/karurosux/saas-go-kit/%s", moduleName),
		fmt.Sprintf("./internal/%s", moduleName),
		fmt.Sprintf("./pkg/%s", moduleName),
		fmt.Sprintf("./%s", moduleName),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

func analyzeHandlerForTypes(route core.Route, routeMeta *RouteMetadata, typeRegistry map[string]*TypeInfo) {
	// Simple heuristic based on route name and method
	if route.Name != "" {
		// Extract potential type names from route name
		parts := strings.Split(route.Name, ".")
		if len(parts) > 1 {
			actionName := parts[len(parts)-1]
			
			// Look for matching request types
			for typeName, typeInfo := range typeRegistry {
				if strings.Contains(strings.ToLower(typeName), "request") &&
					strings.Contains(strings.ToLower(typeName), strings.ToLower(actionName)) {
					routeMeta.RequestType = typeInfo
				}
				if strings.Contains(strings.ToLower(typeName), "response") &&
					strings.Contains(strings.ToLower(typeName), strings.ToLower(actionName)) {
					routeMeta.ResponseType = typeInfo
				}
			}
		}
	}
}

func extractTypesFromSource(modulePath string, typeRegistry map[string]*TypeInfo) {
	fset := token.NewFileSet()

	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		node, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		// Look for struct types
		ast.Inspect(node, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if structType, ok := x.Type.(*ast.StructType); ok {
					typeInfo := extractStructInfo(x.Name.Name, structType)
					typeRegistry[typeInfo.Name] = typeInfo
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		log.Printf("Warning: Failed to extract types from %s: %v", modulePath, err)
	}
}

func extractStructInfo(name string, structType *ast.StructType) *TypeInfo {
	typeInfo := &TypeInfo{
		Name:   name,
		Fields: []TypeField{},
	}

	if structType.Fields != nil {
		for _, field := range structType.Fields.List {
			// Handle fields with names
			for _, fieldName := range field.Names {
				// Skip private fields
				if !ast.IsExported(fieldName.Name) {
					continue
				}

				fieldInfo := TypeField{
					Name:     fieldName.Name,
					Type:     typeToString(field.Type),
					Required: !hasOmitemptyTag(field.Tag),
				}

				// Extract JSON tag
				if field.Tag != nil {
					tagValue := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
					if jsonTag := tagValue.Get("json"); jsonTag != "" {
						parts := strings.Split(jsonTag, ",")
						if parts[0] != "-" {
							fieldInfo.JSONName = parts[0]
						} else {
							fieldInfo.JSONName = "-"
						}
						
						// Check for omitempty
						for _, part := range parts[1:] {
							if part == "omitempty" {
								fieldInfo.Required = false
							}
						}
					}
				}

				typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
			}

			// Handle embedded fields
			if len(field.Names) == 0 {
				if ident, ok := field.Type.(*ast.Ident); ok {
					// Embedded type
					fieldInfo := TypeField{
						Name:     ident.Name,
						Type:     typeToString(field.Type),
						Required: true,
					}
					typeInfo.Fields = append(typeInfo.Fields, fieldInfo)
				}
			}
		}
	}

	return typeInfo
}

func typeToString(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return convertGoTypeToTS(x.Name)
	case *ast.SelectorExpr:
		// Handle types like time.Time
		pkg := typeToString(x.X)
		if pkg == "time" && x.Sel.Name == "Time" {
			return "string"
		}
		return x.Sel.Name
	case *ast.StarExpr:
		return typeToString(x.X) + " | null"
	case *ast.ArrayType:
		return typeToString(x.Elt) + "[]"
	case *ast.MapType:
		keyType := typeToString(x.Key)
		valueType := typeToString(x.Value)
		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	case *ast.InterfaceType:
		return "any"
	default:
		return "any"
	}
}

func convertGoTypeToTS(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "time.Time", "Time":
		return "string"
	case "interface{}":
		return "any"
	default:
		return goType
	}
}

func hasOmitemptyTag(tag *ast.BasicLit) bool {
	if tag == nil {
		return false
	}
	return strings.Contains(tag.Value, "omitempty")
}

func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}
	
	return params
}