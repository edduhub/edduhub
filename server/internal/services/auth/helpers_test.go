package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKetoService implements KetoService for testing.
type mockKetoService struct {
	checkFn  func(ctx context.Context, namespace, subject, action, resource string) (bool, error)
	createFn func(ctx context.Context, namespace, object, relation, subject string) error
	deleteFn func(ctx context.Context, namespace, object, relation, subject string) error
}

func (m *mockKetoService) CheckPermission(ctx context.Context, namespace, subject, action, resource string) (bool, error) {
	return m.checkFn(ctx, namespace, subject, action, resource)
}

func (m *mockKetoService) CreateRelation(ctx context.Context, namespace, object, relation, subject string) error {
	return m.createFn(ctx, namespace, object, relation, subject)
}

func (m *mockKetoService) DeleteRelation(ctx context.Context, namespace, object, relation, subject string) error {
	return m.deleteFn(ctx, namespace, object, relation, subject)
}

// ---------------------------------------------------------------------------
// Assigner tests
// ---------------------------------------------------------------------------

func TestAssignFacultyToCourse_Success(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "courses", namespace)
			assert.Equal(t, "course-1", object)
			assert.Equal(t, "faculty-1", subject)
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	assigner.AssignFacultyToCourse(context.Background(), "faculty-1", "course-1")

	assert.Equal(t, []string{"faculty", "manage_qr", "view_attendance", "grade_assignments"}, created)
}

func TestAssignFacultyToCourse_FailureOnSecondRelation(t *testing.T) {
	callCount := 0
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, relation, _ string) error {
			callCount++
			if relation == "manage_qr" {
				return errors.New("keto error")
			}
			return nil
		},
	}

	assigner := NewAssigner(mock)
	// Should not panic; silently returns after first error.
	assigner.AssignFacultyToCourse(context.Background(), "faculty-1", "course-1")

	// "faculty" succeeds (call 1), "manage_qr" fails (call 2), then returns early.
	assert.Equal(t, 2, callCount)
}

func TestAssignStudentToCourse_Success(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "courses", namespace)
			assert.Equal(t, "course-2", object)
			assert.Equal(t, "student-1", subject)
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignStudentToCourse(context.Background(), "student-1", "course-2")

	require.NoError(t, err)
	assert.Equal(t, []string{"student", "mark_attendance", "submit_assignment"}, created)
}

func TestAssignStudentToCourse_Failure(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("keto error")
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignStudentToCourse(context.Background(), "student-1", "course-2")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign student to course")
}

func TestAssignDepartmentHead_Success(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "departments", namespace)
			assert.Equal(t, "dept-1", object)
			assert.Equal(t, "faculty-1", subject)
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignDepartmentHead(context.Background(), "faculty-1", "dept-1")

	require.NoError(t, err)
	assert.Equal(t, []string{"head", "manage_courses", "view_analytics"}, created)
}

func TestAssignDepartmentHead_Failure(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("keto error")
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignDepartmentHead(context.Background(), "faculty-1", "dept-1")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign department head")
}

func TestAssignResourcePermissions_Success(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "resources", namespace)
			assert.Equal(t, "res-1", object)
			assert.Equal(t, "user-1", subject)
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	perms := []string{"read", "write", "delete"}
	err := assigner.AssignResourcePermissions(context.Background(), "user-1", "res-1", perms)

	require.NoError(t, err)
	assert.Equal(t, perms, created)
}

