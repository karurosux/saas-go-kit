package authservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
	authmodel "{{.Project.GoModule}}/internal/auth/model"
	"{{.Project.GoModule}}/internal/core"
	"github.com/google/uuid"
)

type GoogleOAuthStrategy struct {
	accountRepo    authinterface.AccountRepository
	clientID       string
	clientSecret   string
	redirectURI    string
	scopes         []string
}

func NewGoogleOAuthStrategy(
	accountRepo authinterface.AccountRepository,
	clientID string,
	clientSecret string,
	redirectURI string,
) authinterface.OAuthStrategy {
	return &GoogleOAuthStrategy{
		accountRepo:  accountRepo,
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		scopes:       []string{"openid", "email", "profile"},
	}
}

func (s *GoogleOAuthStrategy) Name() string {
	return "google"
}

func (s *GoogleOAuthStrategy) Type() authinterface.AuthStrategyType {
	return authinterface.StrategyTypeOAuth
}

func (s *GoogleOAuthStrategy) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", s.clientID)
	params.Add("redirect_uri", s.redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", strings.Join(s.scopes, " "))
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")
	
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s", params.Encode())
}

func (s *GoogleOAuthStrategy) ExchangeCode(ctx context.Context, code string) (*authinterface.OAuthTokens, error) {
	tokenURL := "https://oauth2.googleapis.com/token"
	
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("redirect_uri", s.redirectURI)
	data.Set("grant_type", "authorization_code")
	
	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to exchange code: %s", string(body))
	}
	
	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	
	return &authinterface.OAuthTokens{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

func (s *GoogleOAuthStrategy) GetUserInfo(ctx context.Context, tokens *authinterface.OAuthTokens) (*authinterface.OAuthUserInfo, error) {
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens.AccessToken))
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}
	
	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	
	return &authinterface.OAuthUserInfo{
		ID:            googleUser.ID,
		Email:         googleUser.Email,
		EmailVerified: googleUser.VerifiedEmail,
		Name:          googleUser.Name,
		Picture:       googleUser.Picture,
		Provider:      "google",
		Raw: map[string]any{
			"id":             googleUser.ID,
			"email":          googleUser.Email,
			"verified_email": googleUser.VerifiedEmail,
			"name":           googleUser.Name,
			"picture":        googleUser.Picture,
		},
	}, nil
}

func (s *GoogleOAuthStrategy) Authenticate(ctx context.Context, credentials map[string]any) (*authinterface.AuthResult, error) {
	code, ok := credentials["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}
	
	tokens, err := s.ExchangeCode(ctx, code)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, "failed to exchange authorization code")
	}
	
	userInfo, err := s.GetUserInfo(ctx, tokens)
	if err != nil {
		return nil, core.NewAppError(core.ErrCodeUnauthorized, "failed to get user information")
	}
	
	account, err := s.accountRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		account = &authmodel.Account{
			ID:            uuid.New(),
			Email:         userInfo.Email,
			EmailVerified: userInfo.EmailVerified,
			PasswordHash:  "", // No password for OAuth users
		}
		
		if err := s.accountRepo.Create(ctx, account); err != nil {
			return nil, core.NewAppError(core.ErrCodeInternalServer, "failed to create account")
		}
	}
	
	return &authinterface.AuthResult{
		AccountID:         account.GetID(),
		Account:           account,
		NeedsVerification: false, // Google already verified the email
		Metadata: map[string]any{
			"auth_method": "google_oauth",
			"provider_id": userInfo.ID,
			"name":        userInfo.Name,
			"picture":     userInfo.Picture,
		},
	}, nil
}

func (s *GoogleOAuthStrategy) ValidateCredentials(credentials map[string]any) error {
	code, ok := credentials["code"].(string)
	if !ok || code == "" {
		return fmt.Errorf("authorization code is required")
	}
	
	return nil
}