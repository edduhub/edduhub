//go:build integration
// +build integration

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"eduhub/server/internal/models"
	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of the AuthService interface.
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateSession(ctx context.Context, sessionToken string) (*auth.Identity, error) {
	args := m.Called(ctx, sessionToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Identity), args.Error(1)
}

func (m *MockAuthService) HasRole(identity *auth.Identity, role string) bool {
	args := m.Called(identity, role)
	return args.Bool(0)
}

func (m *MockAuthService) CheckPermission(ctx context.Context, identity *auth.Identity, subject, resource, action string) (bool, error) {
	args := m.Called(ctx, identity, subject, resource, action)
	return args.Bool(0), args.Error(1)
}

// MockStudentService is a mock implementation of the StudentService interface.
type MockStudentService struct {
	mock.Mock
}

func (m *MockStudentService) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	args := m.Called(ctx, kratosID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}

func TestNewAuthMiddleware(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)

	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)

	assert.NotNil(t, middleware)
	assert.Equal(t, mockAuthSvc, middleware.AuthService)
	assert.Equal(t, mockStudentSvc, middleware.StudentService)
}

func TestAuthMiddleware_ValidateSession(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	validToken := "valid-token"
	invalidToken := "invalid-token"
	identity := &auth.Identity{ID: "test-id"}

	// Valid session
	mockAuthSvc.On("ValidateSession", mock.Anything, validToken).Return(identity, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session-Token", validToken)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	next := func(c echo.Context) error {
		assert.Equal(t, identity, c.Get(identityContextKey))
		return nil
	}
	err := middleware.ValidateSession(next)(c)
	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)

	// No session token
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.ValidateSession(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// Invalid session token
	mockAuthSvc.On("ValidateSession", mock.Anything, invalidToken).Return(nil, errors.New("invalid session"))
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session-Token", invalidToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.ValidateSession(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestAuthMiddleware_RequireCollege(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	collegeID := 123
	identity := &auth.Identity{Traits: auth.Traits{College: auth.College{ID: collegeID}}}

	// Identity in context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("identity", identity)
	next := func(c echo.Context) error {
		assert.Equal(t, collegeID, c.Get("college_id"))
		return nil
	}
	err := middleware.RequireCollege(next)(c)
	assert.NoError(t, err)

	// No identity in context
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.RequireCollege(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_LoadStudentProfile(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	kratosID := "student-kratos-id"
	studentID := 1
	student := &models.Student{StudentID: studentID, KratosID: kratosID, IsActive: true}
	studentNotActive := &models.Student{StudentID: studentID, KratosID: kratosID, IsActive: false}

	// Identity in context, student role, student found
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(student, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleStudent}})
	next := func(c echo.Context) error {
		assert.Equal(t, studentID, c.Get(studentIDContextKey))
		return nil
	}
	err := middleware.LoadStudentProfile(next)(c)
	assert.NoError(t, err)
	mockStudentSvc.AssertExpectations(t)

	// Identity in context, student role, student NOT found
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(nil, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleStudent}})

	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockStudentSvc.AssertExpectations(t)

	// Identity in context, student role, student NOT active
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(studentNotActive, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleStudent}})

	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockStudentSvc.AssertExpectations(t)

	// Identity in context, non-student role
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleAdmin}})
	err = middleware.LoadStudentProfile(next)(c)
	assert.NoError(t, err)
	mockStudentSvc.AssertExpectations(t)

	// No identity in context
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockStudentSvc.AssertExpectations(t)
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	identity := &auth.Identity{ID: "test-id"}
	roleAdmin := "admin"
	roleStudent := "student"

	// User has role
	mockAuthSvc.On("HasRole", identity, roleAdmin).Return(true)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("identity", identity)
	next := func(c echo.Context) error {
		return nil
	}
	err := middleware.RequireRole(roleAdmin)(next)(c)
	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)

	// User does not have role
	mockAuthSvc.On("HasRole", identity, roleStudent).Return(false)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identity)
	err = middleware.RequireRole(roleStudent)(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.RequireRole(roleAdmin)(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_RequirePermission(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	identity := &auth.Identity{ID: "test-id"}
	subject := "test-subject"
	resource := "test-resource"
	action := "test-action"

	// User has permission
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, subject, resource, action).Return(true, nil)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("identity", identity)
	next := func(c echo.Context) error {
		return nil
	}
	err := middleware.RequirePermission(subject, resource, action)(next)(c)
	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)

	// User does not have permission
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, subject, resource, action).Return(false, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identity)
	err = middleware.RequirePermission(subject, resource, action)(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// Permission check error
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, subject, resource, action).Return(false, errors.New("permission check error"))
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identity)
	err = middleware.RequirePermission(subject, resource, action)(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.RequirePermission(subject, resource, action)(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_VerifyStudentOwnership(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	studentID := 123
	otherStudentID := 456
	identity := &auth.Identity{ID: "test-id"}

	// Student accessing own resource
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(identityContextKey, identity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(studentID))

	next := func(c echo.Context) error {
		return nil
	}
	err := middleware.VerifyStudentOwnership(next)(c)
	assert.NoError(t, err)

	// Student accessing other's resource, no admin/faculty override
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, strconv.Itoa(otherStudentID), MarkAction, AttendanceResource).Return(false, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, identity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(otherStudentID))
	err = middleware.VerifyStudentOwnership(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// Student accessing other's resource, with admin/faculty override
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, strconv.Itoa(otherStudentID), MarkAction, AttendanceResource).Return(true, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, identity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(otherStudentID))
	err = middleware.VerifyStudentOwnership(next)(c)
	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)

	// Invalid student ID
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, identity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues("invalid")
	err = middleware.VerifyStudentOwnership(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.VerifyStudentOwnership(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// No studentIDContextKey
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, identity)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(studentID))
	err = middleware.VerifyStudentOwnership(next)(c)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
