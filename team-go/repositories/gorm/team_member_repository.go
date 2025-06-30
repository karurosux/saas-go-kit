package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/team-go"
	"gorm.io/gorm"
)

type TeamMemberRepository struct {
	db *gorm.DB
}

func NewTeamMemberRepository(db *gorm.DB) team.TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

func (r *TeamMemberRepository) Create(ctx context.Context, member *team.TeamMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *TeamMemberRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*team.TeamMember, error) {
	var member team.TeamMember
	query := r.db.WithContext(ctx)
	
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	
	err := query.First(&member, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrTeamMemberNotFound
		}
		return nil, err
	}
	
	return &member, nil
}

func (r *TeamMemberRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]team.TeamMember, error) {
	var members []team.TeamMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("account_id = ?", accountID).
		Order("created_at ASC").
		Find(&members).Error
	
	if err != nil {
		return nil, err
	}
	
	return members, nil
}

func (r *TeamMemberRepository) FindByUserAndAccount(ctx context.Context, userID, accountID uuid.UUID) (*team.TeamMember, error) {
	var member team.TeamMember
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ? AND account_id = ?", userID, accountID).
		First(&member).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrTeamMemberNotFound
		}
		return nil, err
	}
	
	return &member, nil
}

func (r *TeamMemberRepository) Update(ctx context.Context, member *team.TeamMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *TeamMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&team.TeamMember{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return team.ErrTeamMemberNotFound
	}
	
	return nil
}

func (r *TeamMemberRepository) CountByAccountID(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&team.TeamMember{}).
		Where("account_id = ?", accountID).
		Count(&count).Error
	
	return count, err
}

func (r *TeamMemberRepository) GetTeamStats(ctx context.Context, accountID uuid.UUID) (*team.TeamStats, error) {
	var members []team.TeamMember
	err := r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Find(&members).Error
	
	if err != nil {
		return nil, err
	}
	
	stats := &team.TeamStats{
		TotalMembers:  len(members),
		RoleBreakdown: make(map[string]int),
	}
	
	for _, member := range members {
		// Count role breakdown
		roleStr := string(member.Role)
		stats.RoleBreakdown[roleStr]++
		
		// Count active vs pending
		if member.IsActive() {
			stats.ActiveMembers++
		} else {
			stats.PendingMembers++
		}
	}
	
	return stats, nil
}