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

	// Create individual repository instances using modular approach
	studentRepo := repository.NewStudentRepository(cfg.DB)
	attendanceRepo := repository.NewAttendanceRepository(cfg.DB.Pool)
	enrollmentRepo := repository.NewEnrollmentRepository(cfg.DB)
	profileRepo := repository.NewProfileRepository(cfg.DB)
	gradeRepo := repository.NewGradeRepository(cfg.DB)
	collegeRepo := repository.NewCollegeRepository(cfg.DB)
	courseRepo := repository.NewCourseRepository(cfg.DB)
	userRepo := repository.NewUserRepository(cfg.DB)
	lectureRepo := repository.NewLectureRepository(cfg.DB)
	quizRepo := repository.NewQuizRepository(cfg.DB)

	studentService := student.NewstudentService(
		studentRepo,
		attendanceRepo,
		enrollmentRepo,
		profileRepo,
		gradeRepo,
	)
	// systemService := system.NewSystemService(cfg.DB)
	attendanceService := attendance.NewAttendanceService(attendanceRepo, studentRepo, enrollmentRepo)
	collegeService := college.NewCollegeService(collegeRepo)
	courseService := course.NewCourseService(courseRepo, collegeRepo, userRepo)
	gradeService := grades.NewGradeServices(gradeRepo, studentRepo, enrollmentRepo, courseRepo)
	lectureService := lecture.NewLectureService(lectureRepo)
	quizService := quiz.NewQuizService(quizRepo, courseRepo, collegeRepo, enrollmentRepo)

	return &Services{
		Auth:           authService,
		Attendance:     attendanceService,
		StudentService: studentService,
		CollegeService: collegeService,
		CourseService:  courseService,
		GradeService:   gradeService,
		LectureService: lectureService,
		QuizService:    quizService,
	}
}
