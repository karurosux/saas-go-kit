package teamconstants

// Permission constants for team operations
const (
	// PermissionTeamView allows viewing team members
	PermissionTeamView = "team:view"
	
	// PermissionTeamInvite allows inviting new team members
	PermissionTeamInvite = "team:invite"
	
	// PermissionTeamUpdate allows updating team member roles
	PermissionTeamUpdate = "team:update"
	
	// PermissionTeamRemove allows removing team members
	PermissionTeamRemove = "team:remove"
	
	// PermissionTeamManage allows all team management operations
	PermissionTeamManage = "team:manage"
)

// Role permissions mapping
var RolePermissions = map[string][]string{
	"owner": {
		PermissionTeamView,
		PermissionTeamInvite,
		PermissionTeamUpdate,
		PermissionTeamRemove,
		PermissionTeamManage,
	},
	"admin": {
		PermissionTeamView,
		PermissionTeamInvite,
		PermissionTeamUpdate,
		PermissionTeamRemove,
	},
	"member": {
		PermissionTeamView,
	},
	"viewer": {
		PermissionTeamView,
	},
}