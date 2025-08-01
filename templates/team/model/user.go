package teammodel

import (
	"fmt"
	"time"
	
	"github.com/google/uuid"
)

// User represents a user
type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetID returns the user ID
func (u *User) GetID() uuid.UUID {
	return u.ID
}

// GetEmail returns the user email
func (u *User) GetEmail() string {
	return u.Email
}

// GetFirstName returns the user first name
func (u *User) GetFirstName() string {
	return u.FirstName
}

// GetLastName returns the user last name
func (u *User) GetLastName() string {
	return u.LastName
}

// GetFullName returns the user full name
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
}

// GetCreatedAt returns creation time
func (u *User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

// GetUpdatedAt returns last update time
func (u *User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}