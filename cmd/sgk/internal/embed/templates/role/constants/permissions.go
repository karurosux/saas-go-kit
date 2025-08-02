package roleconstants

const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionList   = "list"
	
	PermissionAll = "*"
	PermissionAdmin = "admin:*"
	
	PermissionSeparator = ":"
)

const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleModerator = "moderator"
)

const (
	RoleAdminDesc     = "Full system administrator with all permissions"
	RoleUserDesc      = "Standard user with basic permissions"
	RoleModeratorDesc = "Moderator with content management permissions"
)