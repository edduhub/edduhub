package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrWeakSecret   = errors.New("secret key must be at least 32 characters")
	ErrTokenRevoked = errors.New("token has been revoked")
)

// Allowed signing methods - only HS256 for security
var allowedSigningMethods = map[string]bool{
	"HS256": true,
}

type Claims struct {
	UserID    int    `json:"user_id,omitempty"`
	KratosID  string `json:"sub"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CollegeID string `json:"college_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	jwt.RegisteredClaims
}

// Convert Claims to common JWTClaims format for auth service
func (c *Claims) ToJWTClaims() *JWTClaims {
	return &JWTClaims{
		UserID:    c.UserID,
		KratosID:  c.KratosID,
		Email:     c.Email,
		Role:      c.Role,
		CollegeID: c.CollegeID,
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}
}

// JWTClaims defines the expected JWT claims structure for auth service
type JWTClaims struct {
	UserID    int
	KratosID  string
	Email     string
	Role      string
	CollegeID string
	FirstName string
	LastName  string
}

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) (*JWTManager, error) {
	if len(secretKey) < 32 {
		return nil, ErrWeakSecret
	}
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}, nil
}

func (m *JWTManager) Generate(userID int, kratosID, email, role, collegeID, firstName, lastName string) (string, error) {
	// Generate unique JWT ID for revocation support
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", errors.New("failed to generate token ID")
	}
	jti := hex.EncodeToString(jtiBytes)

	claims := Claims{
		UserID:    userID,
		KratosID:  kratosID,
		Email:     email,
		Role:      role,
		CollegeID: collegeID,
		FirstName: firstName,
		LastName:  lastName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "eduhub",
			Audience:  jwt.ClaimStrings{"eduhub"},
			ID:        jti, // JWT ID for revocation support
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Verify(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			// Explicit algorithm validation - only allow HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			// Verify signing method is HS256
			if token.Method.Alg() != "HS256" {
				return nil, ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, ErrExpiredToken
	}

	return claims.ToJWTClaims(), nil
}
