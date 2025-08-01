package roleconstants

// Error messages
const (
	ErrRoleNotFound          = "role not found"
	ErrRoleAlreadyExists     = "role already exists"
	ErrSystemRoleDelete      = "cannot delete system role"
	ErrInvalidPermission     = "invalid permission format"
	ErrUserRoleNotFound      = "user role assignment not found"
	ErrRoleAlreadyAssigned   = "role already assigned to user"
	ErrSelfRoleModification  = "cannot modify own role assignments"
)