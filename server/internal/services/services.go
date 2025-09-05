package services

import (
	"eduhub/server/internal/config"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/college"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/lecture"
	"eduhub/server/internal/services/quiz" // Added Quiz service import
	"eduhub/server/internal/services/student"
)

type Services struct {
	Auth           auth.AuthService
	Attendance     attendance.AttendanceService
	StudentService student.StudentService
	CollegeService college.CollegeService
	CourseService  course.CourseService
	GradeService   grades.GradeServices
	LectureService lecture.LectureService
	QuizService    quiz.QuizService // Added QuizService field

	// Fee *Fee.FeeService
}

func NewServices(cfg *config.Config) *Services {
	kratosService := auth.NewKratosService()
	ketoService := auth.NewKetoService()
	authService := auth.NewAuthService(kratosService, ketoService)
	repo := repository.NewRepository(cfg.DB)

	studentService := student.NewstudentService(
		repo.StudentRepository,
		repo.AttendanceRepository,
		repo.EnrollmentRepository,
		repo.ProfileRepository, // Added ProfileRepository
		repo.GradeRepository,   // Added GradeRepository
	)
	// systemService := system.NewSystemService(cfg.DB)
	attendanceService := attendance.NewAttendanceService(repo.AttendanceRepository, repo.StudentRepository, repo.EnrollmentRepository)
	collegeService := college.NewCollegeService(repo.CollegeRepository)
	courseService := course.NewCourseService(repo.CourseRepository)
	gradeService := grades.NewGradeServices(repo.GradeRepository, repo.StudentRepository, repo.EnrollmentRepository, repo.CourseRepository)
	lectureService := lecture.NewLectureService(repo.LectureRepository)
	quizService := quiz.NewQuizService(repo.QuizRepository, repo.CourseRepository, repo.CollegeRepository, repo.EnrollmentRepository) // Initialize QuizService

	return &Services{
		Auth:           authService,
		Attendance:     attendanceService,
		StudentService: studentService,
		CollegeService: collegeService,
		CourseService:  courseService,
		GradeService:   gradeService,
		LectureService: lectureService,
		QuizService:    quizService, // Add QuizService to the struct
	}
}
