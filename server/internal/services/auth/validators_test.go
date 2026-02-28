package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNamespaceAndRelation(t *testing.T) {
	t.Run("valid course relations", func(t *testing.T) {
		validRelations := []string{
			"faculty", "student", "admin", "manage_qr",
			"view_attendance", "mark_attendance",
			"manage_assignments", "submit_assignments", "grade_assignments",
		}
		for _, rel := range validRelations {
			err := validateNamespaceAndRelation("courses", rel)
			assert.NoError(t, err, "expected relation %q to be valid for courses", rel)
		}
	})

	t.Run("valid department relations", func(t *testing.T) {
		validRelations := []string{"head", "faculty_member", "manage_courses", "view_analytics"}
		for _, rel := range validRelations {
			err := validateNamespaceAndRelation("departments", rel)
			assert.NoError(t, err, "expected relation %q to be valid for departments", rel)
		}
	})

	t.Run("valid resource relations", func(t *testing.T) {
		validRelations := []string{"owner", "viewer", "editor", "uploader", "download"}
		for _, rel := range validRelations {
			err := validateNamespaceAndRelation("resources", rel)
			assert.NoError(t, err, "expected relation %q to be valid for resources", rel)
		}
	})

	t.Run("valid assignment relations", func(t *testing.T) {
		validRelations := []string{"creator", "submitter", "grader", "viewer"}
		for _, rel := range validRelations {
			err := validateNamespaceAndRelation("assignments", rel)
			assert.NoError(t, err, "expected relation %q to be valid for assignments", rel)
		}
	})

	t.Run("valid announcement relations", func(t *testing.T) {
		validRelations := []string{"publisher", "viewer", "manager"}
		for _, rel := range validRelations {
			err := validateNamespaceAndRelation("announcements", rel)
			assert.NoError(t, err, "expected relation %q to be valid for announcements", rel)
		}
	})

	t.Run("invalid namespace", func(t *testing.T) {
		err := validateNamespaceAndRelation("unknown", "viewer")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid namespace")
	})

	t.Run("invalid relation for valid namespace", func(t *testing.T) {
		err := validateNamespaceAndRelation("courses", "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid relation")
	})

	t.Run("empty namespace", func(t *testing.T) {
		err := validateNamespaceAndRelation("", "viewer")
		require.Error(t, err)
	})

	t.Run("empty relation", func(t *testing.T) {
		err := validateNamespaceAndRelation("courses", "")
		require.Error(t, err)
	})

	t.Run("cross-namespace relation is invalid", func(t *testing.T) {
		// "head" is a department relation, not a course relation
		err := validateNamespaceAndRelation("courses", "head")
		require.Error(t, err)
	})
}
