package handler

import (
	"bytes"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/announcement"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnnouncementService is a mock implementation of the AnnouncementService
type MockAnnouncementService struct {
	mock.Mock
}

func (m *MockAnnouncementService) CreateAnnouncement(c echo.Context, req *models.CreateAnnouncementRequest, collegeID, userID int) (*models.Announcement, error) {
	args := m.Called(c, req, collegeID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementService) GetAnnouncementByID(c echo.Context, id int) (*models.Announcement, error) {
	args := m.Called(c, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementService) GetAnnouncementsByCollegeID(c echo.Context, collegeID, limit, offset int) ([]*models.Announcement, error) {
	args := m.Called(c, collegeID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementService) UpdateAnnouncement(c echo.Context, id int, req *models.UpdateAnnouncementRequest) (*models.Announcement, error) {
	args := m.Called(c, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Announcement), args.Error(1)
}

func (m *MockAnnouncementService) DeleteAnnouncement(c echo.Context, id int) error {
	args := m.Called(c, id)
	return args.Error(0)
}

func TestCreateAnnouncementHandler(t *testing.T) {
	e := echo.New()
	reqBody := &models.CreateAnnouncementRequest{
		Title:   "Test Announcement",
		Content: "This is a test announcement.",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/announcements", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("college_id", 1)
	c.Set("user_id", 1)

	mockService := new(MockAnnouncementService)
	handler := NewAnnouncementHandler(mockService)

	expectedAnnouncement := &models.Announcement{
		ID:        1,
		Title:     reqBody.Title,
		Content:   reqBody.Content,
		CollegeID: 1,
		UserID:    1,
	}

	mockService.On("CreateAnnouncement", mock.Anything, reqBody, 1, 1).Return(expectedAnnouncement, nil)

	if assert.NoError(t, handler.CreateAnnouncement(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		var response models.Announcement
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedAnnouncement, &response)
	}
}
