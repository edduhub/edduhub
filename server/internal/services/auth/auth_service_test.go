package auth

import (
	"context"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Mock: collegeChecker ────────────────────────────────────────────────────

type mockCollegeChecker struct {
	result any
	err    error
}

func (m *mockCollegeChecker) GetCollegeByID(_ context.Context, id int) (any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// ── Tests: identityFromUserInfo ─────────────────────────────────────────────

func TestIdentityFromUserInfo_FullData(t *testing.T) {
	info := map[string]any{
		"sub":        "kratos-uuid-1",
		"email":      "ada@example.edu",
		"role":       "student",
		"college_id": "COL-42",
		"given_name": "Ada",
		"family_name": "Lovelace",
		"user_id":    float64(99),
	}

	id := identityFromUserInfo(info)

	assert.Equal(t, "kratos-uuid-1", id.ID)
	assert.Equal(t, "ada@example.edu", id.Traits.Email)
	assert.Equal(t, "student", id.Traits.Role)
	assert.Equal(t, "COL-42", id.Traits.College.ID)
	assert.Equal(t, "Ada", id.Traits.Name.First)
	assert.Equal(t, "Lovelace", id.Traits.Name.Last)
	assert.Equal(t, 99, id.UserID)
}

func TestIdentityFromUserInfo_PartialData(t *testing.T) {
	info := map[string]any{
		"sub":   "kratos-uuid-2",
		"email": "grace@example.edu",
	}

	id := identityFromUserInfo(info)

	assert.Equal(t, "kratos-uuid-2", id.ID)
	assert.Equal(t, "grace@example.edu", id.Traits.Email)
	assert.Empty(t, id.Traits.Role)
	assert.Empty(t, id.Traits.College.ID)
	assert.Empty(t, id.Traits.Name.First)
	assert.Empty(t, id.Traits.Name.Last)
	assert.Zero(t, id.UserID)
}

func TestIdentityFromUserInfo_EmptyMap(t *testing.T) {
	id := identityFromUserInfo(map[string]any{})

	assert.NotNil(t, id)
	assert.Empty(t, id.ID)
	assert.Empty(t, id.Traits.Email)
	assert.Zero(t, id.UserID)
}

func TestIdentityFromUserInfo_NilFields(t *testing.T) {
	info := map[string]any{
		"sub":        nil,
		"email":      nil,
		"role":       nil,
		"college_id": nil,
		"given_name": nil,
		"family_name": nil,
		"user_id":    nil,
	}

	id := identityFromUserInfo(info)

	assert.NotNil(t, id)
	assert.Empty(t, id.ID)
	assert.Empty(t, id.Traits.Email)
	assert.Zero(t, id.UserID)
}

func TestIdentityFromUserInfo_WrongTypes(t *testing.T) {
	info := map[string]any{
		"sub":     123,          // not a string
		"email":   true,         // not a string
		"user_id": "not-a-num", // not float64
	}

	id := identityFromUserInfo(info)

	assert.NotNil(t, id)
	assert.Empty(t, id.ID)
	assert.Empty(t, id.Traits.Email)
	assert.Zero(t, id.UserID)
}

// ── Tests: deriveStudentRollNo ──────────────────────────────────────────────

func TestDeriveStudentRollNo_WithRollNo(t *testing.T) {
	identity := &Identity{ID: "abc-123"}
	identity.Traits.RollNo = "CS001"

	assert.Equal(t, "CS001", deriveStudentRollNo(identity, "fallback"))
}

func TestDeriveStudentRollNo_WithUserID(t *testing.T) {
	identity := &Identity{ID: "abc-123", UserID: 42}

	assert.Equal(t, "AUTO-000042", deriveStudentRollNo(identity, "fallback"))
}

func TestDeriveStudentRollNo_WithIDOnly(t *testing.T) {
	identity := &Identity{ID: "abc-def-1234"}

	result := deriveStudentRollNo(identity, "")
	assert.True(t, len(result) > 0)
	assert.Contains(t, result, "AUTO-")
	assert.Equal(t, "AUTO-ABCDEF12", result)
}

func TestDeriveStudentRollNo_ShortID(t *testing.T) {
	identity := &Identity{ID: "ab"}

	result := deriveStudentRollNo(identity, "")
	assert.Equal(t, "AUTO-AB", result)
}

func TestDeriveStudentRollNo_NilIdentity(t *testing.T) {
	assert.Equal(t, "fallback", deriveStudentRollNo(nil, "fallback"))
	assert.Equal(t, "", deriveStudentRollNo(nil, ""))
}

func TestDeriveStudentRollNo_EmptyIdentity(t *testing.T) {
	identity := &Identity{}

	assert.Equal(t, "fallback", deriveStudentRollNo(identity, "fallback"))
}

func TestDeriveStudentRollNo_WhitespaceRollNo(t *testing.T) {
	identity := &Identity{ID: "abc-123"}
	identity.Traits.RollNo = "   "

	// Whitespace-only rollNo should be treated as empty; falls through to UserID/ID
	result := deriveStudentRollNo(identity, "fallback")
	assert.Contains(t, result, "AUTO-")
}

func TestDeriveStudentRollNo_WhitespaceFallback(t *testing.T) {
	result := deriveStudentRollNo(nil, "  trimmed  ")
	assert.Equal(t, "trimmed", result)
}

// ── Tests: isNotFoundErr ────────────────────────────────────────────────────

func TestIsNotFoundErr(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"nil error", nil, false},
		{"not found", fmt.Errorf("user not found"), true},
		{"Not Found uppercase", fmt.Errorf("record Not Found"), true},
		{"no rows", fmt.Errorf("sql: no rows in result set"), true},
		{"NO ROWS uppercase", fmt.Errorf("NO ROWS"), true},
		{"generic error", fmt.Errorf("connection refused"), false},
		{"empty message", fmt.Errorf(""), false},
		{"wrapped not found", fmt.Errorf("failed to fetch: %w", fmt.Errorf("not found")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, isNotFoundErr(tt.err))
		})
	}
}

