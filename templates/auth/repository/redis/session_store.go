package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	
	"{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/auth/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// SessionStore implements session storage using Redis
type SessionStore struct {
	client *redis.Client
	prefix string
}

// NewSessionStore creates a new Redis session store
func NewSessionStore(client *redis.Client, prefix string) authinterface.SessionStore {
	if prefix == "" {
		prefix = "session"
	}
	return &SessionStore{
		client: client,
		prefix: prefix,
	}
}

// Store stores a session
func (s *SessionStore) Store(ctx context.Context, session authinterface.Session) error {
	// Marshal session
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	
	// Calculate TTL
	ttl := time.Until(session.GetExpiresAt())
	if ttl <= 0 {
		return errors.New("session already expired")
	}
	
	// Store by token
	tokenKey := s.key("token", session.GetToken())
	if err := s.client.Set(ctx, tokenKey, data, ttl).Err(); err != nil {
		return err
	}
	
	// Store by refresh token (with refresh TTL)
	refreshTTL := time.Until(session.GetRefreshExpiresAt())
	if refreshTTL > 0 {
		refreshKey := s.key("refresh", session.GetRefreshToken())
		if err := s.client.Set(ctx, refreshKey, data, refreshTTL).Err(); err != nil {
			// Clean up token key on failure
			s.client.Del(ctx, tokenKey)
			return err
		}
	}
	
	// Store user-to-session mapping
	userKey := s.key("user", session.GetUserID().String())
	if err := s.client.Set(ctx, userKey, session.GetToken(), ttl).Err(); err != nil {
		// Clean up on failure
		s.client.Del(ctx, tokenKey)
		s.client.Del(ctx, s.key("refresh", session.GetRefreshToken()))
		return err
	}
	
	return nil
}

// Get gets a session by token
func (s *SessionStore) Get(ctx context.Context, token string) (authinterface.Session, error) {
	key := s.key("token", token)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	
	var session authmodel.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	
	return &session, nil
}

// GetByRefreshToken gets a session by refresh token
func (s *SessionStore) GetByRefreshToken(ctx context.Context, refreshToken string) (authinterface.Session, error) {
	key := s.key("refresh", refreshToken)
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	
	var session authmodel.Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	
	return &session, nil
}

// Delete deletes all sessions for a user
func (s *SessionStore) Delete(ctx context.Context, userID uuid.UUID) error {
	userKey := s.key("user", userID.String())
	
	// Get the token first
	token, err := s.client.Get(ctx, userKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	
	// If we have a token, get the session to find the refresh token
	if token != "" {
		session, err := s.Get(ctx, token)
		if err == nil {
			// Delete refresh token key
			s.client.Del(ctx, s.key("refresh", session.GetRefreshToken()))
		}
		// Delete token key
		s.client.Del(ctx, s.key("token", token))
	}
	
	// Delete user key
	return s.client.Del(ctx, userKey).Err()
}

// DeleteByToken deletes a session by token
func (s *SessionStore) DeleteByToken(ctx context.Context, token string) error {
	// Get session first to find associated keys
	session, err := s.Get(ctx, token)
	if err != nil {
		return err
	}
	
	// Delete all associated keys
	keys := []string{
		s.key("token", token),
		s.key("refresh", session.GetRefreshToken()),
		s.key("user", session.GetUserID().String()),
	}
	
	return s.client.Del(ctx, keys...).Err()
}

// DeleteExpired deletes expired sessions (handled automatically by Redis TTL)
func (s *SessionStore) DeleteExpired(ctx context.Context) error {
	// Redis handles expiration automatically via TTL
	// This method exists to satisfy the interface but doesn't need to do anything
	return nil
}

// key generates a Redis key
func (s *SessionStore) key(parts ...string) string {
	key := s.prefix
	for _, part := range parts {
		key = fmt.Sprintf("%s:%s", key, part)
	}
	return key
}