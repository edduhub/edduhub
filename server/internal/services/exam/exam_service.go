package exam

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type ExamService interface {
	// Exam Management
	CreateExam(ctx context.Context, exam *models.Exam) error
	GetExam(ctx context.Context, collegeID, examID int) (*models.Exam, error)
	ListExams(ctx context.Context, collegeID int, filters map[string]interface{}, limit, offset int) ([]*models.Exam, error)
	ListExamsByCourse(ctx context.Context, collegeID, courseID int, limit, offset int) ([]*models.Exam, error)
	UpdateExam(ctx context.Context, exam *models.Exam) error
	DeleteExam(ctx context.Context, collegeID, examID int) error
	GetExamStats(ctx context.Context, collegeID, examID int) (*ExamStats, error)

	// Enrollment Management
	EnrollStudent(ctx context.Context, enrollment *models.ExamEnrollment) error
	EnrollMultipleStudents(ctx context.Context, examID, collegeID int, studentIDs []int) error
	GetEnrollment(ctx context.Context, examID, studentID int) (*models.ExamEnrollment, error)
	ListEnrollments(ctx context.Context, examID int) ([]*models.ExamEnrollment, error)
	UpdateEnrollment(ctx context.Context, enrollment *models.ExamEnrollment) error
	DeleteEnrollment(ctx context.Context, examID, studentID int) error
	GetStudentEnrollments(ctx context.Context, studentID, collegeID int) ([]*models.ExamEnrollment, error)

	// Seat Allocation
	AllocateSeats(ctx context.Context, examID int) error
	GenerateHallTicket(ctx context.Context, examID, studentID int) (*models.HallTicketResponse, error)
	GenerateAllHallTickets(ctx context.Context, examID int) error

	// Result Management
	CreateResult(ctx context.Context, result *models.ExamResult) error
	GetResult(ctx context.Context, examID, studentID int) (*models.ExamResult, error)
	ListResults(ctx context.Context, examID int) ([]*models.ExamResult, error)
	UpdateResult(ctx context.Context, result *models.ExamResult) error
	GetStudentResults(ctx context.Context, studentID, collegeID int) ([]*models.ExamResult, error)
	BulkGradeResults(ctx context.Context, examID int, results map[int]*ResultInput) error
	CalculateGrade(marks, totalMarks float64) string
	GetResultStats(ctx context.Context, examID int) (*ResultStats, error)

	// Revaluation Management
	CreateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error
	GetRevaluationRequest(ctx context.Context, requestID int) (*models.RevaluationRequest, error)
	ListRevaluationRequests(ctx context.Context, collegeID int, filters map[string]interface{}) ([]*models.RevaluationRequest, error)
	UpdateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error
	ApproveRevaluationRequest(ctx context.Context, requestID int, reviewedBy int, revisedMarks float64, comments string) error
	RejectRevaluationRequest(ctx context.Context, requestID int, reviewedBy int, comments string) error

	// Room Management
	CreateRoom(ctx context.Context, room *models.ExamRoom) error
	GetRoom(ctx context.Context, collegeID, roomID int) (*models.ExamRoom, error)
	ListRooms(ctx context.Context, collegeID int, activeOnly bool) ([]*models.ExamRoom, error)
	UpdateRoom(ctx context.Context, room *models.ExamRoom) error
	DeleteRoom(ctx context.Context, collegeID, roomID int) error
	CheckRoomAvailability(ctx context.Context, roomID int, startTime, endTime string) (bool, error)
}

// ResultInput represents input for grading an exam
type ResultInput struct {
	MarksObtained float64
	Remarks       string
}

// ExamStats represents statistics for an exam
type ExamStats struct {
	TotalEnrolled    int
	Appeared         int
	Absent           int
	ResultsPublished int
	AverageMarks     float64
	PassRate         float64
}

// ResultStats represents statistics for exam results
type ResultStats struct {
	TotalStudents  int
	Passed         int
	Failed         int
	Absent         int
	PassPercentage float64
	AverageMarks   float64
	HighestMarks   float64
	LowestMarks    float64
}

type examService struct {
	repo        repository.ExamRepository
	studentRepo repository.StudentRepository
	courseRepo  repository.CourseRepository
}

func NewExamService(
	repo repository.ExamRepository,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
) ExamService {
	return &examService{
		repo:        repo,
		studentRepo: studentRepo,
		courseRepo:  courseRepo,
	}
}

// ===========================
// Exam Management
// ===========================

