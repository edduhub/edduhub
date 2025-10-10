package assignment

import (
	"context"
	"errors"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/storage"
)

type AssignmentService interface {
	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	GetAssignment(ctx context.Context, collegeID, assignmentID int) (*models.Assignment, error)
	GetAssignmentsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.Assignment, error)
	UpdateAssignment(ctx context.Context, collegeID, assignmentID int, req *models.UpdateAssignmentRequest) error
	DeleteAssignment(ctx context.Context, collegeID, assignmentID int) error
	SubmitAssignment(ctx context.Context, submission *models.AssignmentSubmission) error
	GradeSubmission(ctx context.Context, collegeID, submissionID int, grade *int, feedback *string) error
}

type assignmentService struct {
	repo        repository.AssignmentRepository
	minioClient *storage.MinioClient
}

func NewAssignmentService(repo repository.AssignmentRepository, minioClient *storage.MinioClient) *assignmentService {
	return &assignmentService{
		repo:        repo,
		minioClient: minioClient,
	}
}

func (a *assignmentService) CreateAssignment(ctx context.Context, assignment *models.Assignment) error {
	if assignment.Title == "" {
		return errors.New("assignment title cannot be empty")
	}
	if assignment.CourseID == 0 || assignment.CollegeID == 0 {
		return errors.New("courseID or collegeID cannot be 0")
	}
	return a.repo.CreateAssignment(ctx, assignment)
}
func (a *assignmentService) GetAssignment(ctx context.Context, collegeID, assignmentID int) (*models.Assignment, error) {
	if collegeID == 0 || assignmentID == 0 {
		return &models.Assignment{}, errors.New("invalid collegeID or assignmentID")
	}
	return a.repo.GetAssignmentByID(ctx, collegeID, assignmentID)

}

func (a *assignmentService) GetAssignmentsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.Assignment, error) {
	return a.repo.FindAssignmentsByCourse(ctx, collegeID, courseID, 100, 0)
}

func (a *assignmentService) UpdateAssignment(ctx context.Context, collegeID, assignmentID int, req *models.UpdateAssignmentRequest) error {
	if collegeID == 0 || assignmentID == 0 {
		return errors.New("invalid collegeID or assignmentID")
	}
	return a.repo.UpdateAssignmentPartial(ctx, collegeID, req)
}

func (a *assignmentService) DeleteAssignment(ctx context.Context, collegeID, assignmentID int) error {
	return a.repo.DeleteAssignment(ctx, collegeID, assignmentID)
}

func (a *assignmentService) SubmitAssignment(ctx context.Context, submission *models.AssignmentSubmission) error {
	if submission.AssignmentID == 0 || submission.StudentID == 0 {
		return errors.New("assignmentID and studentID are required")
	}
	submission.SubmissionTime = time.Now()
	return a.repo.CreateSubmission(ctx, submission)
}

func (a *assignmentService) GradeSubmission(ctx context.Context, collegeID, submissionID int, grade *int, feedback *string) error {
	submission, err := a.repo.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return err
	}
	
	if grade != nil {
		submission.Grade = grade
	}
	if feedback != nil {
		submission.Feedback = feedback
	}
	
	return a.repo.UpdateSubmission(ctx, submission)
}
