package roleconstants

// Context keys for storing role-related data
const (
	// ContextKeyUserPermissions stores the user's permissions in context
	ContextKeyUserPermissions = "user_permissions"
	
	// ContextKeyUserRoles stores the user's roles in context
	ContextKeyUserRoles = "user_roles"
	
	// ContextKeyHasPermissionPrefix is the prefix for permission check results
	ContextKeyHasPermissionPrefix = "has_permission_"
)