package authservice

import (
	"context"
	"fmt"
	
	authconstants "{{.Project.GoModule}}/internal/auth/constants"
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	"{{.Project.GoModule}}/internal/core"
)

type EmailPasswordStrategy struct {
	accountRepo    authinterface.AccountRepository
	passwordHasher authinterface.PasswordHasher
}

func NewEmailPasswordStrategy(
	accountRepo authinterface.AccountRepository,
	passwordHasher authinterface.PasswordHasher,
) authinterface.AuthStrategy {
	return &EmailPasswordStrategy{
		accountRepo:    accountRepo,
		passwordHasher: passwordHasher,
	}
}

func (s *EmailPasswordStrategy) Name() string {
	return "email_password"
}

func (s *EmailPasswordStrategy) Type() authinterface.AuthStrategyType {
	return authinterface.StrategyTypeLocal
}

func (s *EmailPasswordStrategy) Authenticate(ctx context.Context, credentials map[string]any) (*authinterface.AuthResult, error) {
	email, ok := credentials["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email is required")
	}
	
	password, ok := credentials["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("password is required")
	}
	
	account, err := s.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidCredentials)
	}
	
	if err := s.passwordHasher.Verify(password, account.GetPasswordHash()); err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, authconstants.ErrInvalidCredentials)
	}
	
	return &authinterface.AuthResult{
		AccountID:         account.GetID(),
		Account:           account,
		NeedsVerification: !account.GetEmailVerified(),
		Metadata: map[string]any{
			"auth_method": "email_password",
		},
	}, nil
}

func (s *EmailPasswordStrategy) ValidateCredentials(credentials map[string]any) error {
	email, ok := credentials["email"].(string)
	if !ok || email == "" {
		return fmt.Errorf("email is required")
	}
	
	password, ok := credentials["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("password is required")
	}
	
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	
	return nil
}