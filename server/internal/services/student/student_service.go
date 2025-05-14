package student

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"golang.org/x/sync/errgroup"
)

// StudentDetailedProfile aggregates student information with their profile and enrollments.
type StudentDetailedProfile struct {
	models.Student
	Profile     *models.Profile      `json:"profile,omitempty"`
	Enrollments []*models.Enrollment `json:"enrollments,omitempty"`
	// We could add GradeSummary or AttendanceSummary here later
}

type StudentService interface {
	FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error)
	GetStudentDetailedProfile(ctx context.Context, collegeID int, studentID int) (*StudentDetailedProfile, error)
	// Add other student-specific business logic methods here
	// For example:
	// GetStudentDashboardData(ctx context.Context, collegeID int, studentID int) (*StudentDashboard, error)
	// UpdateStudentAcademicInfo(ctx context.Context, student *models.Student, academicDetails models.AcademicDetails) error
}

type studentService struct {
	studentRepo    repository.StudentRepository
	attendanceRepo repository.AttendanceRepository
	enrollmentRepo repository.EnrollmentRepository
	profileRepo    repository.ProfileRepository
	gradeRepo      repository.GradeRepository
}

func NewstudentService(
	studentRepo repository.StudentRepository,
	attendanceRepo repository.AttendanceRepository,
	enrollmentRepo repository.EnrollmentRepository,
	profileRepo repository.ProfileRepository,
	gradeRepo repository.GradeRepository,
) StudentService {
	return &studentService{
		studentRepo:    studentRepo,
		attendanceRepo: attendanceRepo,
		enrollmentRepo: enrollmentRepo,
		profileRepo:    profileRepo,
		gradeRepo:      gradeRepo,
	}
}

func (a *studentService) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	student, err := a.studentRepo.FindByKratosID(ctx, kratosID)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func (s *studentService) GetStudentDetailedProfile(ctx context.Context, collegeID int, studentID int) (*StudentDetailedProfile, error) {
	student, err := s.studentRepo.GetStudentByID(ctx, collegeID, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student by ID: %w", err)
	}
	if student == nil {
		return nil, fmt.Errorf("student with ID %d not found in college %d", studentID, collegeID)
	}

	detailedProfile := &StudentDetailedProfile{
		Student: *student,
	}

	// Use errgroup for concurrent fetching of related data
	g, gCtx := errgroup.WithContext(ctx)

	// Fetch profile
	g.Go(func() error {
		profile, err := s.profileRepo.GetProfileByUserID(gCtx, student.KratosIdentityID)                                                // Assuming KratosIdentityID is the UserID for profile
		if err != nil && err.Error() != fmt.Sprintf("GetProfileByUserID: profile for user ID %s not found", student.KratosIdentityID) { // Don't error if profile simply not found
			return fmt.Errorf("failed to get profile: %w", err)
		}
		detailedProfile.Profile = profile
		return nil
	})

	// Fetch enrollments
	g.Go(func() error {
		enrollments, err := s.enrollmentRepo.FindEnrollmentsByStudent(gCtx, collegeID, studentID, 0, 0) // 0,0 for no limit/offset for now
		if err != nil {
			return fmt.Errorf("failed to get enrollments: %w", err)
		}
		detailedProfile.Enrollments = enrollments
		return nil
	})

	return detailedProfile, g.Wait()
}
