package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"eduhub/server/internal/models"
	"eduhub/server/internal/services/auth"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock types ---

type mockTokenValidator struct {
	validateJWTFunc       func(ctx context.Context, token string) (*auth.Identity, error)
	validateTokenFunc     func(ctx context.Context, accessToken string) (*auth.Identity, error)
	hasRoleFunc           func(identity *auth.Identity, role string) bool
	checkPermissionFunc   func(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error)
}

func (m *mockTokenValidator) ValidateJWT(ctx context.Context, token string) (*auth.Identity, error) {
	if m.validateJWTFunc != nil {
		return m.validateJWTFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, accessToken string) (*auth.Identity, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, accessToken)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTokenValidator) HasRole(identity *auth.Identity, role string) bool {
	if m.hasRoleFunc != nil {
		return m.hasRoleFunc(identity, role)
	}
	return false
}

func (m *mockTokenValidator) CheckPermission(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error) {
	if m.checkPermissionFunc != nil {
		return m.checkPermissionFunc(ctx, identity, action, resource)
	}
	return false, nil
}

type mockStudentLoader struct {
	findByKratosIDFunc func(ctx context.Context, kratosID string) (*models.Student, error)
}

func (m *mockStudentLoader) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	if m.findByKratosIDFunc != nil {
		return m.findByKratosIDFunc(ctx, kratosID)
	}
	return nil, errors.New("not found")
}

func newAuthEchoContext(method, path string, headers map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// --- extractBearer ---

func TestExtractBearer(t *testing.T) {
	t.Run("valid bearer token", func(t *testing.T) {
		assert.Equal(t, "abc123", extractBearer("Bearer abc123"))
	})

	t.Run("case insensitive", func(t *testing.T) {
		assert.Equal(t, "token", extractBearer("bearer token"))
	})

	t.Run("BEARER uppercase", func(t *testing.T) {
		assert.Equal(t, "tok", extractBearer("BEARER tok"))
	})

	t.Run("trims whitespace from token", func(t *testing.T) {
		assert.Equal(t, "tok", extractBearer("Bearer  tok "))
	})

	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", extractBearer(""))
	})

	t.Run("no bearer prefix", func(t *testing.T) {
		assert.Equal(t, "", extractBearer("Basic abc123"))
	})

	t.Run("bearer with no token", func(t *testing.T) {
		assert.Equal(t, "", extractBearer("Bearer"))
	})

	t.Run("just bearer space", func(t *testing.T) {
		assert.Equal(t, "", extractBearer("Bearer "))
	})
}

// --- GetIdentity ---

func TestGetIdentity(t *testing.T) {
	t.Run("returns identity when set", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		expected := &auth.Identity{ID: "abc"}
		c.Set("identity", expected)

		id, ok := GetIdentity(c)
		assert.True(t, ok)
		assert.Equal(t, expected, id)
	})

	t.Run("returns false when not set", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		_, ok := GetIdentity(c)
		assert.False(t, ok)
	})

	t.Run("returns false when wrong type", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", "not-an-identity")

		_, ok := GetIdentity(c)
		assert.False(t, ok)
	})
}

// --- GetStudentID ---

func TestGetStudentIDHelper(t *testing.T) {
	t.Run("returns student ID when set", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("student_id", 42)

		id, ok := GetStudentID(c)
		assert.True(t, ok)
		assert.Equal(t, 42, id)
	})

	t.Run("returns false when not set", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		_, ok := GetStudentID(c)
		assert.False(t, ok)
	})

	t.Run("returns false when wrong type", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("student_id", "not-an-int")

		_, ok := GetStudentID(c)
		assert.False(t, ok)
	})
}

// --- GetCollegeID ---

func TestGetCollegeIDHelper(t *testing.T) {
	t.Run("returns int college ID", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("college_id", 5)

		id, ok := GetCollegeID(c)
		assert.True(t, ok)
		assert.Equal(t, 5, id)
	})

	t.Run("returns string college ID", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("college_id", "uuid-abc")

		id, ok := GetCollegeID(c)
		assert.True(t, ok)
		assert.Equal(t, "uuid-abc", id)
	})

	t.Run("returns false when not set", func(t *testing.T) {
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		_, ok := GetCollegeID(c)
		assert.False(t, ok)
	})
}

// --- ValidateToken middleware ---