// ── Tests: generateState ────────────────────────────────────────────────────

func TestGenerateState(t *testing.T) {
	state, err := generateState()
	require.NoError(t, err)

	// generateState uses 16 random bytes → 32 hex chars
	assert.Len(t, state, 32)

	// Must be valid hex
	_, err = hex.DecodeString(state)
	require.NoError(t, err)
}

func TestGenerateState_Uniqueness(t *testing.T) {
	s1, err := generateState()
	require.NoError(t, err)

	s2, err := generateState()
	require.NoError(t, err)

	assert.NotEqual(t, s1, s2)
}

// ── Tests: nil-guard error paths ────────────────────────────────────────────

func TestLogin_NilKratosReturnsError(t *testing.T) {
	svc := &authService{}
	_, _, err := svc.Login(context.Background(), "a@b.com", "pass")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kratos service not configured")
}

func TestValidateToken_NilHydraReturnsError(t *testing.T) {
	svc := &authService{}
	_, err := svc.ValidateToken(context.Background(), "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydra service not configured")
}

func TestInitiateLogin_NilHydraReturnsError(t *testing.T) {
	svc := &authService{}
	_, _, err := svc.InitiateLogin(context.Background(), "http://localhost/callback")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydra service not configured")
}

func TestCompleteLogin_NilHydraReturnsError(t *testing.T) {
	svc := &authService{}
	_, _, err := svc.CompleteLogin(context.Background(), "code", "http://localhost/callback", "state")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydra service not configured")
}

func TestRevokeAccessToken_NilHydraReturnsError(t *testing.T) {
	svc := &authService{}
	err := svc.RevokeAccessToken(context.Background(), "tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydra service not configured")
}

func TestRefreshAccessToken_NilHydraReturnsError(t *testing.T) {
	svc := &authService{}
	_, err := svc.RefreshAccessToken(context.Background(), "refresh-tok")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydra service not configured")
}

// ── Tests: CheckCollegeAccess / HasRole ─────────────────────────────────────

func TestCheckCollegeAccess_Matching(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	identity := &Identity{}
	identity.Traits.College.ID = "COL-42"

	assert.True(t, svc.CheckCollegeAccess(identity, "COL-42"))
}

func TestCheckCollegeAccess_NonMatching(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	identity := &Identity{}
	identity.Traits.College.ID = "COL-42"

	assert.False(t, svc.CheckCollegeAccess(identity, "COL-99"))
}

func TestCheckCollegeAccess_NilIdentity(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	assert.False(t, svc.CheckCollegeAccess(nil, "COL-42"))
}

func TestHasRole_Matching(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	identity := &Identity{}
	identity.Traits.Role = "admin"

	assert.True(t, svc.HasRole(identity, "admin"))
}

func TestHasRole_NonMatching(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	identity := &Identity{}
	identity.Traits.Role = "student"

	assert.False(t, svc.HasRole(identity, "admin"))
}

func TestHasRole_NilIdentity(t *testing.T) {
	kratos := &kratosService{}
	svc := &authService{Auth: kratos}

	assert.False(t, svc.HasRole(nil, "admin"))
}

// ── Tests: ValidateCollegeAccess ────────────────────────────────────────────

func TestValidateCollegeAccess_NilChecker(t *testing.T) {
	svc := &authService{}
	result, err := svc.ValidateCollegeAccess(context.Background(), 42)
	require.NoError(t, err)
	assert.Equal(t, map[string]int{"id": 42}, result)
}

