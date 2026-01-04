package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID    string `json:"-"`
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

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
	}
}

func (m *JWTManager) Generate(kratosID, email, role, collegeID, firstName, lastName string) (string, error) {
	claims := Claims{
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
			Issuer:    "eduhub",                   // Identify token issuer
			Audience:  jwt.ClaimStrings{"eduhub"}, // Intended audience
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) Verify(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
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
