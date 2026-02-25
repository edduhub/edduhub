package auth

import (
	"context"
	"fmt"
	"testing"

	"eduhub/server/internal/models"
	jwtpkg "eduhub/server/pkg/jwt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryUserStore struct {
	byKratos map[string]*models.User
	nextID   int
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{
		byKratos: make(map[string]*models.User),
		nextID:   1,
	}
}

func (m *memoryUserStore) GetUserByKratosID(_ context.Context, kratosID string) (*models.User, error) {
	user, ok := m.byKratos[kratosID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copy := *user
	return &copy, nil
}

func (m *memoryUserStore) CreateUser(_ context.Context, user *models.User) error {
	if user.ID == 0 {
		user.ID = m.nextID
		m.nextID++
	}
	copy := *user
	m.byKratos[user.KratosIdentityID] = &copy
	return nil
}

func (m *memoryUserStore) UpdateUser(_ context.Context, user *models.User) error {
	if _, ok := m.byKratos[user.KratosIdentityID]; !ok {
		return fmt.Errorf("user not found")
	}
	copy := *user
	m.byKratos[user.KratosIdentityID] = &copy
	return nil
}

type memoryProfileStore struct {
	byUserID map[int]*models.Profile
	nextID   int
}

func newMemoryProfileStore() *memoryProfileStore {
	return &memoryProfileStore{
		byUserID: make(map[int]*models.Profile),
		nextID:   1,
	}
}

func (m *memoryProfileStore) GetProfileByUserID(_ context.Context, userID int) (*models.Profile, error) {
	profile, ok := m.byUserID[userID]
	if !ok {
		return nil, fmt.Errorf("profile not found")
	}
	copy := *profile
	return &copy, nil
}

func (m *memoryProfileStore) CreateProfile(_ context.Context, profile *models.Profile) error {
	if profile.ID == 0 {
		profile.ID = m.nextID
		m.nextID++
	}
	copy := *profile
	m.byUserID[profile.UserID] = &copy
	return nil
}

type staticCollegeResolver struct {
	college *models.College
}

func (s *staticCollegeResolver) GetCollegeByExternalID(_ context.Context, externalID string) (*models.College, error) {
	if s.college == nil {
		return nil, fmt.Errorf("college %q not found", externalID)
	}
	return s.college, nil
}

type recordingJWTManager struct {
	claims            *jwtpkg.JWTClaims
	lastGeneratedUser int
}

func (r *recordingJWTManager) Generate(userID int, kratosID, email, role, collegeID, firstName, lastName string) (string, error) {
	r.lastGeneratedUser = userID
	return "token", nil
}

func (r *recordingJWTManager) Verify(token string) (*jwtpkg.JWTClaims, error) {
	if r.claims == nil {
		return nil, fmt.Errorf("invalid token")
	}
	return r.claims, nil
}

func TestEnsureLocalUserCreatesUserWhenMissing(t *testing.T) {
	store := newMemoryUserStore()
	service := &authService{UserStore: store}

	identity := &Identity{ID: "kratos-1"}
	identity.Traits.Email = "new@example.edu"
	identity.Traits.Role = "student"
	identity.Traits.Name.First = "Ada"
	identity.Traits.Name.Last = "Lovelace"

	user, err := service.ensureLocalUser(context.Background(), identity)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Ada Lovelace", user.Name)
	assert.Equal(t, "new@example.edu", user.Email)
	assert.Equal(t, "student", user.Role)
	assert.True(t, user.IsActive)
}

func TestEnsureLocalProfileCreatesProfileWhenMissing(t *testing.T) {
	profiles := newMemoryProfileStore()
	service := &authService{
		ProfileStore: profiles,
		CollegeStore: &staticCollegeResolver{college: &models.College{ID: 42}},
	}

	user := &models.User{ID: 10}
	identity := &Identity{ID: "kratos-10"}
	identity.Traits.Name.First = "Grace"
	identity.Traits.Name.Last = "Hopper"
	identity.Traits.College.ID = "COL-42"

	err := service.ensureLocalProfile(context.Background(), user, identity)
	require.NoError(t, err)

	profile, err := profiles.GetProfileByUserID(context.Background(), 10)
	require.NoError(t, err)
	assert.Equal(t, 42, profile.CollegeID)
	assert.Equal(t, "Grace", profile.FirstName)
	assert.Equal(t, "Hopper", profile.LastName)
}

func TestValidateJWTResolvesMissingUserIDFromStore(t *testing.T) {
	store := newMemoryUserStore()
	require.NoError(t, store.CreateUser(context.Background(), &models.User{
		KratosIdentityID: "kratos-55",
		Name:             "User",
		Email:            "u@example.edu",
		Role:             "student",
		IsActive:         true,
	}))

	service := &authService{
		JWTManager: &recordingJWTManager{
			claims: &jwtpkg.JWTClaims{
				UserID:    0,
				KratosID:  "kratos-55",
				Email:     "u@example.edu",
				Role:      "student",
				CollegeID: "1",
				FirstName: "U",
				LastName:  "Ser",
			},
		},
		UserStore: store,
	}

	identity, err := service.ValidateJWT(context.Background(), "token")
	require.NoError(t, err)
	assert.Equal(t, 1, identity.UserID)
	assert.Equal(t, "kratos-55", identity.ID)
}

func TestRefreshTokenUsesResolvedUserID(t *testing.T) {
	store := newMemoryUserStore()
	require.NoError(t, store.CreateUser(context.Background(), &models.User{
		KratosIdentityID: "kratos-77",
		Name:             "User",
		Email:            "u@example.edu",
		Role:             "admin",
		IsActive:         true,
	}))

	jwtMgr := &recordingJWTManager{
		claims: &jwtpkg.JWTClaims{
			UserID:    0,
			KratosID:  "kratos-77",
			Email:     "u@example.edu",
			Role:      "admin",
			CollegeID: "3",
			FirstName: "Admin",
			LastName:  "User",
		},
	}

	service := &authService{
		JWTManager: jwtMgr,
		UserStore:  store,
	}

	token, err := service.RefreshToken(context.Background(), "old-token")
	require.NoError(t, err)
	assert.Equal(t, "token", token)
	assert.Equal(t, 1, jwtMgr.lastGeneratedUser)
}
