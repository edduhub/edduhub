package facultytools

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

var (
	ErrRubricNotFound          = errors.New("rubric not found")
	ErrOfficeHourNotFound      = errors.New("office hour not found")
	ErrBookingNotFound         = errors.New("booking not found")
	ErrFacultyToolsAccess      = errors.New("access denied")
	ErrInvalidStatusTransition = errors.New("invalid booking status transition")
)

var validBookingStatuses = map[string]struct{}{
	"confirmed": {},
	"cancelled": {},
	"completed": {},
	"no_show":   {},
}

type CreateRubricInput struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	CourseID    *int                     `json:"course_id,omitempty"`
	IsTemplate  bool                     `json:"is_template"`
	IsActive    bool                     `json:"is_active"`
	MaxScore    int                      `json:"max_score"`
	Criteria    []models.RubricCriterion `json:"criteria"`
}

type UpdateRubricInput struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	CourseID    *int                     `json:"course_id,omitempty"`
	IsTemplate  bool                     `json:"is_template"`
	IsActive    bool                     `json:"is_active"`
	MaxScore    int                      `json:"max_score"`
	Criteria    []models.RubricCriterion `json:"criteria"`
}

type CreateOfficeHourInput struct {
	DayOfWeek   int     `json:"day_of_week"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Location    *string `json:"location,omitempty"`
	IsVirtual   bool    `json:"is_virtual"`
	VirtualLink *string `json:"virtual_link,omitempty"`
	MaxStudents int     `json:"max_students"`
	IsActive    bool    `json:"is_active"`
}

type UpdateOfficeHourInput struct {
	DayOfWeek   int     `json:"day_of_week"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	Location    *string `json:"location,omitempty"`
	IsVirtual   bool    `json:"is_virtual"`
	VirtualLink *string `json:"virtual_link,omitempty"`
	MaxStudents int     `json:"max_students"`
	IsActive    bool    `json:"is_active"`
}

type CreateBookingInput struct {
	OfficeHourID int     `json:"office_hour_id"`
	BookingDate  string  `json:"booking_date"`
	StartTime    *string `json:"start_time,omitempty"`
	EndTime      *string `json:"end_time,omitempty"`
	Purpose      *string `json:"purpose,omitempty"`
}

type UpdateBookingStatusInput struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes,omitempty"`
}

type FacultyToolsService interface {
	ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error)

	CreateRubric(ctx context.Context, collegeID, facultyID int, input *CreateRubricInput) (*models.GradingRubric, error)
	UpdateRubric(ctx context.Context, collegeID, rubricID int, input *UpdateRubricInput) (*models.GradingRubric, error)
	DeleteRubric(ctx context.Context, collegeID, rubricID int) error
	GetRubric(ctx context.Context, collegeID, rubricID int) (*models.GradingRubric, error)
	ListRubrics(ctx context.Context, collegeID int, facultyID *int) ([]*models.GradingRubric, error)

	CreateOfficeHour(ctx context.Context, collegeID, facultyID int, input *CreateOfficeHourInput) (*models.OfficeHourSlot, error)
	UpdateOfficeHour(ctx context.Context, collegeID, officeHourID int, input *UpdateOfficeHourInput) (*models.OfficeHourSlot, error)
	DeleteOfficeHour(ctx context.Context, collegeID, officeHourID int) error
	GetOfficeHour(ctx context.Context, collegeID, officeHourID int) (*models.OfficeHourSlot, error)
	ListOfficeHours(ctx context.Context, collegeID int, facultyID *int, activeOnly bool) ([]*models.OfficeHourSlot, error)

	CreateBooking(ctx context.Context, collegeID, studentID int, input *CreateBookingInput) (*models.OfficeHourBooking, error)
	ListBookings(ctx context.Context, collegeID int, officeHourID, studentID, facultyID *int) ([]*models.OfficeHourBooking, error)
	UpdateBookingStatus(ctx context.Context, collegeID, bookingID int, role string, actorStudentID *int, input *UpdateBookingStatusInput) (*models.OfficeHourBooking, error)
}

type facultyToolsService struct {
	repo repository.FacultyToolsRepository
}

func NewFacultyToolsService(repo repository.FacultyToolsRepository) FacultyToolsService {
	return &facultyToolsService{repo: repo}
}

