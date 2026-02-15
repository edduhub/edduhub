package models

import "testing"

func TestForumCategoryIsValid(t *testing.T) {
	valid := []ForumCategory{
		CategoryGeneral,
		CategoryAcademic,
		CategoryAssignment,
		CategoryQuestion,
		CategoryAnnouncement,
	}
	for _, category := range valid {
		if !category.IsValid() {
			t.Fatalf("expected category %q to be valid", category)
		}
	}

	invalid := []ForumCategory{"", "unknown", "General"}
	for _, category := range invalid {
		if category.IsValid() {
			t.Fatalf("expected category %q to be invalid", category)
		}
	}
}
