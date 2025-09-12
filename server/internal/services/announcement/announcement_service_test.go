package announcement

import (
	"context"
	"eduhub/server/internal/models"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnnouncementRepository is a mock implementation of the AnnouncementRepository
type MockAnnouncementRepository struct {
	mock.Mock
}

func (m *MockAnnouncementRepository) Create(ctx context.Context, announcement *models.Announcement) (int, error) {
	args := m.Called(ctx, announcement)
	return args.Int(0), args.Error(1)
}

func (m *MockAnnouncementRepository) GetByID(ctx context.Context, id int) (*models.Announcement, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementRepository) GetByCollegeID(ctx context.Context, collegeID int, limit, offset int) ([]*models.Announcement, error) {
	args := m.Called(ctx, collegeID, limit, offset)
	return args.Get(0).([]*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementRepository) Update(ctx context.Context, announcement *models.Announcement) error {
	args := m.Called(ctx, announcement)
	return args.Error(0)
}

func (m *MockAnnouncementRepository) DeleteByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCollegeRepository is a mock implementation of the CollegeRepository
type MockCollegeRepository struct {
	mock.Mock
}

func (m *MockCollegeRepository) GetCollegeByID(ctx context.Context, id int) (*models.College, error) {
	args := m.Called(ctx, id)
	// You might need to adjust the return based on the actual signature
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.College), args.Error(1)
}

// MockUserRepository is a mock implementation of the UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindUserByID(ctx context.Context, id int) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func (m *MockUserRepository) FindUserByKratosID(ctx context.Context, kratosID string) (*models.User, error) {
	args := m.Called(ctx, kratosID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
func TestCreateAnnouncement_Success(t *testing.T) {
	// Arrange
	mockAnnouncementRepo := new(MockAnnouncementRepository)
	mockCollegeRepo := new(MockCollegeRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewService(mockAnnouncementRepo, mockCollegeRepo, mockUserRepo)

	req := &models.CreateAnnouncementRequest{
		Title:   "Valid Title",
		Content: "This is some valid content for the announcement.",
	}
	collegeID := 1
	userID := 1
	expectedID := 100

	// Mock expectations
	mockCollegeRepo.On("GetCollegeByID", context.Background(), collegeID).Return(&models.College{ID: collegeID}, nil)
	mockUserRepo.On("FindUserByID", context.Background(), userID).Return(&models.User{ID: userID}, nil)
	mockAnnouncementRepo.On("Create", context.Background(), mock.AnythingOfType("*models.Announcement")).Return(expectedID, nil)

	// Act
	announcement, err := service.CreateAnnouncement(context.Background(), req, collegeID, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, announcement)
	assert.Equal(t, expectedID, announcement.ID)
	assert.Equal(t, req.Title, announcement.Title)
	mockAnnouncementRepo.AssertExpectations(t)
	mockCollegeRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestCreateAnnouncement_ValidationFailure(t *testing.T) {
	// Arrange
	mockAnnouncementRepo := new(MockAnnouncementRepository)
	mockCollegeRepo := new(MockCollegeRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewService(mockAnnouncementRepo, mockCollegeRepo, mockUserRepo)

	req := &models.CreateAnnouncementRequest{
		Title:   "No", // Invalid title
		Content: "short", // Invalid content
	}

	// Act
	_, err := service.CreateAnnouncement(context.Background(), req, 1, 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestCreateAnnouncement_CollegeNotFound(t *testing.T) {
	// Arrange
	mockAnnouncementRepo := new(MockAnnouncementRepository)
	mockCollegeRepo := new(MockCollegeRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewService(mockAnnouncementRepo, mockCollegeRepo, mockUserRepo)

	req := &models.CreateAnnouncementRequest{
		Title:   "Valid Title",
		Content: "This is some valid content.",
	}
	collegeID := 999 // Non-existent college

	// Mock expectations
	mockCollegeRepo.On("GetCollegeByID", context.Background(), collegeID).Return(nil, errors.New("not found"))

	// Act
	_, err := service.CreateAnnouncement(context.Background(), req, collegeID, 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify college")
	mockCollegeRepo.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "FindUserByID")
	mockAnnouncementRepo.AssertNotCalled(t, "Create")
}

func TestCreateAnnouncement_UserNotFound(t *testing.T) {
	// Arrange
	mockAnnouncementRepo := new(MockAnnouncementRepository)
	mockCollegeRepo := new(MockCollegeRepository)
	mockUserRepo := new(MockUserRepository)
	service := NewService(mockAnnouncementRepo, mockCollegeRepo, mockUserRepo)
	req := &models.CreateAnnouncementRequest{
		Title:   "Valid Title",
		Content: "This is some valid content.",
	}
	collegeID := 1
	userID := 999 // Non-existent user

	// Mock expectations
	mockCollegeRepo.On("GetCollegeByID", context.Background(), collegeID).Return(&models.College{ID: collegeID}, nil)
	mockUserRepo.On("FindUserByID", context.Background(), userID).Return(nil, errors.New("not found"))

	// Act
	_, err := service.CreateAnnouncement(context.Background(), req, collegeID, userID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify user")
	mockCollegeRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockAnnouncementRepo.AssertNotCalled(t, "Create")
}
// Add dummy implementations for the methods that are not used in these tests
func (m *MockCollegeRepository) CreateCollege(ctx context.Context, college *models.College) (int, error) {
	return 0, nil
}
func (m *MockCollegeRepository) UpdateCollege(ctx context.Context, college *models.College) error {
	return nil
}
func (m *MockCollegeRepository) DeleteCollege(ctx context.Context, id int) error {
	return nil
}
func (m *MockCollegeRepository) GetAllColleges(ctx context.Context) ([]*models.College, error) {
	return nil, nil
}
func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return nil
}
func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return nil
}
func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	return nil
}
func (m *MockUserRepository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}
func (m *MockUserRepository) GetUserRole(ctx context.Context, userID int) (string, error) {
	return "", nil
}
func (m *MockUserRepository) SetUserRole(ctx context.Context, userID int, role string) error {
	return nil
}
func (m *MockUserRepository) GetUserStatus(ctx context.Context, userID int) (bool, error) {
	return false, nil
}
func (m *MockUserRepository) SetUserStatus(ctx context.Context, userID int, isActive bool) error {
	return nil
}
