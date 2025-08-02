package authservice

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	
	authinterface "{{.Project.GoModule}}/internal/auth/interface"
)

type DefaultTokenGenerator struct{}

func NewTokenGenerator() authinterface.TokenGenerator {
	return &DefaultTokenGenerator{}
}

func (g *DefaultTokenGenerator) GenerateToken() string {
	max := big.NewInt(1000000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", n.Int64())
}

func (g *DefaultTokenGenerator) GenerateSecureToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}