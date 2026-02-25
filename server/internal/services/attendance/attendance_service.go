package attendance

import (
	"context"
	"fmt"

	"eduhub/server/internal/cache"
	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

const (
	Present = "Present"
	Absent  = "Absent"
	Freezed = "Freezed"
)

type AttendanceService interface {
	GenerateQRCode(ctx context.Context, collegeID, courseID int, lectureID int) (string, error)
	GetAttendanceByLecture(ctx context.Context, collegeID, courseID int, lectureID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceByCourse(ctx context.Context, collegeID, courseID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceByStudent(ctx context.Context, collegeID, studentID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceByStudentAndCourse(ctx context.Context, collegeID, studentID int, courseID int, limit, offset uint64) ([]*models.Attendance, error)
	MarkAttendance(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int) (bool, error)
	UpdateAttendanceStatus(ctx context.Context, collegeID, studentID int, courseID int, lectureID int, newStatus string) (bool, error)
	FreezeAttendance(ctx context.Context, collegeID, studentID int) (bool, error)
	VerifyStudentStateAndEnrollment(ctx context.Context, collegeID, studentID, courseID int) (bool, error)
	ProcessQRCode(ctx context.Context, collegeID int, studentID int, qrCodeContent string) error
	MarkBulkAttendance(ctx context.Context, collegeID, courseID, lectureID int, studentStatuses []models.StudentAttendanceStatus) error
}
type attendanceService struct {
	repo           repository.AttendanceRepository
	studentRepo    repository.StudentRepository
	enrollmentRepo repository.EnrollmentRepository
	cache          cache.Cache // optional, nil when Redis disabled
}

func NewAttendanceService(repo repository.AttendanceRepository, studentRepo repository.StudentRepository, enrollmentRepo repository.EnrollmentRepository) AttendanceService {
	return &attendanceService{
		repo:           repo,
		studentRepo:    studentRepo,
		enrollmentRepo: enrollmentRepo,
		cache:          nil,
	}
}

// NewAttendanceServiceWithCache creates an attendance service with Redis cache for QR token validation
func NewAttendanceServiceWithCache(repo repository.AttendanceRepository, studentRepo repository.StudentRepository, enrollmentRepo repository.EnrollmentRepository, c cache.Cache) AttendanceService {
	return &attendanceService{
		repo:           repo,
		studentRepo:    studentRepo,
		enrollmentRepo: enrollmentRepo,
		cache:          c,
	}
}

func (a *attendanceService) GetAttendanceByLecture(ctx context.Context, collegeID int, courseID int, lectureID int, limit, offset uint64) ([]*models.Attendance, error) {
	return a.repo.GetAttendanceByLecture(ctx, collegeID, courseID, lectureID, limit, offset)
}

// to get attendance of all students in a course
func (a *attendanceService) GetAttendanceByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Attendance, error) {
	return a.repo.GetAttendanceByCourse(ctx, collegeID, courseID, limit, offset)
}

func (a *attendanceService) GetAttendanceByStudent(ctx context.Context, collegeID int, studentID int, limit uint64, offset uint64) ([]*models.Attendance, error) {
	return a.repo.GetAttendanceStudent(ctx, collegeID, studentID, limit, offset)
}

func (a *attendanceService) GetAttendanceByStudentAndCourse(ctx context.Context, collegeID int, studentID int, courseID int, limit uint64, offset uint64) ([]*models.Attendance, error) {
	return a.repo.GetAttendanceStudentInCourse(ctx, collegeID, studentID, courseID, limit, offset)
}

// manually mark attendance for a student
func (a *attendanceService) MarkAttendance(ctx context.Context, collegeID int, studentID, courseID, lectureID int) (bool, error) {
	ok, err := a.VerifyStudentStateAndEnrollment(ctx, collegeID, studentID, courseID)

	if !ok {
		return false, err
	}
	if err != nil {
		return false, err
	}

	// Mark attendance only if all verifications pass
	ok, err = a.repo.MarkAttendance(ctx, collegeID, studentID, courseID, lectureID)
	if err != nil {
		return false, fmt.Errorf("failed to mark attendance: %w", err)
	}

	return ok, nil
}

func (a *attendanceService) VerifyStudentStateAndEnrollment(ctx context.Context, collegeID int, studentID int, courseID int) (bool, error) {
	student, err := a.studentRepo.GetStudentByID(ctx, collegeID, studentID)
	if err != nil {
		return false, err
	}
	if !student.IsActive {
		return false, nil
	}
	if student.CollegeID != collegeID {
		return false, nil
	}
	// enrolled, err := a.studentRepo.VerifyStudentEnrollment(ctx, collegeID, studentID, courseID)
	exists, err := a.enrollmentRepo.IsStudentEnrolled(ctx, collegeID, studentID, courseID)
	return exists, err
}

func (a *attendanceService) GenerateAndProcessQRCode(ctx context.Context, collegeID, studentID int, courseID int, lectureID int) error {
	qrCode, err := a.GenerateQRCode(ctx, collegeID, courseID, lectureID)
	if err != nil {
		return err
	}
	err = a.ProcessQRCode(ctx, collegeID, studentID, qrCode)
	if err != nil {
		return err
	}
	return nil
}

func (a *attendanceService) UpdateAttendanceStatus(ctx context.Context, collegeID, studentID int, courseID int, lectureID int, newStatus string) (bool, error) {
	// Validate newStatus if necessary (e.g., ensure it's one of "Present", "Absent", "Late", etc.)
	validStatuses := map[string]bool{Present: true, Absent: true, Freezed: true, "Late": true, "Excused": true}
	if !validStatuses[newStatus] {
		return false, fmt.Errorf("invalid attendance status: %s", newStatus)
	}

	// Directly update the specific attendance record
	err := a.repo.UpdateAttendance(ctx, collegeID, studentID, courseID, lectureID, newStatus)
	if err != nil {
		return false, fmt.Errorf("failed to update attendance status: %w", err)
	}
	return true, nil
}

// manually mark attendance of multiple students
func (a *attendanceService) MarkBulkAttendance(ctx context.Context, collegeID, courseID, lectureID int, studentStatuses []models.StudentAttendanceStatus) error {
	var errors []error

	for _, studentStatus := range studentStatuses {
		// 1. Verify student state and enrollment *before* attempting to mark
		ok, err := a.VerifyStudentStateAndEnrollment(ctx, collegeID, studentStatus.StudentID, courseID)
		if err != nil {
			errors = append(errors, fmt.Errorf("error verifying student %d: %w", studentStatus.StudentID, err))
			continue // Skip this student if verification fails
		}
		if !ok {
			errors = append(errors, fmt.Errorf("student %d not active or not enrolled in course %d", studentStatus.StudentID, courseID))
			continue // Skip this student
		}

		// 2. Call repository to set the specific status
		err = a.repo.SetAttendanceStatus(ctx, collegeID, studentStatus.StudentID, courseID, lectureID, studentStatus.Status)
		if err != nil {
			errors = append(errors, fmt.Errorf("error setting attendance for student %d: %w", studentStatus.StudentID, err))
		}
	}
	// Combine errors if any occurred, otherwise return nil
	if len(errors) > 0 {
		// You might want a more sophisticated error aggregation depending on needs
		return fmt.Errorf("encountered %d error(s) during bulk attendance marking: %v", len(errors), errors)
	}
	return nil
}

func (a *attendanceService) FreezeAttendance(ctx context.Context, collegeID, studentID int) (bool, error) {
	err := a.repo.FreezeAttendance(ctx, collegeID, studentID)
	if err != nil {
		return false, err
	}
	return true, nil
}

