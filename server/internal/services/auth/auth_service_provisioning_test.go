package auth

import (
	"context"
	"fmt"
	"testing"

	"eduhub/server/internal/models"

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

type memoryStudentStore struct {
	byKratos map[string]*models.Student
	nextID   int
}

func newMemoryStudentStore() *memoryStudentStore {
	return &memoryStudentStore{
		byKratos: make(map[string]*models.Student),
		nextID:   1,
	}
}

func (m *memoryStudentStore) FindByKratosID(_ context.Context, kratosID string) (*models.Student, error) {
	student, ok := m.byKratos[kratosID]
	if !ok {
		return nil, fmt.Errorf("student not found")
	}
	copy := *student
	return &copy, nil
}

func (m *memoryStudentStore) CreateStudent(_ context.Context, student *models.Student) error {
	if student.StudentID == 0 {
		student.StudentID = m.nextID
		m.nextID++
	}
	copy := *student
	m.byKratos[student.KratosIdentityID] = &copy
	return nil
}

func (m *memoryStudentStore) UpdateStudent(_ context.Context, student *models.Student) error {
	if _, ok := m.byKratos[student.KratosIdentityID]; !ok {
		return fmt.Errorf("student not found")
	}
	copy := *student
	m.byKratos[student.KratosIdentityID] = &copy
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

func TestResolveAndProvisionLocalIdentityCreatesStudentRecord(t *testing.T) {
	users := newMemoryUserStore()
	profiles := newMemoryProfileStore()
	students := newMemoryStudentStore()
	service := &authService{
		UserStore:    users,
		ProfileStore: profiles,
		CollegeStore: &staticCollegeResolver{college: &models.College{ID: 42}},
		StudentStore: students,
	}

	identity := &Identity{ID: "kratos-student-1"}
	identity.Traits.Email = "student@example.edu"
	identity.Traits.Role = "student"
	identity.Traits.RollNo = "CS001"
	identity.Traits.Name.First = "Ada"
	identity.Traits.Name.Last = "Student"
	identity.Traits.College.ID = "COL-42"

	userID, err := service.resolveAndProvisionLocalIdentity(context.Background(), identity)
	require.NoError(t, err)
	assert.Equal(t, 1, userID)

	student, err := students.FindByKratosID(context.Background(), "kratos-student-1")
	require.NoError(t, err)
	assert.Equal(t, 1, student.UserID)
	assert.Equal(t, 42, student.CollegeID)
	assert.Equal(t, "CS001", student.RollNo)
	assert.True(t, student.IsActive)
}
