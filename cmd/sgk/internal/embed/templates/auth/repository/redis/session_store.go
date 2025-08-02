package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	client *redis.Client
	prefix string
}

func NewSessionStore(client *redis.Client, prefix string) authinterface.SessionStore {
	if prefix == "" {
		prefix = "session"
	}
	return &SessionStore{
		client: client,
		prefix: prefix,
	}
}

func (s *SessionStore) Store(ctx context.Context, session authinterface.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	
	ttl := time.Until(session.GetExpiresAt())
	if ttl <= 0 {
		return errors.New("session already expired")
	}
	
	tokenKey := s.key("token", session.GetToken())
	if err := s.client.Set(ctx, tokenKey, data, ttl).Err(); err != nil {
		return err
	}
	
	refreshTTL := time.Until(session.GetRefreshExpiresAt())
	if refreshTTL > 0 {
		refreshKey := s.key("refresh", session.GetRefreshToken())
		if err := s.client.Set(ctx, refreshKey, data, refreshTTL).Err(); err != nil {
			s.client.Del(ctx, tokenKey)
			return err
		}
	}
	
	userKey := s.key("user", session.GetUserID().String())
	if err := s.client.Set(ctx, userKey, session.GetToken(), ttl).Err(); err != nil {
		s.client.Del(ctx, tokenKey)
		s.client.Del(ctx, s.key("refresh", session.GetRefreshToken()))
		return err
	}
	
	return nil
}

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

func (s *SessionStore) Delete(ctx context.Context, userID uuid.UUID) error {
	userKey := s.key("user", userID.String())
	
	token, err := s.client.Get(ctx, userKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	
	if token != "" {
		session, err := s.Get(ctx, token)
		if err == nil {
			s.client.Del(ctx, s.key("refresh", session.GetRefreshToken()))
		}
		s.client.Del(ctx, s.key("token", token))
	}
	
	return s.client.Del(ctx, userKey).Err()
}

func (s *SessionStore) DeleteByToken(ctx context.Context, token string) error {
	session, err := s.Get(ctx, token)
	if err != nil {
		return err
	}
	
	keys := []string{
		s.key("token", token),
		s.key("refresh", session.GetRefreshToken()),
		s.key("user", session.GetUserID().String()),
	}
	
	return s.client.Del(ctx, keys...).Err()
}

func (s *SessionStore) DeleteExpired(ctx context.Context) error {
	return nil
}

func (s *SessionStore) key(parts ...string) string {
	key := s.prefix
	for _, part := range parts {
		key = fmt.Sprintf("%s:%s", key, part)
	}
	return key
}