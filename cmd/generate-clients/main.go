package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	Method        string     `json:"method"`
	Path          string     `json:"path"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	RequestType   *TypeInfo  `json:"request_type,omitempty"`
	ResponseType  *TypeInfo  `json:"response_type,omitempty"`
	QueryParams   []TypeField `json:"query_params,omitempty"`
	PathParams    []string   `json:"path_params,omitempty"`
}

type ModuleMetadata struct {
	Name   string          `json:"name"`
	Routes []RouteMetadata `json:"routes"`
	Types  []TypeInfo      `json:"types,omitempty"`
}

func main() {
	var (
		routesPath = flag.String("routes", "./generated/routes.json", "Path to routes metadata")
		outputDir  = flag.String("output", "./generated/clients", "Output directory")
		outputDirShort = flag.String("o", "./generated/clients", "Output directory (shorthand)")
		format     = flag.String("format", "axios", "Client format")
	)
	flag.Parse()
	
	// Use -o flag if provided, otherwise use --output
	finalOutputDir := *outputDir
	if *outputDirShort != "./generated/clients" {
		finalOutputDir = *outputDirShort
	}

	// Read routes metadata
	routesData, err := os.ReadFile(*routesPath)
	if err != nil {
		log.Fatal("Failed to read routes metadata:", err)
	}

	var modules []ModuleMetadata
	if err := json.Unmarshal(routesData, &modules); err != nil {
		log.Fatal("Failed to parse routes metadata:", err)
	}

	// Generate clients
	if err := generateClients(finalOutputDir, modules, *format); err != nil {
		log.Fatal("Failed to generate clients:", err)
	}

	fmt.Printf("Clients generated in %s\n", finalOutputDir)
}

func generateClients(outputDir string, modules []ModuleMetadata, format string) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate client for each module
	for _, module := range modules {
		if err := generateModuleClient(outputDir, module, format); err != nil {
			return fmt.Errorf("failed to generate client for %s: %w", module.Name, err)
		}
	}

	// Generate index file
	if err := generateIndexFile(outputDir, modules); err != nil {
		return fmt.Errorf("failed to generate index file: %w", err)
	}

	return nil
}

func generateModuleClient(outputDir string, module ModuleMetadata, format string) error {
	// Generate types
	typesContent := generateTypesStub(module)
	typesPath := filepath.Join(outputDir, fmt.Sprintf("%s.types.ts", module.Name))
	if err := os.WriteFile(typesPath, []byte(typesContent), 0644); err != nil {
		return err
	}

	// Generate client
	clientContent := generateClient(module, format)
	clientPath := filepath.Join(outputDir, fmt.Sprintf("%s.client.ts", module.Name))
	if err := os.WriteFile(clientPath, []byte(clientContent), 0644); err != nil {
		return err
	}

	return nil
}

func generateTypesStub(module ModuleMetadata) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("// Types for %s module\n", module.Name))
	sb.WriteString("// Generated from Go structs\n\n")
	
	// Generate all types from the extracted type information
	for _, typeInfo := range module.Types {
		// Skip internal types like Handlers
		if shouldSkipType(typeInfo.Name) {
			continue
		}
		
		sb.WriteString(generateTypeInterface(typeInfo))
		sb.WriteString("\n")
	}
	
	// Generate additional types for routes that don't have explicit request/response types
	generatedTypes := make(map[string]bool)
	for _, typeInfo := range module.Types {
		generatedTypes[typeInfo.Name] = true
	}
	
	for _, route := range module.Routes {
		// Generate missing request types
		if route.RequestType == nil && (route.Method == "POST" || route.Method == "PUT" || route.Method == "PATCH") {
			typeName := getRequestTypeName(route)
			if !generatedTypes[typeName] {
				generatedTypes[typeName] = true
				sb.WriteString(fmt.Sprintf("export interface %s {\n", typeName))
				sb.WriteString("  // TODO: Define request fields\n")
				sb.WriteString("}\n\n")
			}
		}
		
		// Generate missing response types
		if route.ResponseType == nil && route.Method != "DELETE" {
			typeName := getResponseTypeName(route)
			if !generatedTypes[typeName] {
				generatedTypes[typeName] = true
				sb.WriteString(fmt.Sprintf("export interface %s {\n", typeName))
				sb.WriteString("  // TODO: Define response fields\n")
				sb.WriteString("}\n\n")
			}
		}
	}
	
	return sb.String()
}

func shouldSkipType(typeName string) bool {
	skipTypes := []string{"Handlers", "Service", "Repository", "Module", "Config"}
	for _, skip := range skipTypes {
		if strings.Contains(typeName, skip) {
			return true
		}
	}
	return false
}

func generateTypeInterface(typeInfo TypeInfo) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("export interface %s {\n", typeInfo.Name))
	
	for _, field := range typeInfo.Fields {
		// Skip private fields (lowercase first letter) or fields with json:"-"
		if field.JSONName == "-" || (field.JSONName == "" && strings.ToLower(field.Name[:1]) == field.Name[:1]) {
			continue
		}
		
		fieldName := field.JSONName
		if fieldName == "" {
			fieldName = toSnakeCase(field.Name)
		}
		
		fieldType := field.Type
		optional := !field.Required || strings.Contains(fieldType, "| null")
		
		if optional {
			sb.WriteString(fmt.Sprintf("  %s?: %s;\n", fieldName, fieldType))
		} else {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
		}
	}
	
	sb.WriteString("}\n")
	return sb.String()
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func generateClient(module ModuleMetadata, format string) string {
	var sb strings.Builder
	
	sb.WriteString(fmt.Sprintf("// Client for %s module\n", module.Name))
	sb.WriteString("// Generated from route metadata\n\n")
	
	sb.WriteString(fmt.Sprintf("import type * as Types from './%s.types';\n", module.Name))
	
	switch format {
	case "axios":
		sb.WriteString(generateAxiosClient(module))
	case "fetch":
		sb.WriteString(generateFetchClient(module))
	default:
		sb.WriteString(generateAxiosClient(module))
	}
	
	return sb.String()
}

func generateAxiosClient(module ModuleMetadata) string {
	var sb strings.Builder
	
	sb.WriteString("import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';\n\n")
	
	className := capitalizeFirst(module.Name) + "Client"
	sb.WriteString(fmt.Sprintf("export class %s {\n", className))
	sb.WriteString("  private client: AxiosInstance;\n")
	sb.WriteString("  private basePath: string;\n")
	sb.WriteString("  private getToken?: () => string | null;\n\n")
	
	sb.WriteString("  constructor(\n")
	sb.WriteString("    baseURL: string,\n")
	sb.WriteString("    basePath = '/api/v1',\n")
	sb.WriteString("    options?: {\n")
	sb.WriteString("      config?: AxiosRequestConfig;\n")
	sb.WriteString("      getToken?: () => string | null;\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ) {\n")
	sb.WriteString("    this.basePath = basePath;\n")
	sb.WriteString("    this.getToken = options?.getToken;\n")
	sb.WriteString("    this.client = axios.create({\n")
	sb.WriteString("      baseURL,\n")
	sb.WriteString("      ...options?.config,\n")
	sb.WriteString("    });\n")
	sb.WriteString("  }\n\n")
	
	sb.WriteString("  private getAuthConfig(config?: AxiosRequestConfig): AxiosRequestConfig {\n")
	sb.WriteString("    const token = this.getToken?.();\n")
	sb.WriteString("    if (token) {\n")
	sb.WriteString("      return {\n")
	sb.WriteString("        ...config,\n")
	sb.WriteString("        headers: {\n")
	sb.WriteString("          ...config?.headers,\n")
	sb.WriteString("          Authorization: `Bearer ${token}`,\n")
	sb.WriteString("        },\n")
	sb.WriteString("      };\n")
	sb.WriteString("    }\n")
	sb.WriteString("    return config || {};\n")
	sb.WriteString("  }\n\n")
	
	// Group routes by resource for better organization
	for _, route := range module.Routes {
		sb.WriteString(generateAxiosMethod(route, module.Name))
		sb.WriteString("\n")
	}
	
	sb.WriteString("}\n")
	return sb.String()
}

func generateAxiosMethod(route RouteMetadata, moduleName string) string {
	var sb strings.Builder
	
	methodName := getMethodName(route)
	
	// Build parameters
	params := []string{}
	
	// Add path parameters
	for _, param := range route.PathParams {
		params = append(params, fmt.Sprintf("%s: string", param))
	}
	
	// Add request body for mutations
	if route.Method == "POST" || route.Method == "PUT" || route.Method == "PATCH" {
		var requestType string
		if route.RequestType != nil {
			requestType = "Types." + route.RequestType.Name
		} else {
			requestType = "Types." + getRequestTypeName(route)
		}
		params = append(params, fmt.Sprintf("data: %s", requestType))
	}
	
	// Add query parameters for GET requests that might have filters
	if route.Method == "GET" && (strings.Contains(methodName, "list") || strings.Contains(methodName, "search")) {
		params = append(params, "params?: Record<string, any>")
	}
	
	// Always add config as last parameter
	params = append(params, "config?: AxiosRequestConfig")
	
	// Determine return type
	returnType := "void"
	if route.Method != "DELETE" {
		if route.ResponseType != nil {
			returnType = "Types." + route.ResponseType.Name
		} else {
			returnType = "Types." + getResponseTypeName(route)
		}
	}
	
	// Add JSDoc
	sb.WriteString(fmt.Sprintf("  /**\n   * %s\n", route.Description))
	if route.Name != "" {
		sb.WriteString(fmt.Sprintf("   * @route %s %s\n", route.Method, route.Path))
	}
	
	// Add parameter documentation
	for _, param := range route.PathParams {
		sb.WriteString(fmt.Sprintf("   * @param %s - Path parameter\n", param))
	}
	if route.RequestType != nil {
		sb.WriteString(fmt.Sprintf("   * @param data - Request body of type %s\n", route.RequestType.Name))
	}
	
	sb.WriteString("   */\n")
	
	// Method signature
	sb.WriteString(fmt.Sprintf("  async %s(%s): Promise<%s> {\n", methodName, strings.Join(params, ", "), returnType))
	
	// Build URL
	url := buildURLTemplate(route.Path, route.PathParams)
	
	// Determine if this endpoint needs authentication
	needsAuth := !isPublicEndpoint(route, methodName)
	
	// Method implementation
	switch route.Method {
	case "GET":
		if strings.Contains(methodName, "list") || strings.Contains(methodName, "search") {
			if needsAuth {
				sb.WriteString(fmt.Sprintf("    const response = await this.client.get(`${this.basePath}%s`, { params, ...this.getAuthConfig(config) });\n", url))
			} else {
				sb.WriteString(fmt.Sprintf("    const response = await this.client.get(`${this.basePath}%s`, { params, ...config });\n", url))
			}
		} else {
			if needsAuth {
				sb.WriteString(fmt.Sprintf("    const response = await this.client.get(`${this.basePath}%s`, this.getAuthConfig(config));\n", url))
			} else {
				sb.WriteString(fmt.Sprintf("    const response = await this.client.get(`${this.basePath}%s`, config);\n", url))
			}
		}
	case "POST":
		if needsAuth {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.post(`${this.basePath}%s`, data, this.getAuthConfig(config));\n", url))
		} else {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.post(`${this.basePath}%s`, data, config);\n", url))
		}
	case "PUT":
		if needsAuth {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.put(`${this.basePath}%s`, data, this.getAuthConfig(config));\n", url))
		} else {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.put(`${this.basePath}%s`, data, config);\n", url))
		}
	case "PATCH":
		if needsAuth {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.patch(`${this.basePath}%s`, data, this.getAuthConfig(config));\n", url))
		} else {
			sb.WriteString(fmt.Sprintf("    const response = await this.client.patch(`${this.basePath}%s`, data, config);\n", url))
		}
	case "DELETE":
		if needsAuth {
			sb.WriteString(fmt.Sprintf("    await this.client.delete(`${this.basePath}%s`, this.getAuthConfig(config));\n", url))
		} else {
			sb.WriteString(fmt.Sprintf("    await this.client.delete(`${this.basePath}%s`, config);\n", url))
		}
		sb.WriteString("  }\n")
		return sb.String()
	}
	
	sb.WriteString("    return response.data;\n")
	sb.WriteString("  }\n")
	
	return sb.String()
}

func generateFetchClient(module ModuleMetadata) string {
	// TODO: Implement fetch-based client
	return "// Fetch client implementation\n"
}

func generateIndexFile(outputDir string, modules []ModuleMetadata) error {
	var sb strings.Builder
	
	sb.WriteString("// Generated index file\n\n")
	
	// Export all types and import clients for use in factory function
	for _, module := range modules {
		sb.WriteString(fmt.Sprintf("export * from './%s.types';\n", module.Name))
		sb.WriteString(fmt.Sprintf("import { %sClient } from './%s.client';\n", 
			capitalizeFirst(module.Name), module.Name))
	}
	
	sb.WriteString("\n// Client factory function\n")
	sb.WriteString("export interface ClientOptions {\n")
	sb.WriteString("  baseURL: string;\n")
	sb.WriteString("  basePath?: string;\n")
	sb.WriteString("  getToken?: () => string | null;\n")
	sb.WriteString("  config?: any; // AxiosRequestConfig\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("export function createClients(options: ClientOptions) {\n")
	sb.WriteString("  const { baseURL, basePath = '/api/v1', getToken, config } = options;\n")
	sb.WriteString("  const clientOptions = { config, getToken };\n\n")
	sb.WriteString("  return {\n")
	for _, module := range modules {
		sb.WriteString(fmt.Sprintf("    %s: new %sClient(baseURL, basePath, clientOptions),\n", 
			module.Name, capitalizeFirst(module.Name)))
	}
	sb.WriteString("  };\n")
	sb.WriteString("}\n\n")
	
	sb.WriteString("// Individual client classes for manual instantiation\n")
	sb.WriteString("export {\n")
	for _, module := range modules {
		sb.WriteString(fmt.Sprintf("  %sClient,\n", capitalizeFirst(module.Name)))
	}
	sb.WriteString("};\n")
	
	indexPath := filepath.Join(outputDir, "index.ts")
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// Helper functions

func isPublicEndpoint(route RouteMetadata, methodName string) bool {
	// Health endpoints are always public
	if strings.Contains(route.Path, "/health") {
		return true
	}
	
	// Auth endpoints that are typically public (don't require existing auth)
	publicAuthEndpoints := []string{
		"/auth/register", "/auth/login", "/auth/forgot-password", 
		"/auth/reset-password", "/auth/verify-email",
	}
	
	for _, publicPath := range publicAuthEndpoints {
		if route.Path == publicPath {
			return true
		}
	}
	
	// If it's an auth endpoint not in the public list, it probably needs auth
	// All other modules (role, etc.) typically need authentication
	return false
}

func buildURLTemplate(path string, params []string) string {
	result := path
	for _, param := range params {
		result = strings.Replace(result, ":"+param, "${"+param+"}", 1)
	}
	return result
}

func getMethodName(route RouteMetadata) string {
	if route.Name != "" {
		// Convert route name like "auth.verify-email" to "verifyEmail"
		parts := strings.Split(route.Name, ".")
		if len(parts) > 1 {
			methodName := parts[1]
			// Convert kebab-case to camelCase
			methodName = strings.ReplaceAll(methodName, "-", "_")
			return toCamelCase(methodName)
		}
	}
	
	// Fallback: generate from method and path
	path := strings.TrimPrefix(route.Path, "/")
	path = strings.ReplaceAll(path, "/", "_")
	path = strings.ReplaceAll(path, ":", "by_")
	path = strings.ReplaceAll(path, "-", "_")
	
	return toCamelCase(strings.ToLower(route.Method) + "_" + path)
}

func getRequestTypeName(route RouteMetadata) string {
	methodName := getMethodName(route)
	return capitalizeFirst(methodName) + "Request"
}

func getResponseTypeName(route RouteMetadata) string {
	methodName := getMethodName(route)
	
	// Special cases for common patterns
	if strings.HasPrefix(methodName, "list") || strings.HasSuffix(methodName, "s") {
		// listRoles -> ListRolesResponse (not Role[])
		// getUserRoles -> GetUserRolesResponse (not Role[])
		return capitalizeFirst(methodName) + "Response"
	}
	
	return capitalizeFirst(methodName) + "Response"
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
		} else {
			parts[i] = capitalizeFirst(strings.ToLower(parts[i]))
		}
	}
	return strings.Join(parts, "")
}