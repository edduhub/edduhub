package auth

import (
	"context"
	"fmt"

	"eduhub/server/pkg/jwt"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *Identity, error)
	InitiateRegistrationFlow(ctx context.Context) (map[string]any, error)
	CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (*Identity, error)
	ValidateSession(ctx context.Context, sessionToken string) (*Identity, error)
	ValidateJWT(ctx context.Context, jwtToken string) (*Identity, error)
	CheckCollegeAccess(identity *Identity, collegeID string) bool
	HasRole(identity *Identity, role string) bool
	CheckPermission(ctx context.Context, identity *Identity, action, resource string) (bool, error)
	AssignRole(ctx context.Context, identityID string, role string) error
	RemoveRole(ctx context.Context, identityID string, role string) error
	AddPermission(ctx context.Context, identityID, action, resource string) error
	RemovePermission(ctx context.Context, identityID, action, resource string) error
	GetPublicURL() string
	ExtractStudentID(identity *Identity) (int, error)
	Logout(ctx context.Context, sessionToken string) error
	RefreshSession(ctx context.Context, sessionToken string) (string, error)
	InitiatePasswordReset(ctx context.Context, email string) error
	CompletePasswordReset(ctx context.Context, flowID string, newPassword string) error
	VerifyEmail(ctx context.Context, flowID string, token string) error
	InitiateEmailVerification(ctx context.Context, identityID string) (map[string]any, error)
	ChangePassword(ctx context.Context, identityID string, oldPassword string, newPassword string) error
}

type authService struct {
	Auth       *kratosService
	AuthZ      *ketoService
	JWTManager JWTManager
}

type JWTManager interface {
	Generate(kratosID, email, role, collegeID, firstName, lastName string) (string, error)
	Verify(token string) (*jwt.JWTClaims, error)
}

func NewAuthService(kratos *kratosService, keto *ketoService, jwtManager JWTManager) AuthService {
	return &authService{
		Auth:       kratos,
		AuthZ:      keto,
		JWTManager: jwtManager,
	}
}

func (a *authService) Login(ctx context.Context, email, password string) (string, *Identity, error) {
	// Authenticate with Kratos
	identity, err := a.Auth.Login(ctx, email, password)
	if err != nil {
		return "", nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Generate JWT token
	token, err := a.JWTManager.Generate(
		identity.ID,
		identity.Traits.Email,
		identity.Traits.Role,
		identity.Traits.College.ID,
		identity.Traits.Name.First,
		identity.Traits.Name.Last,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, identity, nil
}

func (a *authService) ExtractStudentID(identity *Identity) (int, error) {
	// In JWT-based authentication, the student ID can be extracted directly from JWT claims
	// without needing middleware context. However, we need to implement the logic to
	// find the student by kratos ID.

	// We cannot access services from this layer, so this method signature should be
	// changed to accept the student service, or the logic should be moved to middleware.
	// For now, return error indicating this should be handled by middleware.
	return 0, fmt.Errorf("ExtractStudentID requires service dependencies - use helpers.ExtractStudentID(c) with LoadStudentProfile middleware instead")
}

func (a *authService) InitiateRegistrationFlow(ctx context.Context) (map[string]any, error) {
	return a.Auth.InitiateRegistrationFlow(ctx)
}

func (a *authService) CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (*Identity, error) {
	return a.Auth.CompleteRegistration(ctx, flowID, req)
}

func (a *authService) ValidateSession(ctx context.Context, sessionToken string) (*Identity, error) {
	return a.Auth.ValidateSession(ctx, sessionToken)
}

func (a *authService) ValidateJWT(ctx context.Context, jwtToken string) (*Identity, error) {
	return a.Auth.ValidateJWT(ctx, jwtToken)
}

func (a *authService) CheckCollegeAccess(identity *Identity, collegeID string) bool {
	return a.Auth.CheckCollegeAccess(identity, collegeID)
}

func (a *authService) HasRole(identity *Identity, role string) bool {
	return a.Auth.HasRole(identity, role)
}

func (a *authService) CheckPermission(ctx context.Context, identity *Identity, action, resource string) (bool, error) {
	return a.AuthZ.CheckPermission(ctx, "app", identity.ID, action, resource)
}

func (a *authService) AssignRole(ctx context.Context, identityID string, role string) error {
	return a.AuthZ.CreateRelation(ctx, "app", "role:"+role, "member", identityID)
}

func (a *authService) RemoveRole(ctx context.Context, identityID string, role string) error {
	return a.AuthZ.DeleteRelation(ctx, "app", "role:"+role, "member", identityID)
}

func (a *authService) AddPermission(ctx context.Context, identityID, action, resource string) error {
	return a.AuthZ.CreateRelation(ctx, "app", resource, action, identityID)
}

func (a *authService) RemovePermission(ctx context.Context, identityID, action, resource string) error {
	return a.AuthZ.DeleteRelation(ctx, "app", resource, action, identityID)
}

func (a *authService) GetPublicURL() string {
	return a.Auth.GetPublicURL()
}

func (a *authService) Logout(ctx context.Context, sessionToken string) error {
	return a.Auth.Logout(ctx, sessionToken)
}

func (a *authService) RefreshSession(ctx context.Context, sessionToken string) (string, error) {
	return a.Auth.RefreshSession(ctx, sessionToken)
}

func (a *authService) InitiatePasswordReset(ctx context.Context, email string) error {
	return a.Auth.InitiatePasswordReset(ctx, email)
}

func (a *authService) CompletePasswordReset(ctx context.Context, flowID string, newPassword string) error {
	return a.Auth.CompletePasswordReset(ctx, flowID, newPassword)
}

func (a *authService) VerifyEmail(ctx context.Context, flowID string, token string) error {
	return a.Auth.VerifyEmail(ctx, flowID, token)
}

func (a *authService) InitiateEmailVerification(ctx context.Context, identityID string) (map[string]any, error) {
	return a.Auth.InitiateEmailVerification(ctx, identityID)
}

func (a *authService) ChangePassword(ctx context.Context, identityID string, oldPassword string, newPassword string) error {
	return a.Auth.ChangePassword(ctx, identityID, oldPassword, newPassword)
}
