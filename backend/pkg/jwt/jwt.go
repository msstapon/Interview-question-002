// Package jwt issues and parses RS256 access tokens.
package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims is the access-token payload. Username is embedded so the /me endpoint can
// render "Welcome User : xxx" straight from the verified token when desired.
type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"username,omitempty"`
	jwtv5.RegisteredClaims
}

type Manager struct {
	priv      *rsa.PrivateKey
	pub       *rsa.PublicKey
	issuer    string
	accessTTL time.Duration
}

func New(privPath, pubPath, issuer string, accessTTL time.Duration) (*Manager, error) {
	privBytes, err := os.ReadFile(privPath)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	priv, err := jwtv5.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, fmt.Errorf("read public key: %w", err)
	}
	pub, err := jwtv5.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}
	return &Manager{priv: priv, pub: pub, issuer: issuer, accessTTL: accessTTL}, nil
}

type IssueResult struct {
	Token     string
	JTI       string
	ExpiresAt time.Time
}

func (m *Manager) Issue(userID, username string) (*IssueResult, error) {
	now := time.Now()
	exp := now.Add(m.accessTTL)
	jti := uuid.NewString()

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID,
			ID:        jti,
			ExpiresAt: jwtv5.NewNumericDate(exp),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}
	tok := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	s, err := tok.SignedString(m.priv)
	if err != nil {
		return nil, err
	}
	return &IssueResult{Token: s, JTI: jti, ExpiresAt: exp}, nil
}

func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwtv5.ParseWithClaims(tokenStr, claims, func(t *jwtv5.Token) (any, error) {
		if _, ok := t.Method.(*jwtv5.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.pub, nil
	}, jwtv5.WithIssuer(m.issuer), jwtv5.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (m *Manager) AccessTTL() time.Duration { return m.accessTTL }
