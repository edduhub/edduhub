package auth

import (
	"context"
)

type AuthService interface {
	InitiateRegistrationFlow(ctx context.Context) (map[string]any, error)
	CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (*Identity, error)
	ValidateSession(ctx context.Context, sessionToken string) (*Identity, error)
	CheckCollegeAccess(identity *Identity, collegeID string) bool
	HasRole(identity *Identity, role string) bool
	CheckPermission(ctx context.Context, identity *Identity, action, resource string) (bool, error)
	AssignRole(ctx context.Context, identityID string, role string) error
	RemoveRole(ctx context.Context, identityID string, role string) error
	AddPermission(ctx context.Context, identityID, action, resource string) error
	RemovePermission(ctx context.Context, identityID, action, resource string) error
	GetPublicURL() string
	ExtractStudentID(ctx context.Context)
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

func (a *authService) InitiateRegistrationFlow(ctx context.Context) (map[string]any, error) {
	return a.Auth.InitiateRegistrationFlow(ctx)
}

func (a *authService) CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (*Identity, error) {
	return a.Auth.CompleteRegistration(ctx, flowID, req)
}

func (a *authService) ValidateSession(ctx context.Context, sessionToken string) (*Identity, error) {
	return a.Auth.ValidateSession(ctx, sessionToken)
}

func (a *authService) CheckCollegeAccess(identity *Identity, collegeID string) bool {
	return a.Auth.CheckCollegeAccess(identity, collegeID)
}

func (a *authService) HasRole(identity *Identity, role string) bool {
	return a.Auth.HasRole(identity, role)
}

func (a *authService) CheckPermission(ctx context.Context, identity *Identity, action, resource string) (bool, error) {
	return a.AuthZ.CheckPermission(ctx, identity.ID, action, resource)
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
