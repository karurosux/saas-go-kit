package gorm

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/karurosux/saas-go-kit/team-go"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) team.UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *team.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID, preloads ...string) (*team.User, error) {
	var user team.User
	query := r.db.WithContext(ctx)
	
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	
	err := query.First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*team.User, error) {
	var user team.User
	err := r.db.WithContext(ctx).
		Preload("TeamMembers").
		First(&user, "email = ?", email).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, team.ErrUserNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *team.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&team.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		return team.ErrUserNotFound
	}
	
	return nil
}