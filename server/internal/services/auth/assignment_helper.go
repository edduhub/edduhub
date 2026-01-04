package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type Assigner struct {
	keto KetoService
}

func NewAssigner(keto KetoService) *Assigner {
	return &Assigner{
		keto: keto,
	}
}

func (a *Assigner) AssignFacultyToCourse(ctx context.Context, facultyID, courseID string) {
	relations := []string{"faculty", "manage_qr", "view_attendance", "grade_assignments"}
	for _, relation := range relations {
		if err := a.keto.CreateRelation(ctx, "courses", courseID, relation, facultyID); err != nil {
			log.Printf("failed to assign relation")
			return
		}
	}
	return
}

func (a *Assigner) AssignStudentToCourse(ctx context.Context, studentID, courseID string) error {
	relations := []string{"student", "mark_attendance", "submit_assignment"}
	for _, relation := range relations {
		if err := a.keto.CreateRelation(ctx, "courses", courseID, relation, studentID); err != nil {
			return errors.New("failed to assign student to course")
		}
	}
	return nil
}

func (a *Assigner) AssignDepartmentHead(ctx context.Context, facultyID, deparmentID string) error {
	relations := []string{"head", "manage_courses", "view_analytics"}
	for _, relation := range relations {
		if err := a.keto.CreateRelation(ctx, "departments", deparmentID, relation, facultyID); err != nil {
			return errors.New("failed to assign department head")
		}
	}
	return nil
}

func (a *Assigner) AssignResourcePermissions(ctx context.Context, userID, resourceID string, permissions []string) error {
	for _, permission := range permissions {
		if err := a.keto.CreateRelation(ctx, "resources", resourceID, permission, userID); err != nil {
			return fmt.Errorf("failed to assign resource permission %s: %w", permission, err)
		}
	}
	return nil
}

// Assignment Relations
func (a *Assigner) AssignAssignmentPermissions(ctx context.Context, userID, assignmentID string, role string) error {
	var permissions []string
	switch role {
	case "creator":
		permissions = []string{"creator", "grader", "viewer"}
	case "student":
		permissions = []string{"submitter", "viewer"}
	default:
		return fmt.Errorf("invalid assignment role: %s", role)
	}

	for _, permission := range permissions {
		if err := a.keto.CreateRelation(ctx, "assignments", assignmentID, permission, userID); err != nil {
			return fmt.Errorf("failed to assign assignment permission %s: %w", permission, err)
		}
	}
	return nil
}

// Announcement Relations
func (a *Assigner) AssignAnnouncementPermissions(ctx context.Context, userID, announcementID string, isPublisher bool) error {
	permission := "viewer"
	if isPublisher {
		permission = "publisher"
	}

	if err := a.keto.CreateRelation(ctx, "announcements", announcementID, permission, userID); err != nil {
		return fmt.Errorf("failed to assign announcement permission %s: %w", permission, err)
	}
	return nil
}
