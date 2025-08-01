package roleconstants

// Common permission patterns and constants
const (
	// Resource actions
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	
	// Special permissions
	PermissionAll = "*"
	PermissionAdmin = "admin:*"
	
	// Permission separator
	PermissionSeparator = ":"
)

// Default system roles
const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleModerator = "moderator"
)

// Default role descriptions
const (
	RoleAdminDesc     = "Full system administrator with all permissions"
	RoleUserDesc      = "Standard user with basic permissions"
	RoleModeratorDesc = "Moderator with content management permissions"
)