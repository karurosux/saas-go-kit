package authinterface

import (
	"context"
	
	"github.com/google/uuid"
)

type AuthStrategy interface {
	Name() string
	Type() AuthStrategyType
	Authenticate(ctx context.Context, credentials map[string]any) (*AuthResult, error)
	ValidateCredentials(credentials map[string]any) error
}

type OAuthStrategy interface {
	AuthStrategy
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)
	GetUserInfo(ctx context.Context, tokens *OAuthTokens) (*OAuthUserInfo, error)
}

type AuthStrategyType string

const (
	StrategyTypeLocal AuthStrategyType = "local"
	StrategyTypeOAuth AuthStrategyType = "oauth"
	StrategyTypeSAML  AuthStrategyType = "saml"
)

type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
}

type OAuthUserInfo struct {
	ID            string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Provider      string
	Raw           map[string]any
}

type AuthResult struct {
	AccountID      uuid.UUID
	Account        Account
	NeedsVerification bool
	Metadata       map[string]any
}

type StrategyRegistry interface {
	Register(strategy AuthStrategy) error
	Get(name string) (AuthStrategy, error)
	List() []string
}