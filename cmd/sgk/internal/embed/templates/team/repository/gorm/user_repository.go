package gorm

import (
	"context"
	"errors"
	
	"{{.Project.GoModule}}/internal/team/interface"
	"{{.Project.GoModule}}/internal/team/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository implements user repository using GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) teaminterface.UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user teaminterface.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (teaminterface.User, error) {
	var user teammodel.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (teaminterface.User, error) {
	var user teammodel.User
	err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user teaminterface.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&teammodel.User{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&teammodel.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&teammodel.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}