func (s *facultyToolsService) ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error) {
	if strings.TrimSpace(kratosID) == "" {
		return 0, fmt.Errorf("kratos id is required")
	}
	return s.repo.ResolveUserIDByKratosID(ctx, kratosID)
}

func (s *facultyToolsService) CreateRubric(ctx context.Context, collegeID, facultyID int, input *CreateRubricInput) (*models.GradingRubric, error) {
	if input == nil {
		return nil, fmt.Errorf("rubric input is required")
	}
	rubric := &models.GradingRubric{
		FacultyID:   facultyID,
		CollegeID:   collegeID,
		Name:        strings.TrimSpace(input.Name),
		Description: trimStringPtr(input.Description),
		CourseID:    input.CourseID,
		IsTemplate:  input.IsTemplate,
		IsActive:    input.IsActive,
		MaxScore:    input.MaxScore,
		Criteria:    sanitizeCriteria(input.Criteria),
	}
	if err := validateRubric(rubric); err != nil {
		return nil, err
	}
	if err := s.repo.CreateRubric(ctx, rubric); err != nil {
		return nil, err
	}
	return rubric, nil
}

func (s *facultyToolsService) UpdateRubric(ctx context.Context, collegeID, rubricID int, input *UpdateRubricInput) (*models.GradingRubric, error) {
	if input == nil {
		return nil, fmt.Errorf("rubric input is required")
	}
	rubric := &models.GradingRubric{
		ID:          rubricID,
		CollegeID:   collegeID,
		Name:        strings.TrimSpace(input.Name),
		Description: trimStringPtr(input.Description),
		CourseID:    input.CourseID,
		IsTemplate:  input.IsTemplate,
		IsActive:    input.IsActive,
		MaxScore:    input.MaxScore,
		Criteria:    sanitizeCriteria(input.Criteria),
	}
	if err := validateRubric(rubric); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateRubric(ctx, rubric); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, ErrRubricNotFound
		}
		return nil, err
	}
	return s.GetRubric(ctx, collegeID, rubricID)
}

func (s *facultyToolsService) DeleteRubric(ctx context.Context, collegeID, rubricID int) error {
	if err := s.repo.DeleteRubric(ctx, collegeID, rubricID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return ErrRubricNotFound
		}
		return err
	}
	return nil
}

func (s *facultyToolsService) GetRubric(ctx context.Context, collegeID, rubricID int) (*models.GradingRubric, error) {
	item, err := s.repo.GetRubricByID(ctx, collegeID, rubricID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrRubricNotFound
	}
	return item, nil
}

func (s *facultyToolsService) ListRubrics(ctx context.Context, collegeID int, facultyID *int) ([]*models.GradingRubric, error) {
	return s.repo.ListRubrics(ctx, collegeID, facultyID)
}

func (s *facultyToolsService) CreateOfficeHour(ctx context.Context, collegeID, facultyID int, input *CreateOfficeHourInput) (*models.OfficeHourSlot, error) {
	if input == nil {
		return nil, fmt.Errorf("office hour input is required")
	}
	slot := &models.OfficeHourSlot{
		FacultyID:   facultyID,
		CollegeID:   collegeID,
		DayOfWeek:   input.DayOfWeek,
		StartTime:   strings.TrimSpace(input.StartTime),
		EndTime:     strings.TrimSpace(input.EndTime),
		Location:    trimStringPtr(input.Location),
		IsVirtual:   input.IsVirtual,
		VirtualLink: trimStringPtr(input.VirtualLink),
		MaxStudents: input.MaxStudents,
		IsActive:    input.IsActive,
	}
	if err := validateOfficeHourSlot(slot); err != nil {
		return nil, err
	}
	if err := s.repo.CreateOfficeHour(ctx, slot); err != nil {
		return nil, err
	}
	return slot, nil
}

