package role

import (
	"regexp"
	"strings"
)

type permissionUtils struct{}

// NewPermissionUtils creates a new permission utilities instance
func NewPermissionUtils() PermissionUtils {
	return &permissionUtils{}
}

// ParsePermission parses a permission tag into resource and action
// Example: "users:read" -> ("users", "read")
func (p *permissionUtils) ParsePermission(permission string) (resource, action string) {
	parts := strings.SplitN(permission, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// MatchesPattern checks if a permission matches a pattern with wildcard support
// Examples:
// - "users:read" matches "users:*"
// - "users:read" matches "*:read"
// - "users:read" matches "*"
func (p *permissionUtils) MatchesPattern(permission, pattern string) bool {
	return matchesPermission(permission, pattern)
}

// BuildPermission builds a permission tag from resource and action
// Example: ("users", "read") -> "users:read"
func (p *permissionUtils) BuildPermission(resource, action string) string {
	if resource == "" || action == "" {
		return ""
	}
	return resource + ":" + action
}

// IsValidPermission validates if a permission string follows the correct format
// Valid formats: "resource:action", "*", "resource:*", "*:action"
func (p *permissionUtils) IsValidPermission(permission string) bool {
	if permission == "" {
		return false
	}

	// Global wildcard
	if permission == "*" {
		return true
	}

	// Must contain exactly one colon
	if strings.Count(permission, ":") != 1 {
		return false
	}

	resource, action := p.ParsePermission(permission)
	
	// Both parts must be non-empty
	if resource == "" || action == "" {
		return false
	}

	// Validate resource and action names (alphanumeric, hyphens, underscores, or wildcard)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$|^\*$`)
	
	return validName.MatchString(resource) && validName.MatchString(action)
}