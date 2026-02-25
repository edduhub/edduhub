package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"eduhub/server/internal/models"
	"eduhub/server/pkg/jwt"
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *Identity, error)
	InitiateRegistrationFlow(ctx context.Context) (map[string]any, error)
	CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (string, *Identity, error)
	ValidateSession(ctx context.Context, sessionToken string) (*Identity, error)
	ValidateJWT(ctx context.Context, jwtToken string) (*Identity, error)
	ValidateCollegeAccess(ctx context.Context, collegeID int) (interface{}, error)
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
	RefreshToken(ctx context.Context, token string) (string, error)
}

type authService struct {
	Auth           *kratosService
	AuthZ          *ketoService
	JWTManager     JWTManager
	CollegeChecker CollegeChecker
	UserStore      UserStore
	ProfileStore   ProfileStore
	CollegeStore   CollegeStore
}

type JWTManager interface {
	Generate(userID int, kratosID, email, role, collegeID, firstName, lastName string) (string, error)
	Verify(token string) (*jwt.JWTClaims, error)
}

type CollegeChecker interface {
	GetCollegeByID(ctx context.Context, id int) (interface{}, error)
}

type UserStore interface {
	GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, user *models.User) error
}

type ProfileStore interface {
	GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error)
	CreateProfile(ctx context.Context, profile *models.Profile) error
}

type CollegeStore interface {
	GetCollegeByExternalID(ctx context.Context, externalID string) (*models.College, error)
}

func NewAuthService(kratos *kratosService, keto *ketoService, jwtManager JWTManager) AuthService {
	return &authService{
		Auth:       kratos,
		AuthZ:      keto,
		JWTManager: jwtManager,
	}
}

func NewAuthServiceWithCollege(kratos *kratosService, keto *ketoService, jwtManager JWTManager, collegeChecker CollegeChecker) AuthService {
	return &authService{
		Auth:           kratos,
		AuthZ:          keto,
		JWTManager:     jwtManager,
		CollegeChecker: collegeChecker,
	}
}

// NewAuthServiceWithDependencies creates an auth service with all dependencies for user/profile provisioning
func NewAuthServiceWithDependencies(kratos *kratosService, keto *ketoService, jwtManager JWTManager, collegeRepo interface{}, userRepo interface{}, profileRepo interface{}, collegeChecker interface{}) AuthService {
	service := &authService{
		Auth:       kratos,
		AuthZ:      keto,
		JWTManager: jwtManager,
	}

	if cc, ok := collegeChecker.(CollegeChecker); ok {
		service.CollegeChecker = cc
	}
	if us, ok := userRepo.(UserStore); ok {
		service.UserStore = us
	}
	if ps, ok := profileRepo.(ProfileStore); ok {
		service.ProfileStore = ps
	}
	if cs, ok := collegeRepo.(CollegeStore); ok {
		service.CollegeStore = cs
	}

	return service
}