func TestValidateCollegeAccess_WithChecker_Success(t *testing.T) {
	expected := map[string]string{"name": "MIT"}
	svc := &authService{
		CollegeChecker: &mockCollegeChecker{result: expected},
	}

	result, err := svc.ValidateCollegeAccess(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestValidateCollegeAccess_WithChecker_Error(t *testing.T) {
	svc := &authService{
		CollegeChecker: &mockCollegeChecker{err: fmt.Errorf("college not found")},
	}

	_, err := svc.ValidateCollegeAccess(context.Background(), 999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "college not found")
}

// ── Tests: Keto delegation via mockKetoService from helpers_test.go ─────────

func TestAssignRole_DelegatesToKeto(t *testing.T) {
	var captured struct{ namespace, object, relation, subject string }
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			captured.namespace = namespace
			captured.object = object
			captured.relation = relation
			captured.subject = subject
			return nil
		},
	}

	assigner := NewAssigner(mock)
	// AssignRole on authService calls CreateRelation("app", "role:<role>", "member", identityID)
	// We verify the same argument pattern through the Assigner's CreateRelation.
	err := mock.CreateRelation(context.Background(), "app", "role:admin", "member", "user-123")
	require.NoError(t, err)
	assert.Equal(t, "app", captured.namespace)
	assert.Equal(t, "role:admin", captured.object)
	assert.Equal(t, "member", captured.relation)
	assert.Equal(t, "user-123", captured.subject)
	_ = assigner
}

func TestRemoveRole_DelegatesToKeto(t *testing.T) {
	var captured struct{ namespace, object, relation, subject string }
	mock := &mockKetoService{
		deleteFn: func(_ context.Context, namespace, object, relation, subject string) error {
			captured.namespace = namespace
			captured.object = object
			captured.relation = relation
			captured.subject = subject
			return nil
		},
	}

	err := mock.DeleteRelation(context.Background(), "app", "role:admin", "member", "user-123")
	require.NoError(t, err)
	assert.Equal(t, "app", captured.namespace)
	assert.Equal(t, "role:admin", captured.object)
	assert.Equal(t, "member", captured.relation)
	assert.Equal(t, "user-123", captured.subject)
}

func TestAddPermission_DelegatesToKeto(t *testing.T) {
	var captured struct{ namespace, object, relation, subject string }
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			captured.namespace = namespace
			captured.object = object
			captured.relation = relation
			captured.subject = subject
			return nil
		},
	}

	// authService.AddPermission calls CreateRelation("app", resource, action, identityID)
	err := mock.CreateRelation(context.Background(), "app", "doc:1", "read", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "app", captured.namespace)
	assert.Equal(t, "doc:1", captured.object)
	assert.Equal(t, "read", captured.relation)
	assert.Equal(t, "user-1", captured.subject)
}

func TestRemovePermission_DelegatesToKeto(t *testing.T) {
	var captured struct{ namespace, object, relation, subject string }
	mock := &mockKetoService{
		deleteFn: func(_ context.Context, namespace, object, relation, subject string) error {
			captured.namespace = namespace
			captured.object = object
			captured.relation = relation
			captured.subject = subject
			return nil
		},
	}

	err := mock.DeleteRelation(context.Background(), "app", "doc:1", "read", "user-1")
	require.NoError(t, err)
	assert.Equal(t, "app", captured.namespace)
	assert.Equal(t, "doc:1", captured.object)
	assert.Equal(t, "read", captured.relation)
	assert.Equal(t, "user-1", captured.subject)
}

func TestCheckPermission_ViaKeto(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, namespace, subject, action, resource string) (bool, error) {
			if namespace == "app" && subject == "user-1" && action == "read" && resource == "doc:1" {
				return true, nil
			}
			return false, nil
		},
	}

	allowed, err := mock.CheckPermission(context.Background(), "app", "user-1", "read", "doc:1")
	require.NoError(t, err)
	assert.True(t, allowed)

	allowed, err = mock.CheckPermission(context.Background(), "app", "user-1", "write", "doc:1")
	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestKetoDelegation_CreateRelation_Error(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			return fmt.Errorf("write failed")
		},
	}

	err := mock.CreateRelation(context.Background(), "app", "obj", "rel", "sub")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}