func (s *facultyToolsService) UpdateOfficeHour(ctx context.Context, collegeID, officeHourID int, input *UpdateOfficeHourInput) (*models.OfficeHourSlot, error) {
	if input == nil {
		return nil, fmt.Errorf("office hour input is required")
	}
	slot := &models.OfficeHourSlot{
		ID:          officeHourID,
		CollegeID:   collegeID,
		DayOfWeek:   input.DayOfWeek,
		StartTime:   strings.TrimSpace(input.StartTime),
		EndTime:     strings.TrimSpace(input.EndTime),
		Location:    trimStringPtr(input.Location),
		IsVirtual:   input.IsVirtual,
		VirtualLink: trimStringPtr(input.VirtualLink),
		MaxStudents: input.MaxStudents,
		IsActive:    input.IsActive,
	}
	if err := validateOfficeHourSlot(slot); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateOfficeHour(ctx, slot); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, ErrOfficeHourNotFound
		}
		return nil, err
	}
	return s.GetOfficeHour(ctx, collegeID, officeHourID)
}

func (s *facultyToolsService) DeleteOfficeHour(ctx context.Context, collegeID, officeHourID int) error {
	if err := s.repo.DeleteOfficeHour(ctx, collegeID, officeHourID); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return ErrOfficeHourNotFound
		}
		return err
	}
	return nil
}

func (s *facultyToolsService) GetOfficeHour(ctx context.Context, collegeID, officeHourID int) (*models.OfficeHourSlot, error) {
	slot, err := s.repo.GetOfficeHourByID(ctx, collegeID, officeHourID)
	if err != nil {
		return nil, err
	}
	if slot == nil {
		return nil, ErrOfficeHourNotFound
	}
	return slot, nil
}

func (s *facultyToolsService) ListOfficeHours(ctx context.Context, collegeID int, facultyID *int, activeOnly bool) ([]*models.OfficeHourSlot, error) {
	return s.repo.ListOfficeHours(ctx, collegeID, facultyID, activeOnly)
}

func (s *facultyToolsService) CreateBooking(ctx context.Context, collegeID, studentID int, input *CreateBookingInput) (*models.OfficeHourBooking, error) {
	if input == nil {
		return nil, fmt.Errorf("booking input is required")
	}
	if studentID <= 0 || collegeID <= 0 {
		return nil, fmt.Errorf("invalid student or college id")
	}
	if input.OfficeHourID <= 0 {
		return nil, fmt.Errorf("office_hour_id is required")
	}

	officeHour, err := s.repo.GetOfficeHourByID(ctx, collegeID, input.OfficeHourID)
	if err != nil {
		return nil, err
	}
	if officeHour == nil {
		return nil, ErrOfficeHourNotFound
	}
	if !officeHour.IsActive {
		return nil, fmt.Errorf("office hour slot is inactive")
	}

	bookingDate, err := time.Parse("2006-01-02", strings.TrimSpace(input.BookingDate))
	if err != nil {
		return nil, fmt.Errorf("booking_date must be YYYY-MM-DD")
	}

	startTime := officeHour.StartTime
	if input.StartTime != nil && strings.TrimSpace(*input.StartTime) != "" {
		startTime, err = normalizeTime(*input.StartTime)
		if err != nil {
			return nil, err
		}
	}
	endTime := officeHour.EndTime
	if input.EndTime != nil && strings.TrimSpace(*input.EndTime) != "" {
		endTime, err = normalizeTime(*input.EndTime)
		if err != nil {
			return nil, err
		}
	}
	if startTime >= endTime {
		return nil, fmt.Errorf("start_time must be before end_time")
	}

	existing, err := s.repo.ListBookings(ctx, collegeID, &input.OfficeHourID, nil, nil)
	if err != nil {
		return nil, err
	}
	activeCount := 0
	for _, item := range existing {
		if item.BookingDate.Format("2006-01-02") != bookingDate.Format("2006-01-02") {
			continue
		}
		if item.StartTime == startTime && item.StudentID == studentID && item.Status != "cancelled" {
			return nil, fmt.Errorf("booking already exists for this slot")
		}
		if item.StartTime == startTime && item.EndTime == endTime && item.Status != "cancelled" {
			activeCount++
		}
	}
	if officeHour.MaxStudents > 0 && activeCount >= officeHour.MaxStudents {
		return nil, fmt.Errorf("office hour slot is fully booked")
	}

	booking := &models.OfficeHourBooking{
		OfficeHourID: input.OfficeHourID,
		StudentID:    studentID,
		CollegeID:    collegeID,
		BookingDate:  bookingDate,
		StartTime:    startTime,
		EndTime:      endTime,
		Purpose:      trimStringPtr(input.Purpose),
		Status:       "confirmed",
	}
	if err := s.repo.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}
	return s.repo.GetBookingByID(ctx, collegeID, booking.ID)
}