func (s *examService) CreateExam(ctx context.Context, exam *models.Exam) error {
	// Validation
	if exam.Title == "" {
		return errors.New("exam title is required")
	}
	if exam.CourseID == 0 {
		return errors.New("course ID is required")
	}
	if exam.CollegeID == 0 {
		return errors.New("college ID is required")
	}
	if exam.StartTime.After(exam.EndTime) {
		return errors.New("start time must be before end time")
	}
	if exam.Duration <= 0 {
		return errors.New("duration must be positive")
	}
	if exam.TotalMarks <= 0 {
		return errors.New("total marks must be positive")
	}
	if exam.PassingMarks < 0 || exam.PassingMarks > exam.TotalMarks {
		return errors.New("passing marks must be between 0 and total marks")
	}

	// Set default status if not provided
	if exam.Status == "" {
		exam.Status = "scheduled"
	}

	return s.repo.CreateExam(ctx, exam)
}

func (s *examService) GetExam(ctx context.Context, collegeID, examID int) (*models.Exam, error) {
	if collegeID == 0 || examID == 0 {
		return nil, errors.New("invalid college ID or exam ID")
	}
	return s.repo.GetExamByID(ctx, collegeID, examID)
}

func (s *examService) ListExams(ctx context.Context, collegeID int, filters map[string]interface{}, limit, offset int) ([]*models.Exam, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListExams(ctx, collegeID, filters, limit, offset)
}

