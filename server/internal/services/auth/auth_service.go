package auth

import (
	"context"
	"fmt"
)

type AuthService interface {
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
}

type authService struct {
	Auth  *kratosService
	AuthZ *ketoService
}

func NewAuthService(kratos *kratosService, keto *ketoService) AuthService {
	return &authService{
		Auth:  kratos,
		AuthZ: keto,
	}
}

func (a *authService) ExtractStudentID(identity *Identity) (int, error) {
	// In JWT-based authentication, the student ID lookup is handled by middleware
	// and stored in Echo context under "student_id"
	// This method cannot access Echo context directly
	// For proper JWT integration, use helpers.ExtractStudentID(c) instead
	return 0, fmt.Errorf("ExtractStudentID from identity not implemented - student ID should be extracted from Echo context after LoadStudentProfile middleware")
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
