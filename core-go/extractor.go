package core

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// ExtractorMode enables route extraction mode
var ExtractorMode = false

// ExtractorOutput is the output path for extracted routes
var ExtractorOutput = ""

// ExtractedModule represents module metadata for extraction
type ExtractedModule struct {
	Name         string           `json:"name"`
	Routes       []ExtractedRoute `json:"routes"`
	Types        []ExtractedType  `json:"types,omitempty"`
	Dependencies []string         `json:"dependencies,omitempty"`
}

// ExtractedRoute represents route metadata for extraction
type ExtractedRoute struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PathParams  []string `json:"path_params,omitempty"`
	HandlerInfo string   `json:"handler_info,omitempty"`
}

// ExtractedType represents type metadata for extraction
type ExtractedType struct {
	Name        string              `json:"name"`
	Package     string              `json:"package,omitempty"`
	Fields      []ExtractedField    `json:"fields"`
	Methods     []string            `json:"methods,omitempty"`
	Description string              `json:"description,omitempty"`
}

// ExtractedField represents field metadata for extraction
type ExtractedField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	JSONName string `json:"json_name,omitempty"`
	Required bool   `json:"required"`
}

// extractedModules holds all extracted modules
var extractedModules []ExtractedModule

// InitExtractor initializes the extractor with the given output path
func InitExtractor(outputPath string) {
	ExtractorMode = true
	ExtractorOutput = outputPath
	fmt.Printf("[SaaS Kit] Route extraction mode enabled. Output: %s\n", ExtractorOutput)
}

// EnableExtractor enables the route extraction mode
func EnableExtractor(outputPath string) {
	ExtractorMode = true
	ExtractorOutput = outputPath
}

// captureModule captures module information during registration
func (k *Kit) captureModule(module Module) {
	if !ExtractorMode {
		return
	}

	extractedModule := ExtractedModule{
		Name:         module.Name(),
		Routes:       []ExtractedRoute{},
		Types:        []ExtractedType{},
		Dependencies: module.Dependencies(),
	}

	// Extract routes
	for _, route := range module.Routes() {
		extractedRoute := ExtractedRoute{
			Method:      route.Method,
			Path:        route.Path,
			Name:        route.Name,
			Description: route.Description,
			PathParams:  extractPathParams(route.Path),
		}

		// Extract handler information
		if route.Handler != nil {
			handlerType := reflect.TypeOf(route.Handler)
			extractedRoute.HandlerInfo = fmt.Sprintf("%v", handlerType)
		}

		extractedModule.Routes = append(extractedModule.Routes, extractedRoute)
	}

	// Extract types using reflection
	extractedModule.Types = extractModuleTypes(module)

	extractedModules = append(extractedModules, extractedModule)
}

// extractPathParams extracts path parameters from a route path
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

// extractModuleTypes uses reflection to extract type information
func extractModuleTypes(module Module) []ExtractedType {
	var types []ExtractedType
	
	// Get module type
	moduleType := reflect.TypeOf(module)
	if moduleType.Kind() == reflect.Ptr {
		moduleType = moduleType.Elem()
	}
	
	// Extract struct information
	if moduleType.Kind() == reflect.Struct {
		moduleTypeMeta := ExtractedType{
			Name:    moduleType.Name(),
			Package: moduleType.PkgPath(),
			Fields:  extractFields(moduleType),
			Methods: extractMethods(reflect.TypeOf(module)),
		}
		types = append(types, moduleTypeMeta)
		
		// Look for nested types in fields
		for i := 0; i < moduleType.NumField(); i++ {
			field := moduleType.Field(i)
			fieldType := field.Type
			
			// Dereference pointers
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}
			
			// Extract interfaces and structs
			if fieldType.Kind() == reflect.Interface || fieldType.Kind() == reflect.Struct {
				// Skip well-known types
				if !isWellKnownType(fieldType) {
					nestedType := ExtractedType{
						Name:    fieldType.Name(),
						Package: fieldType.PkgPath(),
						Fields:  extractFields(fieldType),
					}
					types = append(types, nestedType)
				}
			}
		}
	}
	
	return types
}

// extractFields extracts field information from a struct type
func extractFields(t reflect.Type) []ExtractedField {
	var fields []ExtractedField
	
	if t.Kind() != reflect.Struct {
		return fields
	}
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		extractedField := ExtractedField{
			Name:     field.Name,
			Type:     typeToString(field.Type),
			Required: !hasOmitemptyTag(field.Tag),
		}
		
		// Extract JSON tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "-" {
				extractedField.JSONName = parts[0]
			}
			
			// Check for omitempty
			for _, part := range parts[1:] {
				if part == "omitempty" {
					extractedField.Required = false
				}
			}
		}
		
		fields = append(fields, extractedField)
	}
	
	return fields
}

// extractMethods extracts method names from a type
func extractMethods(t reflect.Type) []string {
	var methods []string
	
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.IsExported() {
			methods = append(methods, method.Name)
		}
	}
	
	return methods
}

// typeToString converts a reflect.Type to a string representation
func typeToString(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Ptr:
		return typeToString(t.Elem()) + " | null"
	case reflect.Slice, reflect.Array:
		return typeToString(t.Elem()) + "[]"
	case reflect.Map:
		return fmt.Sprintf("Record<%s, %s>", typeToString(t.Key()), typeToString(t.Elem()))
	case reflect.Interface:
		if t.String() == "interface {}" {
			return "any"
		}
		return t.Name()
	case reflect.Struct:
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "string"
		}
		if t.Name() != "" {
			return t.Name()
		}
		return "object"
	default:
		if t.Name() != "" {
			return t.Name()
		}
		return "any"
	}
}

// hasOmitemptyTag checks if a struct tag contains omitempty
func hasOmitemptyTag(tag reflect.StructTag) bool {
	jsonTag := tag.Get("json")
	return strings.Contains(jsonTag, "omitempty")
}

// isWellKnownType checks if a type is a well-known type that should be skipped
func isWellKnownType(t reflect.Type) bool {
	pkg := t.PkgPath()
	return strings.HasPrefix(pkg, "github.com/labstack/echo") ||
		strings.HasPrefix(pkg, "gorm.io") ||
		strings.HasPrefix(pkg, "database/sql") ||
		strings.HasPrefix(pkg, "time") ||
		strings.HasPrefix(pkg, "sync") ||
		strings.HasPrefix(pkg, "context")
}

// saveExtractedRoutes saves the extracted routes to a file
func saveExtractedRoutes() error {
	if !ExtractorMode || ExtractorOutput == "" {
		return nil
	}

	// Create output directory if needed
	dir := strings.TrimSuffix(ExtractorOutput, "/routes.json")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(extractedModules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal extracted routes: %w", err)
	}

	// Write to file
	if err := os.WriteFile(ExtractorOutput, data, 0644); err != nil {
		return fmt.Errorf("failed to write extracted routes: %w", err)
	}

	fmt.Printf("Routes extracted to %s\n", ExtractorOutput)
	
	// Exit after extraction
	os.Exit(0)
	
	return nil
}

// ExitIfExtracting checks if we're in extraction mode and exits after saving
func ExitIfExtracting() {
	if ExtractorMode {
		if err := saveExtractedRoutes(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save extracted routes: %v\n", err)
			os.Exit(1)
		}
	}
}