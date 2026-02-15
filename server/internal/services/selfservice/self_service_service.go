package selfservice

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

var (
	ErrRequestNotFound = errors.New("self-service request not found")
	ErrAccessDenied    = errors.New("access denied")
)

var validRequestTypes = map[string]struct{}{
	"enrollment": {},
	"schedule":   {},
	"transcript": {},
	"document":   {},
}

var validStatuses = map[string]struct{}{
	"approved":   {},
	"rejected":   {},
	"processing": {},
}

type RequestTypeOption struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SelfServiceService interface {
	ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error)
	ListStudentRequests(ctx context.Context, collegeID, studentID int) ([]*models.SelfServiceRequest, error)
	GetStudentRequest(ctx context.Context, collegeID, studentID, requestID int) (*models.SelfServiceRequest, error)
	CreateStudentRequest(ctx context.Context, req *models.SelfServiceRequest) error
	UpdateRequest(ctx context.Context, collegeID, requestID, responderUserID int, status, response string) (*models.SelfServiceRequest, error)
	RequestTypes() []RequestTypeOption
	DocumentTypes() []RequestTypeOption
	DeliveryMethods() []RequestTypeOption
}

type selfServiceService struct {
	repo repository.SelfServiceRepository
}

func NewSelfServiceService(repo repository.SelfServiceRepository) SelfServiceService {
	return &selfServiceService{repo: repo}
}

func (s *selfServiceService) ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error) {
	if strings.TrimSpace(kratosID) == "" {
		return 0, fmt.Errorf("kratos id is required")
	}
	return s.repo.ResolveUserIDByKratosID(ctx, kratosID)
}

func (s *selfServiceService) ListStudentRequests(ctx context.Context, collegeID, studentID int) ([]*models.SelfServiceRequest, error) {
	if collegeID <= 0 || studentID <= 0 {
		return nil, fmt.Errorf("invalid college or student id")
	}
	return s.repo.ListByStudent(ctx, collegeID, studentID)
}

func (s *selfServiceService) GetStudentRequest(ctx context.Context, collegeID, studentID, requestID int) (*models.SelfServiceRequest, error) {
	if collegeID <= 0 || studentID <= 0 || requestID <= 0 {
		return nil, fmt.Errorf("invalid request identifiers")
	}

	item, err := s.repo.GetByID(ctx, collegeID, requestID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrRequestNotFound
	}
	if item.StudentID != studentID {
		return nil, ErrAccessDenied
	}
	return item, nil
}

func (s *selfServiceService) CreateStudentRequest(ctx context.Context, req *models.SelfServiceRequest) error {
	if req == nil {
		return fmt.Errorf("request payload is required")
	}
	if req.StudentID <= 0 || req.CollegeID <= 0 {
		return fmt.Errorf("student_id and college_id are required")
	}

	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" || req.Description == "" {
		return fmt.Errorf("title and description are required")
	}
	if _, ok := validRequestTypes[req.Type]; !ok {
		return fmt.Errorf("invalid request type")
	}

	req.Status = "pending"
	if req.DocumentType != nil {
		trimmed := strings.TrimSpace(*req.DocumentType)
		req.DocumentType = &trimmed
	}
	if req.DeliveryMethod != nil {
		trimmed := strings.TrimSpace(*req.DeliveryMethod)
		req.DeliveryMethod = &trimmed
	}

	return s.repo.CreateRequest(ctx, req)
}

func (s *selfServiceService) UpdateRequest(ctx context.Context, collegeID, requestID, responderUserID int, status, response string) (*models.SelfServiceRequest, error) {
	if collegeID <= 0 || requestID <= 0 || responderUserID <= 0 {
		return nil, fmt.Errorf("invalid identifiers")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	response = strings.TrimSpace(response)
	if response == "" {
		return nil, fmt.Errorf("response is required")
	}
	if _, ok := validStatuses[status]; !ok {
		return nil, fmt.Errorf("invalid status")
	}

	item, err := s.repo.UpdateRequest(ctx, collegeID, requestID, status, response, responderUserID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrRequestNotFound
	}
	return item, nil
}

func (s *selfServiceService) RequestTypes() []RequestTypeOption {
	return []RequestTypeOption{
		{ID: "enrollment", Name: "Course Enrollment", Description: "Request to enroll in a course"},
		{ID: "schedule", Name: "Schedule Change", Description: "Request to change your class schedule"},
		{ID: "transcript", Name: "Official Transcript", Description: "Request official academic transcripts"},
		{ID: "document", Name: "Document Request", Description: "Request other academic documents"},
	}
}

func (s *selfServiceService) DocumentTypes() []RequestTypeOption {
	return []RequestTypeOption{
		{ID: "transcript", Name: "Official Transcript"},
		{ID: "certificate", Name: "Enrollment Certificate"},
		{ID: "id_card", Name: "ID Card Replacement"},
		{ID: "other", Name: "Other Document"},
	}
}

func (s *selfServiceService) DeliveryMethods() []RequestTypeOption {
	return []RequestTypeOption{
		{ID: "pickup", Name: "In-Person Pickup"},
		{ID: "email", Name: "Email (Digital Copy)"},
		{ID: "postal", Name: "Postal Mail"},
	}
}
