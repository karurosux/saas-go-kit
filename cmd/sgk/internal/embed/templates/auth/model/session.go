package authmodel

import (
	"time"
	
	"github.com/google/uuid"
)

type Session struct {
	UserID             uuid.UUID `json:"user_id"`
	Token              string    `json:"token"`
	RefreshToken       string    `json:"refresh_token"`
	ExpiresAt          time.Time `json:"expires_at"`
	RefreshExpiresAt   time.Time `json:"refresh_expires_at"`
}

func (s *Session) GetUserID() uuid.UUID {
	return s.UserID
}

func (s *Session) GetToken() string {
	return s.Token
}

func (s *Session) GetRefreshToken() string {
	return s.RefreshToken
}

func (s *Session) GetExpiresAt() time.Time {
	return s.ExpiresAt
}

func (s *Session) GetRefreshExpiresAt() time.Time {
	return s.RefreshExpiresAt
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsRefreshExpired() bool {
	return time.Now().After(s.RefreshExpiresAt)
}