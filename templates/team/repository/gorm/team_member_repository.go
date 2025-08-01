package gorm

import (
	"context"
	"errors"
	
	"{{.Project.GoModule}}/internal/team/interface"
	"{{.Project.GoModule}}/internal/team/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TeamMemberRepository implements team member repository using GORM
type TeamMemberRepository struct {
	db *gorm.DB
}

// NewTeamMemberRepository creates a new team member repository
func NewTeamMemberRepository(db *gorm.DB) teaminterface.TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

// Create creates a new team member
func (r *TeamMemberRepository) Create(ctx context.Context, member teaminterface.TeamMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// GetByID gets a team member by ID
func (r *TeamMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (teaminterface.TeamMember, error) {
	var member teammodel.TeamMember
	err := r.db.WithContext(ctx).Preload("User").First(&member, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team member not found")
		}
		return nil, err
	}
	return &member, nil
}

// GetByAccountID gets all team members for an account
func (r *TeamMemberRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]teaminterface.TeamMember, error) {
	var members []teammodel.TeamMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("account_id = ? AND is_active = ?", accountID, true).
		Order("created_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to interface slice
	result := make([]teaminterface.TeamMember, len(members))
	for i, m := range members {
		member := m // Create a copy to avoid pointer issues
		result[i] = &member
	}
	return result, nil
}

// GetByUserAndAccount gets a team member by user ID and account ID
func (r *TeamMemberRepository) GetByUserAndAccount(ctx context.Context, userID, accountID uuid.UUID) (teaminterface.TeamMember, error) {
	var member teammodel.TeamMember
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&member, "user_id = ? AND account_id = ?", userID, accountID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("team member not found")
		}
		return nil, err
	}
	return &member, nil
}

// Update updates a team member
func (r *TeamMemberRepository) Update(ctx context.Context, member teaminterface.TeamMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

// Delete deletes a team member
func (r *TeamMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&teammodel.TeamMember{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("team member not found")
	}
	return nil
}

// CountByAccountID counts team members for an account
func (r *TeamMemberRepository) CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&teammodel.TeamMember{}).
		Where("account_id = ? AND is_active = ?", accountID, true).
		Count(&count).Error
	return count, err
}

// GetTeamStats gets team statistics
func (r *TeamMemberRepository) GetTeamStats(ctx context.Context, accountID uuid.UUID) (teaminterface.TeamStats, error) {
	stats := &teammodel.TeamStats{
		MembersByRole: make(map[teaminterface.MemberRole]int64),
	}
	
	// Get total members
	err := r.db.WithContext(ctx).
		Model(&teammodel.TeamMember{}).
		Where("account_id = ?", accountID).
		Count(&stats.TotalMembers).Error
	if err != nil {
		return nil, err
	}
	
	// Get active members
	err = r.db.WithContext(ctx).
		Model(&teammodel.TeamMember{}).
		Where("account_id = ? AND is_active = ? AND is_pending = ?", accountID, true, false).
		Count(&stats.ActiveMembers).Error
	if err != nil {
		return nil, err
	}
	
	// Get pending invitations
	err = r.db.WithContext(ctx).
		Model(&teammodel.TeamMember{}).
		Where("account_id = ? AND is_active = ? AND is_pending = ?", accountID, true, true).
		Count(&stats.PendingInvitations).Error
	if err != nil {
		return nil, err
	}
	
	// Get members by role
	type roleCount struct {
		Role  teaminterface.MemberRole
		Count int64
	}
	
	var roleCounts []roleCount
	err = r.db.WithContext(ctx).
		Model(&teammodel.TeamMember{}).
		Select("role, COUNT(*) as count").
		Where("account_id = ? AND is_active = ?", accountID, true).
		Group("role").
		Scan(&roleCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, rc := range roleCounts {
		stats.MembersByRole[rc.Role] = rc.Count
	}
	
	return stats, nil
}