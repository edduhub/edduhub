package services

import (
	"eduhub/server/internal/config"
	"eduhub/server/internal/repository"
	"eduhub/server/internal/services/analytics"
	"eduhub/server/internal/services/announcement"
	"eduhub/server/internal/services/assignment"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/audit"
	"eduhub/server/internal/services/auth"
	"eduhub/server/internal/services/batch"
	"eduhub/server/internal/services/calendar"
	"eduhub/server/internal/services/college"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/department"
	"eduhub/server/internal/services/email"
	"eduhub/server/internal/services/enrollment"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/lecture"
	"eduhub/server/internal/services/notification"
	"eduhub/server/internal/services/profile"
	"eduhub/server/internal/services/quiz"
	"eduhub/server/internal/services/report"
	"eduhub/server/internal/services/storage"
	"eduhub/server/internal/services/student"
	"eduhub/server/internal/services/user"
	"eduhub/server/internal/services/webhook"
)

type Services struct {
	Auth                auth.AuthService
	Attendance          attendance.AttendanceService
	StudentService      student.StudentService
	CollegeService      college.CollegeService
	CourseService       course.CourseService
	EnrollmentService   enrollment.EnrollmentService
	GradeService        grades.GradeServices
	LectureService      lecture.LectureService
	QuizService         quiz.QuizService
	CalendarService     calendar.CalendarService
	DepartmentService   department.DepartmentService
	AssignmentService   assignment.AssignmentService
	UserService         user.UserService
	AnnouncementService announcement.AnnouncementService
	ProfileService      profile.ProfileService
	QuestionService     quiz.QuestionServiceSimple
	QuizAttemptService  quiz.QuizAttemptServiceSimple
	StorageService      storage.StorageService
	NotificationService notification.NotificationService
	AnalyticsService    analytics.AnalyticsService
	BatchService        batch.BatchService
	ReportService       report.ReportService
	WebhookService      webhook.WebhookService
	AuditService        audit.AuditService
	EmailService        email.EmailService
	DB                  *repository.DB
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
	calendarRepo := repository.NewCalendarRepository(cfg.DB)
	departmentRepo := repository.NewDepartmentRepository(cfg.DB)
	assignmentRepo := repository.NewAssignmentRepository(cfg.DB, nil)
	announcementRepo := repository.NewAnnouncementRepository(cfg.DB)

	studentService := student.NewstudentService(
		studentRepo,
		attendanceRepo,
		enrollmentRepo,
		profileRepo,
		gradeRepo,
	)
	// systemService := system.NewSystemService(cfg.DB)
	attendanceService := attendance.NewAttendanceService(attendanceRepo, studentRepo, enrollmentRepo)
	enrollmentService := enrollment.NewEnrollmentService(enrollmentRepo)
	collegeService := college.NewCollegeService(collegeRepo)
	courseService := course.NewCourseService(courseRepo, collegeRepo, userRepo)
	gradeService := grades.NewGradeServices(gradeRepo, studentRepo, enrollmentRepo, courseRepo)
	lectureService := lecture.NewLectureService(lectureRepo)
	quizService := quiz.NewQuizService(quizRepo, courseRepo, collegeRepo, enrollmentRepo)
	calendarService := calendar.NewCalendarService(calendarRepo)
	departmentService := department.NewDepartmentService(departmentRepo)
	assignmentService := assignment.NewAssignmentService(assignmentRepo, nil)
	userService := user.NewUserService(userRepo)
	announcementService := announcement.NewAnnouncementService(announcementRepo)
	profileService := profile.NewProfileService(profileRepo)

	// New services
	questionRepo := repository.NewQuestionRepository(cfg.DB)
	quizAttemptRepo := repository.NewQuizAttemptRepository(cfg.DB)
	studentAnswerRepo := repository.NewStudentAnswerRepository(cfg.DB)
	notificationRepo := repository.NewNotificationRepository(cfg.DB)
	webhookRepo := repository.NewWebhookRepository(cfg.DB)
	auditRepo := repository.NewAuditLogRepository(cfg.DB)

	questionService := quiz.NewSimpleQuestionService(questionRepo)
	quizAttemptService := quiz.NewSimpleQuizAttemptService(quizAttemptRepo, studentAnswerRepo, quizRepo)
	storageService := storage.NewStorageService(nil, "eduhub", "localhost:9000", false) // MinioClient will be nil for now
	notificationService := notification.NewNotificationService(notificationRepo)
	analyticsService := analytics.NewAnalyticsService(studentRepo, attendanceRepo, gradeRepo, courseRepo, assignmentRepo, cfg.DB)
	batchService := batch.NewBatchService(studentRepo, enrollmentRepo, gradeRepo)
	reportService := report.NewReportService(studentRepo, gradeRepo, attendanceRepo, enrollmentRepo, courseRepo)
	webhookService := webhook.NewWebhookService(webhookRepo)
	auditService := audit.NewAuditService(auditRepo)
	emailService := email.NewEmailService("", "", "", "", "")

	return &Services{
		Auth:                authService,
		Attendance:          attendanceService,
		StudentService:      studentService,
		CollegeService:      collegeService,
		CourseService:       courseService,
		EnrollmentService:   enrollmentService,
		GradeService:        gradeService,
		LectureService:      lectureService,
		QuizService:         quizService,
		CalendarService:     calendarService,
		DepartmentService:   departmentService,
		AssignmentService:   assignmentService,
		UserService:         userService,
		AnnouncementService: announcementService,
		ProfileService:      profileService,
		QuestionService:     questionService,
		QuizAttemptService:  quizAttemptService,
		StorageService:      storageService,
		NotificationService: notificationService,
		AnalyticsService:    analyticsService,
		BatchService:        batchService,
		ReportService:       reportService,
		WebhookService:      webhookService,
		AuditService:        auditService,
		EmailService:        emailService,
		DB:                  cfg.DB,
	}
}
