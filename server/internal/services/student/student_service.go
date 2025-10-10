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
	UpdateStudentPartial(ctx context.Context, collegeID int, studentID int, req *models.UpdateStudentRequest) error
	ListStudents(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error)
	CreateStudent(ctx context.Context, student *models.Student) error
	DeleteStudent(ctx context.Context, collegeID int, studentID int) error
	FreezeStudent(ctx context.Context, collegeID int, studentID int) error
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
		profile, err := s.profileRepo.GetProfileByUserID(gCtx, student.KratosIdentityID)
		if err != nil && err.Error() != fmt.Sprintf("GetProfileByUserID: profile for user ID %s not found", student.KratosIdentityID) {
			return fmt.Errorf("failed to get profile: %w", err)
		}
		detailedProfile.Profile = profile
		return nil
	})

	// Fetch enrollments
	g.Go(func() error {
		enrollments, err := s.enrollmentRepo.FindEnrollmentsByStudent(gCtx, collegeID, studentID, 0, 0)
		if err != nil {
			return fmt.Errorf("failed to get enrollments: %w", err)
		}
		detailedProfile.Enrollments = enrollments
		return nil
	})

	return detailedProfile, g.Wait()
}

func (s *studentService) UpdateStudentPartial(ctx context.Context, collegeID int, studentID int, req *models.UpdateStudentRequest) error {
	return s.studentRepo.UpdateStudentPartial(ctx, collegeID, studentID, req)
}

func (s *studentService) ListStudents(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error) {
	if limit > 100 {
		limit = 100
	}
	return s.studentRepo.FindAllStudentsByCollege(ctx, collegeID, limit, offset)
}

func (s *studentService) CreateStudent(ctx context.Context, student *models.Student) error {
	return s.studentRepo.CreateStudent(ctx, student)
}

func (s *studentService) DeleteStudent(ctx context.Context, collegeID int, studentID int) error {
	return s.studentRepo.DeleteStudent(ctx, collegeID, studentID)
}

func (s *studentService) FreezeStudent(ctx context.Context, collegeID int, studentID int) error {
	return s.attendanceRepo.FreezeAttendance(ctx, collegeID, studentID)
}