func (s *examService) ListExamsByCourse(ctx context.Context, collegeID, courseID int, limit, offset int) ([]*models.Exam, error) {
	if collegeID == 0 || courseID == 0 {
		return nil, errors.New("invalid college ID or course ID")
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListExamsByCourse(ctx, collegeID, courseID, limit, offset)
}

func (s *examService) UpdateExam(ctx context.Context, exam *models.Exam) error {
	if exam.ID == 0 || exam.CollegeID == 0 {
		return errors.New("invalid exam ID or college ID")
	}

	// Validate if exam exists
	_, err := s.repo.GetExamByID(ctx, exam.CollegeID, exam.ID)
	if err != nil {
		return fmt.Errorf("exam not found: %w", err)
	}

	// Validation
	if exam.Title != "" && exam.StartTime.After(exam.EndTime) {
		return errors.New("start time must be before end time")
	}
	if exam.TotalMarks > 0 && exam.PassingMarks > exam.TotalMarks {
		return errors.New("passing marks cannot exceed total marks")
	}

	return s.repo.UpdateExam(ctx, exam)
}

func (s *examService) DeleteExam(ctx context.Context, collegeID, examID int) error {
	if collegeID == 0 || examID == 0 {
		return errors.New("invalid college ID or exam ID")
	}
	return s.repo.DeleteExam(ctx, collegeID, examID)
}

func (s *examService) GetExamStats(ctx context.Context, collegeID, examID int) (*ExamStats, error) {
	enrollments, err := s.repo.ListEnrollments(ctx, examID)
	if err != nil {
		return nil, err
	}

	results, err := s.repo.ListResults(ctx, examID)
	if err != nil {
		return nil, err
	}

	exam, err := s.repo.GetExamByID(ctx, collegeID, examID)
	if err != nil {
		return nil, err
	}

	stats := &ExamStats{
		TotalEnrolled: len(enrollments),
	}

	// Count status
	for _, enrollment := range enrollments {
		if enrollment.Status == "appeared" {
			stats.Appeared++
		} else if enrollment.Status == "absent" {
			stats.Absent++
		}
	}

	// Calculate result stats
	var totalMarks float64
	passCount := 0
	for _, result := range results {
		if result.MarksObtained != nil {
			stats.ResultsPublished++
			totalMarks += *result.MarksObtained
			if *result.MarksObtained >= exam.PassingMarks {
				passCount++
			}
		}
	}

	if stats.ResultsPublished > 0 {
		stats.AverageMarks = totalMarks / float64(stats.ResultsPublished)
		stats.PassRate = float64(passCount) / float64(stats.ResultsPublished) * 100
	}

	return stats, nil
}

// ===========================
// Enrollment Management
// ===========================

func (s *examService) EnrollStudent(ctx context.Context, enrollment *models.ExamEnrollment) error {
	if enrollment.ExamID == 0 || enrollment.StudentID == 0 {
		return errors.New("exam ID and student ID are required")
	}
	if enrollment.CollegeID == 0 {
		return errors.New("college ID is required")
	}

	// Check if already enrolled
	existing, _ := s.repo.GetEnrollment(ctx, enrollment.ExamID, enrollment.StudentID)
	if existing != nil {
		return errors.New("student already enrolled in this exam")
	}

	// Set default status
	if enrollment.Status == "" {
		enrollment.Status = "enrolled"
	}

	return s.repo.EnrollStudent(ctx, enrollment)
}

func (s *examService) EnrollMultipleStudents(ctx context.Context, examID, collegeID int, studentIDs []int) error {
	if examID == 0 || collegeID == 0 {
		return errors.New("exam ID and college ID are required")
	}
	if len(studentIDs) == 0 {
		return errors.New("no students provided")
	}

	for _, studentID := range studentIDs {
		enrollment := &models.ExamEnrollment{
			ExamID:    examID,
			StudentID: studentID,
			CollegeID: collegeID,
			Status:    "enrolled",
		}
		// Continue on error to enroll as many as possible
		_ = s.repo.EnrollStudent(ctx, enrollment)
	}

	return nil
}

func (s *examService) GetEnrollment(ctx context.Context, examID, studentID int) (*models.ExamEnrollment, error) {
	if examID == 0 || studentID == 0 {
		return nil, errors.New("exam ID and student ID are required")
	}
	return s.repo.GetEnrollment(ctx, examID, studentID)
}

func (s *examService) ListEnrollments(ctx context.Context, examID int) ([]*models.ExamEnrollment, error) {
	if examID == 0 {
		return nil, errors.New("exam ID is required")
	}
	return s.repo.ListEnrollments(ctx, examID)
}

func (s *examService) UpdateEnrollment(ctx context.Context, enrollment *models.ExamEnrollment) error {
	if enrollment.ID == 0 {
		return errors.New("enrollment ID is required")
	}
	return s.repo.UpdateEnrollment(ctx, enrollment)
}

func (s *examService) DeleteEnrollment(ctx context.Context, examID, studentID int) error {
	if examID == 0 || studentID == 0 {
		return errors.New("exam ID and student ID are required")
	}
	return s.repo.DeleteEnrollment(ctx, examID, studentID)
}

func (s *examService) GetStudentEnrollments(ctx context.Context, studentID, collegeID int) ([]*models.ExamEnrollment, error) {
	if studentID == 0 || collegeID == 0 {
		return nil, errors.New("student ID and college ID are required")
	}
	return s.repo.GetStudentEnrollments(ctx, studentID, collegeID)
}

// ===========================
// Seat Allocation
// ===========================

func (s *examService) AllocateSeats(ctx context.Context, examID int) error {
	enrollments, err := s.repo.ListEnrollments(ctx, examID)
	if err != nil {
		return err
	}

	// Simple sequential seat allocation
	for i, enrollment := range enrollments {
		seatNum := fmt.Sprintf("S%03d", i+1)
		enrollment.SeatNumber = &seatNum

		// Assign question paper set (cycle through available sets)
		// Assuming exam has QuestionPaperSets field
		if err := s.repo.UpdateEnrollment(ctx, enrollment); err != nil {
			return err
		}
	}

	return nil
}

func (s *examService) GenerateHallTicket(ctx context.Context, examID, studentID int) (*models.HallTicketResponse, error) {
	enrollment, err := s.repo.GetEnrollment(ctx, examID, studentID)
	if err != nil {
		return nil, err
	}

	exam, err := s.repo.GetExamByID(ctx, enrollment.CollegeID, examID)
	if err != nil {
		return nil, err
	}

	student, err := s.studentRepo.GetStudentByID(ctx, enrollment.CollegeID, studentID)
	if err != nil {
		return nil, err
	}

	hallTicket := &models.HallTicketResponse{
		ExamID:      examID,
		StudentID:   studentID,
		StudentName: student.Name,
		ExamTitle:   exam.Title,
		ExamDate:    exam.StartTime,
		StartTime:   exam.StartTime,
		EndTime:     exam.EndTime,
		Duration:    exam.Duration,
		Instructions: exam.Instructions,
	}

	if enrollment.SeatNumber != nil {
		hallTicket.SeatNumber = *enrollment.SeatNumber
	}
	if enrollment.RoomNumber != nil {
		hallTicket.RoomNumber = *enrollment.RoomNumber
	}
	if enrollment.QuestionPaperSet != nil {
		hallTicket.QuestionPaperSet = *enrollment.QuestionPaperSet
	}

	// Mark hall ticket as generated
	enrollment.HallTicketGenerated = true
	if err := s.repo.UpdateEnrollment(ctx, enrollment); err != nil {
		return nil, err
	}

	return hallTicket, nil
}

func (s *examService) GenerateAllHallTickets(ctx context.Context, examID int) error {
	enrollments, err := s.repo.ListEnrollments(ctx, examID)
	if err != nil {
		return err
	}

	for _, enrollment := range enrollments {
		_, err := s.GenerateHallTicket(ctx, examID, enrollment.StudentID)
		if err != nil {
			// Log error but continue with others
			continue
		}
	}

	return nil
}

// ===========================
// Result Management
// ===========================

func (s *examService) CreateResult(ctx context.Context, result *models.ExamResult) error {
	if result.ExamID == 0 || result.StudentID == 0 {
		return errors.New("exam ID and student ID are required")
	}
	if result.CollegeID == 0 {
		return errors.New("college ID is required")
	}

	// Get exam to validate marks
	exam, err := s.repo.GetExamByID(ctx, result.CollegeID, result.ExamID)
	if err != nil {
		return err
	}

	if result.MarksObtained != nil {
		if *result.MarksObtained < 0 || *result.MarksObtained > exam.TotalMarks {
			return errors.New("marks obtained must be between 0 and total marks")
		}

		// Calculate percentage and grade
		percentage := (*result.MarksObtained / exam.TotalMarks) * 100
		result.Percentage = &percentage

		grade := s.CalculateGrade(*result.MarksObtained, exam.TotalMarks)
		result.Grade = &grade

		// Determine pass/fail
		if *result.MarksObtained >= exam.PassingMarks {
			result.Result = "pass"
		} else {
			result.Result = "fail"
		}
	} else {
		result.Result = "pending"
	}

	// Set evaluation time
	now := time.Now()
	result.EvaluatedAt = &now

	return s.repo.CreateResult(ctx, result)
}

func (s *examService) GetResult(ctx context.Context, examID, studentID int) (*models.ExamResult, error) {
	if examID == 0 || studentID == 0 {
		return nil, errors.New("exam ID and student ID are required")
	}
	return s.repo.GetResult(ctx, examID, studentID)
}

func (s *examService) ListResults(ctx context.Context, examID int) ([]*models.ExamResult, error) {
	if examID == 0 {
		return nil, errors.New("exam ID is required")
	}
	return s.repo.ListResults(ctx, examID)
}

func (s *examService) UpdateResult(ctx context.Context, result *models.ExamResult) error {
	if result.ID == 0 {
		return errors.New("result ID is required")
	}
	return s.repo.UpdateResult(ctx, result)
}

func (s *examService) GetStudentResults(ctx context.Context, studentID, collegeID int) ([]*models.ExamResult, error) {
	if studentID == 0 || collegeID == 0 {
		return nil, errors.New("student ID and college ID are required")
	}
	return s.repo.GetStudentResults(ctx, studentID, collegeID)
}

func (s *examService) BulkGradeResults(ctx context.Context, examID int, results map[int]*ResultInput) error {
	for studentID, resultInput := range results {
		result, err := s.repo.GetResult(ctx, examID, studentID)
		if err != nil {
			// If result doesn't exist, create it
			result = &models.ExamResult{
				ExamID:    examID,
				StudentID: studentID,
			}
		}

		result.MarksObtained = &resultInput.MarksObtained
		result.Remarks = resultInput.Remarks

		if result.ID == 0 {
			if err := s.CreateResult(ctx, result); err != nil {
				continue
			}
		} else {
			if err := s.repo.UpdateResult(ctx, result); err != nil {
				continue
			}
		}
	}

	return nil
}

func (s *examService) CalculateGrade(marks, totalMarks float64) string {
	percentage := (marks / totalMarks) * 100

	switch {
	case percentage >= 90:
		return "A+"
	case percentage >= 80:
		return "A"
	case percentage >= 70:
		return "B+"
	case percentage >= 60:
		return "B"
	case percentage >= 50:
		return "C+"
	case percentage >= 40:
		return "C"
	default:
		return "F"
	}
}

func (s *examService) GetResultStats(ctx context.Context, examID int) (*ResultStats, error) {
	results, err := s.repo.ListResults(ctx, examID)
	if err != nil {
		return nil, err
	}

	stats := &ResultStats{
		TotalStudents: len(results),
		LowestMarks:   999999, // Initialize with high value
	}

	var totalMarks float64
	for _, result := range results {
		if result.MarksObtained == nil {
			continue
		}

		marks := *result.MarksObtained
		totalMarks += marks

		if result.Result == "pass" {
			stats.Passed++
		} else if result.Result == "fail" {
			stats.Failed++
		} else if result.Result == "absent" {
			stats.Absent++
		}

		if marks > stats.HighestMarks {
			stats.HighestMarks = marks
		}
		if marks < stats.LowestMarks {
			stats.LowestMarks = marks
		}
	}

	graded := stats.Passed + stats.Failed
	if graded > 0 {
		stats.AverageMarks = totalMarks / float64(graded)
		stats.PassPercentage = float64(stats.Passed) / float64(graded) * 100
	}

	if stats.LowestMarks == 999999 {
		stats.LowestMarks = 0
	}

	return stats, nil
}

// ===========================
// Revaluation Management
// ===========================

func (s *examService) CreateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error {
	if request.ExamResultID == 0 || request.StudentID == 0 {
		return errors.New("exam result ID and student ID are required")
	}
	if request.Reason == "" {
		return errors.New("reason is required")
	}

	// Set default status
	if request.Status == "" {
		request.Status = "pending"
	}

	return s.repo.CreateRevaluationRequest(ctx, request)
}

func (s *examService) GetRevaluationRequest(ctx context.Context, requestID int) (*models.RevaluationRequest, error) {
	if requestID == 0 {
		return nil, errors.New("request ID is required")
	}
	return s.repo.GetRevaluationRequest(ctx, requestID)
}

func (s *examService) ListRevaluationRequests(ctx context.Context, collegeID int, filters map[string]interface{}) ([]*models.RevaluationRequest, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}
	return s.repo.ListRevaluationRequests(ctx, collegeID, filters)
}

