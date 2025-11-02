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

func (m *MockAuthService) ValidateJWT(ctx context.Context, token string) (*auth.Identity, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Identity), args.Error(1)
}

func (m *MockAuthService) HasRole(identity *auth.Identity, role string) bool {
	args := m.Called(identity, role)
	return args.Bool(0)
}

func (m *MockAuthService) CheckPermission(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error) {
	args := m.Called(ctx, identity, action, resource)
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

	// No session token, fallback to JWT
	mockAuthSvc.On("ValidateJWT", mock.Anything, validToken).Return(identity, nil)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.ValidateSession(next)(c)
	assert.NoError(t, err)
	mockAuthSvc.AssertExpectations(t)

	// No valid session or token
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.ValidateSession(next)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)

	// Invalid session token
	mockAuthSvc.On("ValidateSession", mock.Anything, invalidToken).Return(nil, errors.New("invalid session"))
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Session-Token", invalidToken)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.ValidateSession(next)(c)
	assert.Error(t, err)
	httpErr, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
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
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
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
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(student, nil).Once()
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
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(nil, nil).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleStudent}})

	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	mockStudentSvc.AssertExpectations(t)

	// Identity in context, student role, student NOT active
	mockStudentSvc.On("FindByKratosID", mock.Anything, kratosID).Return(studentNotActive, nil).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleStudent}})

	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	httpErr, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	mockStudentSvc.AssertExpectations(t)

	// Identity in context, non-student role
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, &auth.Identity{ID: kratosID, Traits: auth.Traits{Role: RoleAdmin}})
	err = middleware.LoadStudentProfile(next)(c)
	assert.NoError(t, err)

	// No identity in context
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	err = middleware.LoadStudentProfile(next)(c)
	assert.Error(t, err)
	httpErr, ok = err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusForbidden, httpErr.Code)
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	identityAdmin := &auth.Identity{ID: "admin-id", Traits: auth.Traits{Role: "admin"}}
	identityStudent := &auth.Identity{ID: "student-id", Traits: auth.Traits{Role: "student"}}
	identityFaculty := &auth.Identity{ID: "faculty-id", Traits: auth.Traits{Role: "faculty"}}

	roleAdmin := "admin"
	roleStudent := "student"
	roleFaculty := "faculty"

	// User has role
	mockAuthSvc.On("HasRole", identityAdmin, roleAdmin).Return(true).Once()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("identity", identityAdmin)
	handlerCalled := false
	next := func(c echo.Context) error {
		handlerCalled = true
		return c.JSON(http.StatusOK, "success")
	}
	err := middleware.RequireRole(roleAdmin)(next)(c)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// User does not have role
	mockAuthSvc.On("HasRole", identityStudent, roleAdmin).Return(false).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identityStudent)
	handlerCalled = false
	err = middleware.RequireRole(roleAdmin)(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	err = middleware.RequireRole(roleAdmin)(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Multiple roles, user has one
	mockAuthSvc.On("HasRole", identityFaculty, roleAdmin).Return(false).Once()
	mockAuthSvc.On("HasRole", identityFaculty, roleFaculty).Return(true).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identityFaculty)
	handlerCalled = false
	err = middleware.RequireRole(roleAdmin, roleFaculty)(next)(c)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthSvc.AssertExpectations(t)
}

func TestAuthMiddleware_RequirePermission(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	identity := &auth.Identity{ID: "test-id"}
	resource := "grades"
	action := "read"

	// User has permission
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, action, resource).Return(true, nil).Once()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("identity", identity)
	handlerCalled := false
	next := func(c echo.Context) error {
		handlerCalled = true
		return c.JSON(http.StatusOK, "success")
	}
	err := middleware.RequirePermission(identity.ID, resource, action)(next)(c)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// User does not have permission
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, "write", resource).Return(false, nil).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identity)
	handlerCalled = false
	err = middleware.RequirePermission(identity.ID, resource, "write")(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// Permission check error
	mockAuthSvc.On("CheckPermission", mock.Anything, identity, action, resource).Return(false, errors.New("permission check error")).Once()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set("identity", identity)
	handlerCalled = false
	err = middleware.RequirePermission(identity.ID, resource, action)(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockAuthSvc.AssertExpectations(t)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	handlerCalled = false
	err = middleware.RequirePermission(identity.ID, resource, action)(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_VerifyStudentOwnership(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockStudentSvc := new(MockStudentService)
	middleware := NewAuthMiddleware(mockAuthSvc, mockStudentSvc)
	studentID := 123
	otherStudentID := 456
	studentIdentity := &auth.Identity{ID: "student-id", Traits: auth.Traits{Role: RoleStudent}}
	adminIdentity := &auth.Identity{ID: "admin-id", Traits: auth.Traits{Role: RoleAdmin}}

	// Student accessing own resource
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(identityContextKey, studentIdentity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(studentID))

	handlerCalled := false
	next := func(c echo.Context) error {
		handlerCalled = true
		return nil
	}
	err := middleware.VerifyStudentOwnership()(next)(c)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)

	// Student accessing other's resource
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, studentIdentity)
	c.Set(studentIDContextKey, studentID)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(otherStudentID))
	handlerCalled = false
	err = middleware.VerifyStudentOwnership()(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, rec.Code)

	// Admin accessing student resource
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, adminIdentity)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(studentID))
	handlerCalled = false
	err = middleware.VerifyStudentOwnership()(next)(c)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)

	// Invalid student ID
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.Set(identityContextKey, adminIdentity)
	c.SetParamNames("studentID")
	c.SetParamValues("invalid")
	handlerCalled = false
	err = middleware.VerifyStudentOwnership()(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// No identity
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("studentID")
	c.SetParamValues(strconv.Itoa(studentID))
	handlerCalled = false
	err = middleware.VerifyStudentOwnership()(next)(c)
	assert.NoError(t, err)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}