package teammodel

import (
	"{{.Project.GoModule}}/internal/team/interface"
)

// TeamStats represents team statistics
type TeamStats struct {
	TotalMembers       int64                                      `json:"total_members"`
	ActiveMembers      int64                                      `json:"active_members"`
	PendingInvitations int64                                      `json:"pending_invitations"`
	MembersByRole      map[teaminterface.MemberRole]int64        `json:"members_by_role"`
}

// GetTotalMembers returns total members count
func (ts *TeamStats) GetTotalMembers() int64 {
	return ts.TotalMembers
}

// GetActiveMembers returns active members count
func (ts *TeamStats) GetActiveMembers() int64 {
	return ts.ActiveMembers
}

// GetPendingInvitations returns pending invitations count
func (ts *TeamStats) GetPendingInvitations() int64 {
	return ts.PendingInvitations
}

// GetMembersByRole returns members count by role
func (ts *TeamStats) GetMembersByRole() map[teaminterface.MemberRole]int64 {
	return ts.MembersByRole
}