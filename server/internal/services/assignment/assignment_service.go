package assignment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/storage"
)

type AssignmentService interface {
	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	GetAssignment(ctx context.Context, collegeID, assignmentID int) (*models.Assignment, error)
	GetAssignmentsByCourse(ctx context.Context, collegeID, courseID int) ([]*models.Assignment, error)
	GetAssignmentsByStudent(ctx context.Context, collegeID, studentID int) ([]*models.Assignment, error)
	UpdateAssignment(ctx context.Context, collegeID, assignmentID int, req *models.UpdateAssignmentRequest) error
	DeleteAssignment(ctx context.Context, collegeID, assignmentID int) error
	SubmitAssignment(ctx context.Context, submission *models.AssignmentSubmission) error
	GradeSubmission(ctx context.Context, collegeID, submissionID int, grade *int, feedback *string) error
	GetSubmissionByStudentAndAssignment(ctx context.Context, studentID, assignmentID int) (*models.AssignmentSubmission, error)
	CountPendingSubmissionsByCollege(ctx context.Context, collegeID int) (int, error)

	// Enhanced grading features
	BulkGradeSubmissions(ctx context.Context, collegeID int, grades map[int]*GradeInput) error
	GetSubmissionsByAssignment(ctx context.Context, collegeID, assignmentID int) ([]*models.AssignmentSubmission, error)
	CalculateLatePenalty(submission *models.AssignmentSubmission, assignment *models.Assignment) int
	GetGradingStats(ctx context.Context, collegeID, assignmentID int) (*GradingStats, error)
}

// GradeInput represents grading input for a submission
type GradeInput struct {
	Grade    *int
	Feedback *string
}

// GradingStats represents statistics for assignment grading
type GradingStats struct {
	TotalSubmissions  int
	GradedSubmissions int
	PendingGrading    int
	AverageGrade      float64
	LateSubmissions   int
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

func (a *assignmentService) GetAssignmentsByStudent(ctx context.Context, collegeID, studentID int) ([]*models.Assignment, error) {
	// Get all assignments for courses the student is enrolled in
	// First, get enrollments to find course IDs
	return a.repo.FindAssignmentsByStudent(ctx, collegeID, studentID)
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

func (a *assignmentService) GetSubmissionByStudentAndAssignment(ctx context.Context, studentID, assignmentID int) (*models.AssignmentSubmission, error) {
	if studentID == 0 || assignmentID == 0 {
		return nil, errors.New("studentID and assignmentID are required")
	}
	return a.repo.GetSubmissionByStudentAndAssignment(ctx, studentID, assignmentID)
}

func (a *assignmentService) CountPendingSubmissionsByCollege(ctx context.Context, collegeID int) (int, error) {
	if collegeID == 0 {
		return 0, errors.New("collegeID is required")
	}
	return a.repo.CountPendingSubmissionsByCollege(ctx, collegeID)
}

// BulkGradeSubmissions grades multiple submissions at once
func (a *assignmentService) BulkGradeSubmissions(ctx context.Context, collegeID int, grades map[int]*GradeInput) error {
	var errs []error
	for submissionID, gradeInput := range grades {
		if err := a.GradeSubmission(ctx, collegeID, submissionID, gradeInput.Grade, gradeInput.Feedback); err != nil {
			errs = append(errs, fmt.Errorf("failed to grade submission %d: %w", submissionID, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("bulk grading encountered %d errors: %v", len(errs), errs)
	}
	return nil
}

// GetSubmissionsByAssignment retrieves all submissions for an assignment
func (a *assignmentService) GetSubmissionsByAssignment(ctx context.Context, collegeID, assignmentID int) ([]*models.AssignmentSubmission, error) {
	return a.repo.FindSubmissionsByAssignment(ctx, assignmentID, uint64(1000), uint64(0))
}

// CalculateLatePenalty calculates grade penalty for late submissions
func (a *assignmentService) CalculateLatePenalty(submission *models.AssignmentSubmission, assignment *models.Assignment) int {
	if submission.SubmissionTime.Before(assignment.DueDate) {
		return 0 // Not late
	}

	// Calculate hours late
	hoursLate := int(submission.SubmissionTime.Sub(assignment.DueDate).Hours())

	// Default penalty: 10% per day, max 50%
	daysLate := (hoursLate / 24) + 1
	penalty := daysLate * 10
	if penalty > 50 {
		penalty = 50
	}

	return penalty
}

func (a *assignmentService) GetGradingStats(ctx context.Context, collegeID, assignmentID int) (*GradingStats, error) {
	submissions, err := a.GetSubmissionsByAssignment(ctx, collegeID, assignmentID)
	if err != nil {
		return nil, err
	}

	assignment, err := a.GetAssignment(ctx, collegeID, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment for grading stats: %w", err)
	}

	stats := &GradingStats{
		TotalSubmissions: len(submissions),
	}

	totalGrades := 0

	for _, sub := range submissions {
		if sub.Grade != nil {
			stats.GradedSubmissions++
			totalGrades += *sub.Grade
		} else {
			stats.PendingGrading++
		}

		if sub.SubmissionTime.After(assignment.DueDate) {
			stats.LateSubmissions++
		}
	}

	if stats.GradedSubmissions > 0 {
		stats.AverageGrade = float64(totalGrades) / float64(stats.GradedSubmissions)
	}

	return stats, nil
}