func (s *facultyToolsService) ListBookings(ctx context.Context, collegeID int, officeHourID, studentID, facultyID *int) ([]*models.OfficeHourBooking, error) {
	return s.repo.ListBookings(ctx, collegeID, officeHourID, studentID, facultyID)
}

func (s *facultyToolsService) UpdateBookingStatus(ctx context.Context, collegeID, bookingID int, role string, actorStudentID *int, input *UpdateBookingStatusInput) (*models.OfficeHourBooking, error) {
	if input == nil {
		return nil, fmt.Errorf("status payload is required")
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if _, ok := validBookingStatuses[status]; !ok {
		return nil, fmt.Errorf("invalid status")
	}

	booking, err := s.repo.GetBookingByID(ctx, collegeID, bookingID)
	if err != nil {
		return nil, err
	}
	if booking == nil {
		return nil, ErrBookingNotFound
	}

	if role == "student" {
		if actorStudentID == nil || booking.StudentID != *actorStudentID {
			return nil, ErrFacultyToolsAccess
		}
		if status != "cancelled" {
			return nil, ErrInvalidStatusTransition
		}
	}

	notes := trimStringPtr(input.Notes)
	updated, err := s.repo.UpdateBookingStatus(ctx, collegeID, bookingID, status, notes)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrBookingNotFound
	}
	return updated, nil
}

func validateRubric(rubric *models.GradingRubric) error {
	if rubric.Name == "" {
		return fmt.Errorf("name is required")
	}
	if rubric.MaxScore <= 0 {
		return fmt.Errorf("max_score must be greater than 0")
	}
	if len(rubric.Criteria) == 0 {
		return fmt.Errorf("at least one criterion is required")
	}
	weightTotal := 0.0
	for _, c := range rubric.Criteria {
		if strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("criterion name is required")
		}
		if c.MaxScore <= 0 {
			return fmt.Errorf("criterion max_score must be greater than 0")
		}
		if c.Weight <= 0 {
			return fmt.Errorf("criterion weight must be greater than 0")
		}
		weightTotal += c.Weight
	}
	if weightTotal > 100.01 {
		return fmt.Errorf("total criteria weight cannot exceed 100")
	}
	return nil
}

func validateOfficeHourSlot(slot *models.OfficeHourSlot) error {
	if slot.DayOfWeek < 0 || slot.DayOfWeek > 6 {
		return fmt.Errorf("day_of_week must be between 0 and 6")
	}
	var err error
	slot.StartTime, err = normalizeTime(slot.StartTime)
	if err != nil {
		return err
	}
	slot.EndTime, err = normalizeTime(slot.EndTime)
	if err != nil {
		return err
	}
	if slot.StartTime >= slot.EndTime {
		return fmt.Errorf("start_time must be before end_time")
	}
	if slot.MaxStudents <= 0 {
		slot.MaxStudents = 1
	}
	if slot.IsVirtual {
		if slot.VirtualLink == nil || strings.TrimSpace(*slot.VirtualLink) == "" {
			return fmt.Errorf("virtual_link is required for virtual slots")
		}
	} else if slot.Location == nil || strings.TrimSpace(*slot.Location) == "" {
		return fmt.Errorf("location is required for in-person slots")
	}
	return nil
}

func normalizeTime(in string) (string, error) {
	trimmed := strings.TrimSpace(in)
	if trimmed == "" {
		return "", fmt.Errorf("time is required")
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.Format("15:04"), nil
		}
	}
	return "", fmt.Errorf("invalid time format, expected HH:MM")
}

func trimStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*in)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func sanitizeCriteria(criteria []models.RubricCriterion) []models.RubricCriterion {
	clean := make([]models.RubricCriterion, 0, len(criteria))
	for idx, criterion := range criteria {
		criterion.Name = strings.TrimSpace(criterion.Name)
		criterion.Description = trimStringPtr(criterion.Description)
		if criterion.SortOrder <= 0 {
			criterion.SortOrder = idx + 1
		}
		clean = append(clean, criterion)
	}
	return clean
}
