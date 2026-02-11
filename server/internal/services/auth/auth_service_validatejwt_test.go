package auth

import (
	"context"
	"errors"
	"testing"

	jwtpkg "eduhub/server/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubJWTManager struct {
	claims    *jwtpkg.JWTClaims
	verifyErr error
}

func (s *stubJWTManager) Generate(kratosID, email, role, collegeID, firstName, lastName string) (string, error) {
	return "", nil
}

func (s *stubJWTManager) Verify(token string) (*jwtpkg.JWTClaims, error) {
	if s.verifyErr != nil {
		return nil, s.verifyErr
	}
	return s.claims, nil
}

func TestAuthServiceValidateJWTBuildsIdentityFromClaims(t *testing.T) {
	manager := &stubJWTManager{
		claims: &jwtpkg.JWTClaims{
			KratosID:  "kratos-123",
			Email:     "student@example.edu",
			Role:      "student",
			CollegeID: "42",
			FirstName: "Ada",
			LastName:  "Lovelace",
		},
	}

	service := &authService{
		JWTManager: manager,
	}

	identity, err := service.ValidateJWT(context.Background(), "valid-token")
	require.NoError(t, err)
	require.NotNil(t, identity)

	assert.Equal(t, "kratos-123", identity.ID)
	assert.Equal(t, "student@example.edu", identity.Traits.Email)
	assert.Equal(t, "student", identity.Traits.Role)
	assert.Equal(t, "42", identity.Traits.College.ID)
	assert.Equal(t, "Ada", identity.Traits.Name.First)
	assert.Equal(t, "Lovelace", identity.Traits.Name.Last)
}

func TestAuthServiceValidateJWTReturnsErrorWhenVerificationFails(t *testing.T) {
	manager := &stubJWTManager{
		verifyErr: errors.New("signature mismatch"),
	}

	service := &authService{
		JWTManager: manager,
	}

	identity, err := service.ValidateJWT(context.Background(), "invalid-token")
	require.Error(t, err)
	assert.Nil(t, identity)
	assert.Contains(t, err.Error(), "invalid JWT token")
}
