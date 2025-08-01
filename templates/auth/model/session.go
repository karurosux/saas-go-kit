package authmodel

import (
	"time"
	
	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	UserID             uuid.UUID `json:"user_id"`
	Token              string    `json:"token"`
	RefreshToken       string    `json:"refresh_token"`
	ExpiresAt          time.Time `json:"expires_at"`
	RefreshExpiresAt   time.Time `json:"refresh_expires_at"`
}

// GetUserID returns the user ID
func (s *Session) GetUserID() uuid.UUID {
	return s.UserID
}

// GetToken returns the session token
func (s *Session) GetToken() string {
	return s.Token
}

// GetRefreshToken returns the refresh token
func (s *Session) GetRefreshToken() string {
	return s.RefreshToken
}

// GetExpiresAt returns the token expiration time
func (s *Session) GetExpiresAt() time.Time {
	return s.ExpiresAt
}

// GetRefreshExpiresAt returns the refresh token expiration time
func (s *Session) GetRefreshExpiresAt() time.Time {
	return s.RefreshExpiresAt
}

// IsExpired checks if the session token has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRefreshExpired checks if the refresh token has expired
func (s *Session) IsRefreshExpired() bool {
	return time.Now().After(s.RefreshExpiresAt)
}