func (s *examService) UpdateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error {
	if request.ID == 0 {
		return errors.New("request ID is required")
	}
	return s.repo.UpdateRevaluationRequest(ctx, request)
}

func (s *examService) ApproveRevaluationRequest(ctx context.Context, requestID int, reviewedBy int, revisedMarks float64, comments string) error {
	request, err := s.repo.GetRevaluationRequest(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = "approved"
	request.RevisedMarks = &revisedMarks
	request.ReviewedBy = &reviewedBy
	request.ReviewComments = comments
	now := time.Now()
	request.ReviewedAt = &now

	// Update the original result
	// Note: This would need the exam_result_id to be accessible
	// Implementation depends on your requirements

	return s.repo.UpdateRevaluationRequest(ctx, request)
}

func (s *examService) RejectRevaluationRequest(ctx context.Context, requestID int, reviewedBy int, comments string) error {
	request, err := s.repo.GetRevaluationRequest(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = "rejected"
	request.ReviewedBy = &reviewedBy
	request.ReviewComments = comments
	now := time.Now()
	request.ReviewedAt = &now

	return s.repo.UpdateRevaluationRequest(ctx, request)
}

// ===========================
// Room Management
// ===========================

func (s *examService) CreateRoom(ctx context.Context, room *models.ExamRoom) error {
	if room.RoomNumber == "" || room.RoomName == "" {
		return errors.New("room number and name are required")
	}
	if room.CollegeID == 0 {
		return errors.New("college ID is required")
	}
	if room.Capacity <= 0 {
		return errors.New("capacity must be positive")
	}

	return s.repo.CreateRoom(ctx, room)
}

func (s *examService) GetRoom(ctx context.Context, collegeID, roomID int) (*models.ExamRoom, error) {
	if collegeID == 0 || roomID == 0 {
		return nil, errors.New("invalid college ID or room ID")
	}
	return s.repo.GetRoomByID(ctx, collegeID, roomID)
}

func (s *examService) ListRooms(ctx context.Context, collegeID int, activeOnly bool) ([]*models.ExamRoom, error) {
	if collegeID == 0 {
		return nil, errors.New("college ID is required")
	}
	return s.repo.ListRooms(ctx, collegeID, activeOnly)
}

func (s *examService) UpdateRoom(ctx context.Context, room *models.ExamRoom) error {
	if room.ID == 0 || room.CollegeID == 0 {
		return errors.New("invalid room ID or college ID")
	}
	return s.repo.UpdateRoom(ctx, room)
}

func (s *examService) DeleteRoom(ctx context.Context, collegeID, roomID int) error {
	if collegeID == 0 || roomID == 0 {
		return errors.New("invalid college ID or room ID")
	}
	return s.repo.DeleteRoom(ctx, collegeID, roomID)
}

func (s *examService) CheckRoomAvailability(ctx context.Context, roomID int, startTime, endTime string) (bool, error) {
	if roomID == 0 {
		return false, errors.New("room ID is required")
	}
	return s.repo.CheckRoomAvailability(ctx, roomID, startTime, endTime)
}
