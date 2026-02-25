package profile

import (
	"context"
	"errors"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProfileRepository is a mock implementation of ProfileRepository
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) CreateProfile(ctx context.Context, profile *models.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) GetProfileByUserID(ctx context.Context, userID int) (*models.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetProfileByKratosID(ctx context.Context, kratosID string) (*models.Profile, error) {
	args := m.Called(ctx, kratosID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *MockProfileRepository) GetProfileByID(ctx context.Context, profileID int) (*models.Profile, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Profile), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, profile *models.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) UpdateProfilePartial(ctx context.Context, profileID int, req *models.UpdateProfileRequest) error {
	args := m.Called(ctx, profileID, req)
	return args.Error(0)
}

func (m *MockProfileRepository) DeleteProfile(ctx context.Context, profile *models.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) CreateProfileHistory(ctx context.Context, history *models.ProfileHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockProfileRepository) GetProfileHistory(ctx context.Context, profileID int, limit, offset int) ([]*models.ProfileHistory, error) {
	args := m.Called(ctx, profileID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ProfileHistory), args.Error(1)
}

func TestNewProfileService(t *testing.T) {
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &profileService{}, service)
}

func TestGetProfileByUserID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		expectedProfile := &models.Profile{
			ID:          1,
			UserID:      123,
			CollegeID:   1,
			Bio:         "Test bio",
			PhoneNumber: "1234567890",
			Address:     "Test address",
			DateOfBirth: &now,
			JoinedAt:    time.Now(),
			LastActive:  time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetProfileByUserID", ctx, 123).Return(expectedProfile, nil)

		result, err := service.GetProfileByUserID(ctx, 123)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		mockRepo.On("GetProfileByUserID", ctx, 999).Return(nil, errors.New("profile not found"))

		result, err := service.GetProfileByUserID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProfileByKratosID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	expectedProfile := &models.Profile{
		ID:        1,
		UserID:    123,
		CollegeID: 1,
	}

	mockRepo.On("GetProfileByKratosID", ctx, "kratos-123").Return(expectedProfile, nil)

	result, err := service.GetProfileByKratosID(ctx, "kratos-123")

	assert.NoError(t, err)
	assert.Equal(t, expectedProfile, result)
	mockRepo.AssertExpectations(t)
}

func TestGetProfileByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success", func(t *testing.T) {
		expectedProfile := &models.Profile{
			ID:        1,
			UserID:    123,
			CollegeID: 1,
			Bio:       "Test bio",
		}

		mockRepo.On("GetProfileByID", ctx, 1).Return(expectedProfile, nil)

		result, err := service.GetProfileByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.On("GetProfileByID", ctx, 999).Return(nil, errors.New("profile not found"))

		result, err := service.GetProfileByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success", func(t *testing.T) {
		existingProfile := &models.Profile{
			ID:        1,
			UserID:    123,
			CollegeID: 1,
			Bio:       "Old bio",
		}

		bio := "New bio"
		req := &models.UpdateProfileRequest{
			Bio: &bio,
		}

		mockRepo.On("GetProfileByUserID", ctx, 123).Return(existingProfile, nil)
		mockRepo.On("UpdateProfilePartial", ctx, 1, req).Return(nil)

		err := service.UpdateProfile(ctx, 123, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		req := &models.UpdateProfileRequest{
			Bio: func() *string { s := "test"; return &s }(),
		}

		mockRepo.On("GetProfileByUserID", ctx, 999).Return(nil, errors.New("not found"))

		err := service.UpdateProfile(ctx, 999, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("with multiple fields", func(t *testing.T) {
		existingProfile := &models.Profile{
			ID:        1,
			UserID:    123,
			CollegeID: 1,
		}

		bio := "Updated bio"
		phone := "9876543210"
		address := "New address"

		req := &models.UpdateProfileRequest{
			Bio:         &bio,
			PhoneNumber: &phone,
			Address:     &address,
		}

		mockRepo.On("GetProfileByUserID", ctx, 123).Return(existingProfile, nil)
		mockRepo.On("UpdateProfilePartial", ctx, 1, req).Return(nil)

		err := service.UpdateProfile(ctx, 123, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestCreateProfile(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success", func(t *testing.T) {
		profile := &models.Profile{
			UserID:    123,
			CollegeID: 1,
			Bio:       "Test bio",
		}

		mockRepo.On("CreateProfile", ctx, profile).Return(nil)

		err := service.CreateProfile(ctx, profile)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		profile := &models.Profile{
			UserID:    123,
			CollegeID: 1,
		}

		mockRepo.On("CreateProfile", ctx, profile).Return(errors.New("database error"))

		err := service.CreateProfile(ctx, profile)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetProfileHistory(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success with history", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		bioField := "bio"
		oldVal := "old"
		newVal := "new"
		expectedHistory := []*models.ProfileHistory{
			{
				ID:            1,
				ProfileID:     1,
				UserID:        123,
				ChangedFields: models.JSONMap{"field": "bio"},
				OldValues:     &models.JSONMap{"bio": oldVal},
				NewValues:     &models.JSONMap{"bio": newVal},
				ChangedAt:     time.Now(),
			},
			{
				ID:            2,
				ProfileID:     1,
				UserID:        123,
				ChangedFields: models.JSONMap{"field": "phone"},
				OldValues:     &models.JSONMap{"phone": oldVal},
				NewValues:     &models.JSONMap{"phone": newVal},
				ChangedAt:     time.Now(),
			},
		}

		mockRepo.On("GetProfileHistory", ctx, 1, 10, 0).Return(expectedHistory, nil)

		result, err := service.GetProfileHistory(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, bioField, result[0].ChangedFields["field"])
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty history", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetProfileHistory", ctx, 1, 10, 0).Return([]*models.ProfileHistory{}, nil)

		result, err := service.GetProfileHistory(ctx, 1, 10, 0)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockRepo.ExpectedCalls = nil
		mockRepo.On("GetProfileHistory", ctx, 1, 10, 0).Return(nil, errors.New("database error"))

		result, err := service.GetProfileHistory(ctx, 1, 10, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestCreateProfileHistory(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("success", func(t *testing.T) {
		oldVal := "old"
		newVal := "new"
		history := &models.ProfileHistory{
			ProfileID:     1,
			UserID:        123,
			ChangedFields: models.JSONMap{"field": "bio", "action": "UPDATE"},
			OldValues:     &models.JSONMap{"bio": oldVal},
			NewValues:     &models.JSONMap{"bio": newVal},
			ChangedAt:     time.Now(),
		}

		mockRepo.On("CreateProfileHistory", ctx, history).Return(nil)

		err := service.CreateProfileHistory(ctx, history)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		history := &models.ProfileHistory{
			ProfileID:     1,
			UserID:        123,
			ChangedFields: models.JSONMap{"field": "bio"},
		}

		mockRepo.On("CreateProfileHistory", ctx, history).Return(errors.New("database error"))

		err := service.CreateProfileHistory(ctx, history)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

// Test edge cases
func TestGetProfileByUserID_EdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("zero user ID", func(t *testing.T) {
		expectedProfile := &models.Profile{
			ID:        1,
			UserID:    0,
			CollegeID: 1,
		}

		mockRepo.On("GetProfileByUserID", ctx, 0).Return(expectedProfile, nil)

		result, err := service.GetProfileByUserID(ctx, 0)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateProfile_EdgeCases(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProfileRepository)
	service := NewProfileService(mockRepo)

	t.Run("nil request", func(t *testing.T) {
		existingProfile := &models.Profile{
			ID:        1,
			UserID:    123,
			CollegeID: 1,
		}

		mockRepo.On("GetProfileByUserID", ctx, 123).Return(existingProfile, nil)
		mockRepo.On("UpdateProfilePartial", ctx, 1, (*models.UpdateProfileRequest)(nil)).Return(nil)

		err := service.UpdateProfile(ctx, 123, nil)

		// The service passes the request to the repository, so we need to check
		// how the repository handles nil
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty update request", func(t *testing.T) {
		existingProfile := &models.Profile{
			ID:        1,
			UserID:    123,
			CollegeID: 1,
		}

		req := &models.UpdateProfileRequest{}

		mockRepo.On("GetProfileByUserID", ctx, 123).Return(existingProfile, nil)
		mockRepo.On("UpdateProfilePartial", ctx, 1, req).Return(nil)

		err := service.UpdateProfile(ctx, 123, req)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