func TestAssignResourcePermissions_FailureOnSecondPermission(t *testing.T) {
	callCount := 0
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, relation, _ string) error {
			callCount++
			if relation == "write" {
				return errors.New("keto error")
			}
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignResourcePermissions(context.Background(), "user-1", "res-1", []string{"read", "write", "delete"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign resource permission write")
	assert.Equal(t, 2, callCount)
}

func TestAssignAssignmentPermissions_CreatorRole(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "assignments", namespace)
			assert.Equal(t, "assign-1", object)
			assert.Equal(t, "user-1", subject)
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAssignmentPermissions(context.Background(), "user-1", "assign-1", "creator")

	require.NoError(t, err)
	assert.Equal(t, []string{"creator", "grader", "viewer"}, created)
}

func TestAssignAssignmentPermissions_StudentRole(t *testing.T) {
	var created []string
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, relation, _ string) error {
			created = append(created, relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAssignmentPermissions(context.Background(), "user-1", "assign-1", "student")

	require.NoError(t, err)
	assert.Equal(t, []string{"submitter", "viewer"}, created)
}

func TestAssignAssignmentPermissions_InvalidRole(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			t.Fatal("CreateRelation should not be called for invalid role")
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAssignmentPermissions(context.Background(), "user-1", "assign-1", "admin")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid assignment role: admin")
}

func TestAssignAssignmentPermissions_Failure(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("keto error")
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAssignmentPermissions(context.Background(), "user-1", "assign-1", "creator")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign assignment permission creator")
}

func TestAssignAnnouncementPermissions_Publisher(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, namespace, object, relation, subject string) error {
			assert.Equal(t, "announcements", namespace)
			assert.Equal(t, "ann-1", object)
			assert.Equal(t, "publisher", relation)
			assert.Equal(t, "user-1", subject)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAnnouncementPermissions(context.Background(), "user-1", "ann-1", true)

	require.NoError(t, err)
}

func TestAssignAnnouncementPermissions_Viewer(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, relation, _ string) error {
			assert.Equal(t, "viewer", relation)
			return nil
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAnnouncementPermissions(context.Background(), "user-1", "ann-1", false)

	require.NoError(t, err)
}

func TestAssignAnnouncementPermissions_Failure(t *testing.T) {
	mock := &mockKetoService{
		createFn: func(_ context.Context, _, _, _, _ string) error {
			return errors.New("keto error")
		},
	}

	assigner := NewAssigner(mock)
	err := assigner.AssignAnnouncementPermissions(context.Background(), "user-1", "ann-1", true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to assign announcement permission publisher")
}

// ---------------------------------------------------------------------------
// PermissionHelper tests
// ---------------------------------------------------------------------------

func TestCanManageQR_Allowed(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, namespace, subject, action, resource string) (bool, error) {
			assert.Equal(t, "courses", namespace)
			assert.Equal(t, "faculty-1", subject)
			assert.Equal(t, "manage_qr", action)
			assert.Equal(t, "course-1", resource)
			return true, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanManageQR(context.Background(), "faculty-1", "course-1")

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanManageQR_Denied(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, _, _, _, _ string) (bool, error) {
			return false, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanManageQR(context.Background(), "faculty-1", "course-1")

	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCanViewAttendance_Allowed(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, namespace, subject, action, resource string) (bool, error) {
			assert.Equal(t, "courses", namespace)
			assert.Equal(t, "user-1", subject)
			assert.Equal(t, "view_attendance", action)
			assert.Equal(t, "course-1", resource)
			return true, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanViewAttendance(context.Background(), "user-1", "course-1")

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanViewAttendance_Denied(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, _, _, _, _ string) (bool, error) {
			return false, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanViewAttendance(context.Background(), "user-1", "course-1")

	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCanManageCourses_Allowed(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, namespace, subject, action, resource string) (bool, error) {
			assert.Equal(t, "departments", namespace)
			assert.Equal(t, "faculty-1", subject)
			assert.Equal(t, "manage_courses", action)
			assert.Equal(t, "dept-1", resource)
			return true, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanManageCourses(context.Background(), "faculty-1", "dept-1")

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanManageCourses_Denied(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, _, _, _, _ string) (bool, error) {
			return false, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanManageCourses(context.Background(), "faculty-1", "dept-1")

	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCanAccessResource_Allowed(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, namespace, subject, action, resource string) (bool, error) {
			assert.Equal(t, "resources", namespace)
			assert.Equal(t, "user-1", subject)
			assert.Equal(t, "read", action)
			assert.Equal(t, "res-1", resource)
			return true, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanAccessResource(context.Background(), "user-1", "res-1", "read")

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestCanAccessResource_Denied(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, _, _, _, _ string) (bool, error) {
			return false, nil
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanAccessResource(context.Background(), "user-1", "res-1", "write")

	require.NoError(t, err)
	assert.False(t, allowed)
}

func TestCanAccessResource_Error(t *testing.T) {
	mock := &mockKetoService{
		checkFn: func(_ context.Context, _, _, _, _ string) (bool, error) {
			return false, errors.New("keto unavailable")
		},
	}

	ph := NewPermissionHelper(mock)
	allowed, err := ph.CanAccessResource(context.Background(), "user-1", "res-1", "read")

	require.Error(t, err)
	assert.False(t, allowed)
	assert.Contains(t, err.Error(), "keto unavailable")
}
