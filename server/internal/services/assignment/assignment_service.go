package assignment

import (
	"context"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/storage"
	"errors"
	"fmt"
	"io"
	"time"
)

type AssignmentService interface {
	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	GetAssignment(ctx context.Context, collegeID, assignmentID int) (*models.Assignment, error)
	UpdateAssignment(ctx context.Context, collegeID, assignment *models.Assignment) error
	DeleteAssignment(ctx context.Context, collegeID, assignmentID int) error
	ListAssignmentsByCourse(ctx context.Context, collegeID, courseID int, limit, offset uint64) ([]*models.Assignment, error)
	CountAssignments(ctx context.Context, collegeID int) (int, error)

	CreateSubmission(ctx context.Context, submission *models.AssignmentSubmission, file io.Reader, filePath string) error
	
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

func (a *assignmentService) CreateSubmission(ctx context.Context, submission *models.AssignmentSubmission, file io.Reader, fileName string) error {
	if submission.AssignmentID == 0 || submission.StudentID == 0 {
		return errors.New("assignmentID and studentID are required ")
	}
	submission.SubmissionTime = time.Now()
	bucket := a.minioClient.GetBucketName()

	fileSize, err := a.minioClient.GetFileSize(ctx, bucket, fileName)
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}

	err = a.minioClient.UploadFromReader(ctx, file, bucket, fileName, int64(fileSize), "application/octet-stream")
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	return nil
}