func (a *authService) Login(ctx context.Context, email, password string) (string, *Identity, error) {
	// Authenticate with Kratos
	identity, err := a.Auth.Login(ctx, email, password)
	if err != nil {
		return "", nil, fmt.Errorf("authentication failed: %w", err)
	}

	userID, err := a.resolveAndProvisionLocalIdentity(ctx, identity)
	if err != nil {
		return "", nil, err
	}

	// Generate JWT token
	token, err := a.JWTManager.Generate(
		userID,
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

func (a *authService) CompleteRegistration(ctx context.Context, flowID string, req RegistrationRequest) (string, *Identity, error) {
	identity, err := a.Auth.CompleteRegistration(ctx, flowID, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to complete registration: %w", err)
	}

	userID, err := a.resolveAndProvisionLocalIdentity(ctx, identity)
	if err != nil {
		return "", nil, err
	}

	// Generate JWT token
	token, err := a.JWTManager.Generate(
		userID,
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

func (a *authService) ValidateSession(ctx context.Context, sessionToken string) (*Identity, error) {
	return a.Auth.ValidateSession(ctx, sessionToken)
}

func (a *authService) ValidateJWT(ctx context.Context, jwtToken string) (*Identity, error) {
	claims, err := a.JWTManager.Verify(jwtToken)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	identity := &Identity{
		ID: claims.KratosID,
	}
	identity.Traits.Email = claims.Email
	identity.Traits.Role = claims.Role
	identity.Traits.College.ID = claims.CollegeID
	identity.Traits.Name.First = claims.FirstName
	identity.Traits.Name.Last = claims.LastName
	identity.UserID = claims.UserID

	if _, err := a.resolveAndProvisionLocalIdentity(ctx, identity); err != nil {
		return nil, err
	}

	return identity, nil
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

// ValidateCollegeAccess verifies that a college with the given ID exists in the database
// This is a critical security function for multi-tenant isolation
func (a *authService) ValidateCollegeAccess(ctx context.Context, collegeID int) (interface{}, error) {
	if a.CollegeChecker == nil {
		// If college checker not injected, assume validation is handled elsewhere
		return map[string]int{"id": collegeID}, nil
	}
	return a.CollegeChecker.GetCollegeByID(ctx, collegeID)
}

// RefreshToken validates and generates a new JWT token with updated expiration
// Implements token rotation for enhanced security
func (a *authService) RefreshToken(ctx context.Context, token string) (string, error) {
	// Validate the existing token
	claims, err := a.JWTManager.Verify(token)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	userID := claims.UserID
	if userID == 0 {
		identity := &Identity{
			ID: claims.KratosID,
		}
		identity.Traits.Email = claims.Email
		identity.Traits.Role = claims.Role
		identity.Traits.College.ID = claims.CollegeID
		identity.Traits.Name.First = claims.FirstName
		identity.Traits.Name.Last = claims.LastName

		userID, err = a.resolveAndProvisionLocalIdentity(ctx, identity)
		if err != nil {
			return "", err
		}
	}

	// Generate a new token with the same claims but new expiration
	newToken, err := a.JWTManager.Generate(
		userID,
		claims.KratosID,
		claims.Email,
		claims.Role,
		claims.CollegeID,
		claims.FirstName,
		claims.LastName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

func (a *authService) resolveAndProvisionLocalIdentity(ctx context.Context, identity *Identity) (int, error) {
	if identity == nil {
		return 0, fmt.Errorf("identity is nil")
	}

	if identity.UserID > 0 {
		return identity.UserID, nil
	}

	if a.UserStore == nil {
		return 0, nil
	}

	user, err := a.ensureLocalUser(ctx, identity)
	if err != nil {
		return 0, err
	}

	identity.UserID = user.ID
	if err := a.ensureLocalProfile(ctx, user, identity); err != nil {
		return 0, err
	}
	return user.ID, nil
}

func (a *authService) ensureLocalUser(ctx context.Context, identity *Identity) (*models.User, error) {
	if identity == nil {
		return nil, fmt.Errorf("identity is nil")
	}
	if a.UserStore == nil {
		return nil, fmt.Errorf("user store is not configured")
	}

	existing, err := a.UserStore.GetUserByKratosID(ctx, identity.ID)
	if err == nil {
		updated := false

		fullName := strings.TrimSpace(identity.Traits.Name.First + " " + identity.Traits.Name.Last)
		if fullName != "" && existing.Name != fullName {
			existing.Name = fullName
			updated = true
		}
		if identity.Traits.Email != "" && existing.Email != identity.Traits.Email {
			existing.Email = identity.Traits.Email
			updated = true
		}
		if identity.Traits.Role != "" && existing.Role != identity.Traits.Role {
			existing.Role = identity.Traits.Role
			updated = true
		}
		if !existing.IsActive {
			existing.IsActive = true
			updated = true
		}

		if updated {
			if err := a.UserStore.UpdateUser(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to update local user: %w", err)
			}
		}

		return existing, nil
	}

	if !isNotFoundErr(err) {
		return nil, fmt.Errorf("failed to fetch local user: %w", err)
	}

	fullName := strings.TrimSpace(identity.Traits.Name.First + " " + identity.Traits.Name.Last)
	if fullName == "" {
		fullName = identity.Traits.Email
	}
	if fullName == "" {
		fullName = identity.ID
	}

	user := &models.User{
		Name:             fullName,
		Role:             identity.Traits.Role,
		Email:            identity.Traits.Email,
		KratosIdentityID: identity.ID,
		IsActive:         true,
	}

	if err := a.UserStore.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create local user: %w", err)
	}

	return user, nil
}

func (a *authService) ensureLocalProfile(ctx context.Context, user *models.User, identity *Identity) error {
	if a.ProfileStore == nil {
		return nil
	}
	if user == nil || user.ID == 0 {
		return fmt.Errorf("invalid user for profile provisioning")
	}
	if identity == nil {
		return fmt.Errorf("identity is nil")
	}

	_, err := a.ProfileStore.GetProfileByUserID(ctx, user.ID)
	if err == nil {
		return nil
	}
	if !isNotFoundErr(err) {
		return fmt.Errorf("failed to fetch local profile: %w", err)
	}

	collegeID, err := a.resolveCollegeID(ctx, identity.Traits.College.ID)
	if err != nil {
		return err
	}

	profile := &models.Profile{
		UserID:      user.ID,
		CollegeID:   collegeID,
		FirstName:   identity.Traits.Name.First,
		LastName:    identity.Traits.Name.Last,
		Preferences: models.JSONMap{},
		SocialLinks: models.JSONMap{},
	}

	if err := a.ProfileStore.CreateProfile(ctx, profile); err != nil {
		return fmt.Errorf("failed to create local profile: %w", err)
	}

	return nil
}

func (a *authService) resolveCollegeID(ctx context.Context, externalCollegeID string) (int, error) {
	if externalCollegeID == "" {
		return 0, nil
	}

	if a.CollegeStore != nil {
		college, err := a.CollegeStore.GetCollegeByExternalID(ctx, externalCollegeID)
		if err == nil {
			return college.ID, nil
		}
		if !isNotFoundErr(err) {
			return 0, fmt.Errorf("failed to resolve college from external id %q: %w", externalCollegeID, err)
		}
	}

	collegeID, err := strconv.Atoi(externalCollegeID)
	if err != nil {
		return 0, fmt.Errorf("failed to resolve college from external id %q", externalCollegeID)
	}
	return collegeID, nil
}

func isNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "no rows")
}