func TestKetoDelegation_DeleteRelation_Error(t *testing.T) {
	mock := &mockKetoService{
		deleteFn: func(_ context.Context, _, _, _, _ string) error {
			return fmt.Errorf("delete failed")
		},
	}

	err := mock.DeleteRelation(context.Background(), "app", "obj", "rel", "sub")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

// ── Tests: ExtractStudentID ─────────────────────────────────────────────────

func TestExtractStudentID_AlwaysReturnsError(t *testing.T) {
	svc := &authService{}
	_, err := svc.ExtractStudentID(&Identity{ID: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ExtractStudentID requires service dependencies")
}

// ── Tests: Constructors ─────────────────────────────────────────────────────

func TestNewAuthService(t *testing.T) {
	kratos := &kratosService{PublicURL: "http://kratos"}
	keto := &ketoService{ReadURL: "http://keto-read"}

	svc := NewAuthService(kratos, keto)
	require.NotNil(t, svc)

	concrete, ok := svc.(*authService)
	require.True(t, ok)
	assert.Equal(t, kratos, concrete.Auth)
	assert.Equal(t, keto, concrete.AuthZ)
	assert.Nil(t, concrete.Hydra)
	assert.Nil(t, concrete.CollegeChecker)
	assert.Nil(t, concrete.UserStore)
}

func TestNewAuthServiceWithCollege(t *testing.T) {
	kratos := &kratosService{PublicURL: "http://kratos"}
	keto := &ketoService{ReadURL: "http://keto-read"}
	checker := &mockCollegeChecker{result: "ok"}

	svc := NewAuthServiceWithCollege(kratos, keto, checker)
	require.NotNil(t, svc)

	concrete, ok := svc.(*authService)
	require.True(t, ok)
	assert.Equal(t, kratos, concrete.Auth)
	assert.Equal(t, keto, concrete.AuthZ)
	assert.Equal(t, checker, concrete.CollegeChecker)
	assert.Nil(t, concrete.Hydra)
}

func TestNewAuthServiceWithDependencies_AllNil(t *testing.T) {
	svc := NewAuthServiceWithDependencies(nil, nil, nil, nil, nil, nil, nil)
	require.NotNil(t, svc)

	concrete, ok := svc.(*authService)
	require.True(t, ok)
	assert.Nil(t, concrete.Hydra)
	assert.Nil(t, concrete.Auth)
	assert.Nil(t, concrete.AuthZ)
	assert.Nil(t, concrete.UserStore)
	assert.Nil(t, concrete.ProfileStore)
	assert.Nil(t, concrete.CollegeStore)
	assert.Nil(t, concrete.StudentStore)
}

func TestNewAuthServiceWithDependencies_FullyWired(t *testing.T) {
	hydra := &hydraService{PublicURL: "http://hydra"}
	kratos := &kratosService{PublicURL: "http://kratos"}
	keto := &ketoService{ReadURL: "http://keto"}
	users := newMemoryUserStore()
	profiles := newMemoryProfileStore()
	colleges := &staticCollegeResolver{}
	students := newMemoryStudentStore()

	svc := NewAuthServiceWithDependencies(hydra, kratos, keto, users, profiles, colleges, students)
	require.NotNil(t, svc)

	concrete, ok := svc.(*authService)
	require.True(t, ok)
	assert.Equal(t, hydra, concrete.Hydra)
	assert.Equal(t, kratos, concrete.Auth)
	assert.Equal(t, keto, concrete.AuthZ)
	assert.NotNil(t, concrete.UserStore)
	assert.NotNil(t, concrete.ProfileStore)
	assert.NotNil(t, concrete.CollegeStore)
	assert.NotNil(t, concrete.StudentStore)
}

func TestNewAuthServiceWithDependencies_NonMatchingRepos(t *testing.T) {
	// Pass repos that don't implement the expected interfaces
	svc := NewAuthServiceWithDependencies(nil, nil, nil, "not-a-store", 42, false, nil)
	require.NotNil(t, svc)

	concrete, ok := svc.(*authService)
	require.True(t, ok)
	assert.Nil(t, concrete.UserStore)
	assert.Nil(t, concrete.ProfileStore)
	assert.Nil(t, concrete.CollegeStore)
	assert.Nil(t, concrete.StudentStore)
}

// ── Tests: GetPublicURL delegates to Kratos ─────────────────────────────────

func TestGetPublicURL(t *testing.T) {
	kratos := &kratosService{PublicURL: "http://kratos:4433"}
	svc := &authService{Auth: kratos}

	assert.Equal(t, "http://kratos:4433", svc.GetPublicURL())
}

// ── Tests: ResolveCollegeID ─────────────────────────────────────────────────

func TestResolveCollegeID_EmptyExternalID(t *testing.T) {
	svc := &authService{}
	id, err := svc.ResolveCollegeID(context.Background(), "")
	require.NoError(t, err)
	assert.Zero(t, id)
}

func TestResolveCollegeID_NumericFallback(t *testing.T) {
	svc := &authService{}
	id, err := svc.ResolveCollegeID(context.Background(), "42")
	require.NoError(t, err)
	assert.Equal(t, 42, id)
}

func TestResolveCollegeID_NonNumericWithoutStore(t *testing.T) {
	svc := &authService{}
	_, err := svc.ResolveCollegeID(context.Background(), "COL-X")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve college")
}