func TestAuthMiddleware_ValidateToken(t *testing.T) {
	t.Run("rejects missing auth header", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.ValidateToken(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		require.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		validator := &mockTokenValidator{
			validateTokenFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return nil, errors.New("invalid token")
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer bad-token"})

		handler := mw.ValidateToken(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		require.Error(t, err)
		he := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("sets identity on success", func(t *testing.T) {
		identity := &auth.Identity{ID: "kratos-123", UserID: 10}
		validator := &mockTokenValidator{
			validateTokenFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return identity, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer valid-token"})

		var ctxIdentity *auth.Identity
		handler := mw.ValidateToken(func(c echo.Context) error {
			ctxIdentity, _ = c.Get("identity").(*auth.Identity)
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, identity, ctxIdentity)
	})

	t.Run("sets user_id when identity has UserID", func(t *testing.T) {
		identity := &auth.Identity{ID: "k-1", UserID: 42}
		validator := &mockTokenValidator{
			validateTokenFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return identity, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer tok"})

		var uid any
		handler := mw.ValidateToken(func(c echo.Context) error {
			uid = c.Get("user_id")
			return nil
		})

		_ = handler(c)
		assert.Equal(t, 42, uid)
	})

	t.Run("does not set user_id when UserID is zero", func(t *testing.T) {
		identity := &auth.Identity{ID: "k-1", UserID: 0}
		validator := &mockTokenValidator{
			validateTokenFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return identity, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer tok"})

		var uid any
		handler := mw.ValidateToken(func(c echo.Context) error {
			uid = c.Get("user_id")
			return nil
		})

		_ = handler(c)
		assert.Nil(t, uid)
	})
}

// --- ValidateJWT middleware ---

func TestAuthMiddleware_ValidateJWT(t *testing.T) {
	t.Run("rejects missing auth header", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.ValidateJWT(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
		he := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("rejects invalid JWT", func(t *testing.T) {
		validator := &mockTokenValidator{
			validateJWTFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return nil, errors.New("expired")
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer expired-token"})

		handler := mw.ValidateJWT(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
		he := err.(*echo.HTTPError)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("sets identity on success", func(t *testing.T) {
		identity := &auth.Identity{ID: "jwt-id", UserID: 5}
		validator := &mockTokenValidator{
			validateJWTFunc: func(ctx context.Context, token string) (*auth.Identity, error) {
				return identity, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", map[string]string{"Authorization": "Bearer valid-jwt"})

		var ctxIdentity *auth.Identity
		handler := mw.ValidateJWT(func(c echo.Context) error {
			ctxIdentity, _ = c.Get("identity").(*auth.Identity)
			return nil
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, identity, ctxIdentity)
	})
}

// --- RequireCollege middleware ---

func TestAuthMiddleware_RequireCollege(t *testing.T) {
	t.Run("rejects when no identity", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.RequireCollege(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})

	t.Run("rejects empty college ID", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{College: auth.College{ID: ""}}})

		handler := mw.RequireCollege(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})

	t.Run("sets numeric college ID as int", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{College: auth.College{ID: "42"}}})

		var collegeID any
		handler := mw.RequireCollege(func(c echo.Context) error {
			collegeID = c.Get("college_id")
			return nil
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 42, collegeID)
	})

	t.Run("sets non-numeric college ID as string", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{College: auth.College{ID: "uuid-abc-123"}}})

		var collegeID any
		handler := mw.RequireCollege(func(c echo.Context) error {
			collegeID = c.Get("college_id")
			return nil
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, "uuid-abc-123", collegeID)
	})
}

// --- LoadStudentProfile middleware ---

func TestAuthMiddleware_LoadStudentProfile(t *testing.T) {
	t.Run("rejects when no identity", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})

	t.Run("passes through for non-student roles", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "abc", Traits: auth.Traits{Role: "admin"}})

		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("loads student profile for student role", func(t *testing.T) {
		student := &models.Student{StudentID: 99, IsActive: true}
		loader := &mockStudentLoader{
			findByKratosIDFunc: func(ctx context.Context, kratosID string) (*models.Student, error) {
				return student, nil
			},
		}
		mw := NewAuthMiddleware(&mockTokenValidator{}, loader, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "kratos-1", Traits: auth.Traits{Role: "student"}})

		var studentID any
		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			studentID = c.Get("student_id")
			return nil
		})

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 99, studentID)
	})

	t.Run("rejects when student not found", func(t *testing.T) {
		loader := &mockStudentLoader{
			findByKratosIDFunc: func(ctx context.Context, kratosID string) (*models.Student, error) {
				return nil, errors.New("not found")
			},
		}
		mw := NewAuthMiddleware(&mockTokenValidator{}, loader, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "kratos-1", Traits: auth.Traits{Role: "student"}})

		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})

	t.Run("rejects nil student", func(t *testing.T) {
		loader := &mockStudentLoader{
			findByKratosIDFunc: func(ctx context.Context, kratosID string) (*models.Student, error) {
				return nil, nil
			},
		}
		mw := NewAuthMiddleware(&mockTokenValidator{}, loader, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "kratos-1", Traits: auth.Traits{Role: "student"}})

		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})

	t.Run("rejects inactive student", func(t *testing.T) {
		student := &models.Student{StudentID: 1, IsActive: false}
		loader := &mockStudentLoader{
			findByKratosIDFunc: func(ctx context.Context, kratosID string) (*models.Student, error) {
				return student, nil
			},
		}
		mw := NewAuthMiddleware(&mockTokenValidator{}, loader, nil, nil)
		c, _ := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "kratos-1", Traits: auth.Traits{Role: "student"}})

		handler := mw.LoadStudentProfile(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		require.Error(t, err)
	})
}

// --- RequireRole middleware ---

func TestAuthMiddleware_RequireRole(t *testing.T) {
	makeValidator := func(hasRole bool) *mockTokenValidator {
		return &mockTokenValidator{
			hasRoleFunc: func(identity *auth.Identity, role string) bool {
				return hasRole
			},
		}
	}

	t.Run("rejects when no identity", func(t *testing.T) {
		mw := NewAuthMiddleware(makeValidator(true), &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.RequireRole("admin")(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("allows matching role", func(t *testing.T) {
		mw := NewAuthMiddleware(makeValidator(true), &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "admin"}})

		handler := mw.RequireRole("admin")(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects non-matching role", func(t *testing.T) {
		mw := NewAuthMiddleware(makeValidator(false), &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})

		handler := mw.RequireRole("admin")(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

// --- RequirePermission middleware ---

func TestAuthMiddleware_RequirePermission(t *testing.T) {
	t.Run("rejects when no identity", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)

		handler := mw.RequirePermission("user", "resource", "action")(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("allows when permission granted", func(t *testing.T) {
		validator := &mockTokenValidator{
			checkPermissionFunc: func(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error) {
				return true, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "u1"})

		handler := mw.RequirePermission("user", "resource", "action")(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects when permission denied", func(t *testing.T) {
		validator := &mockTokenValidator{
			checkPermissionFunc: func(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error) {
				return false, nil
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "u1"})

		handler := mw.RequirePermission("user", "resource", "action")(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("returns 500 on permission check error", func(t *testing.T) {
		validator := &mockTokenValidator{
			checkPermissionFunc: func(ctx context.Context, identity *auth.Identity, action, resource string) (bool, error) {
				return false, errors.New("keto error")
			},
		}
		mw := NewAuthMiddleware(validator, &mockStudentLoader{}, nil, nil)
		c, rec := newAuthEchoContext(http.MethodGet, "/", nil)
		c.Set("identity", &auth.Identity{ID: "u1"})

		handler := mw.RequirePermission("user", "resource", "action")(func(c echo.Context) error {
			return nil
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

// --- VerifyStudentOwnership middleware ---

func TestAuthMiddleware_VerifyStudentOwnership(t *testing.T) {
	t.Run("rejects when no identity", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("1")

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("rejects missing studentID param", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rejects invalid studentID param", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("abc")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("allows student accessing own data", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/42", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("42")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})
		c.Set("student_id", 42)

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		_ = handler(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects student accessing other student data", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/99", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("99")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})
		c.Set("student_id", 42)

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("allows admin accessing any student data", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/99", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("99")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "admin"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		_ = handler(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("allows faculty accessing any student data", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/99", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("99")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "faculty"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		_ = handler(c)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("rejects unknown role", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("1")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "unknown"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("rejects student with no student_id in context", func(t *testing.T) {
		mw := NewAuthMiddleware(&mockTokenValidator{}, &mockStudentLoader{}, nil, nil)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/students/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("studentID")
		c.SetParamValues("1")
		c.Set("identity", &auth.Identity{Traits: auth.Traits{Role: "student"}})

		handler := mw.VerifyStudentOwnership()(func(c echo.Context) error {
			return nil
		})

		_ = handler(c)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
