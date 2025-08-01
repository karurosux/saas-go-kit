package authservice

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	
	"{{.Project.GoModule}}/internal/auth/interface"
)

// DefaultTokenGenerator implements token generation
type DefaultTokenGenerator struct{}

// NewTokenGenerator creates a new token generator
func NewTokenGenerator() authinterface.TokenGenerator {
	return &DefaultTokenGenerator{}
}

// GenerateToken generates a simple numeric token (for SMS codes)
func (g *DefaultTokenGenerator) GenerateToken() string {
	// Generate 6-digit code
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

// GenerateSecureToken generates a secure random token
func (g *DefaultTokenGenerator) GenerateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}