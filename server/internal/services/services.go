package services

import (
	"log"
	"time"

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
	"eduhub/server/internal/services/file"
	"eduhub/server/internal/services/grades"
	"eduhub/server/internal/services/lecture"
	"eduhub/server/internal/services/notification"
	"eduhub/server/internal/services/profile"
	"eduhub/server/internal/services/quiz"
	"eduhub/server/internal/services/report"
	storageservice "eduhub/server/internal/services/storage"
	"eduhub/server/internal/services/student"
	"eduhub/server/internal/services/user"
	"eduhub/server/internal/services/webhook"
	storageclient "eduhub/server/internal/storage"
	"eduhub/server/pkg/jwt"
	minio "github.com/minio/minio-go/v7"
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
	StorageService      storageservice.StorageService
	FileService         file.FileService
	NotificationService notification.NotificationService
	WebSocketService    notification.WebSocketService
	AnalyticsService    analytics.AnalyticsService
	AdvancedAnalyticsService analytics.AdvancedAnalyticsService
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

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(
		cfg.AuthConfig.JWTSecret,
		24*time.Hour, // Token valid for 24 hours
	)

	// Create auth service with JWT manager
	authService := auth.NewAuthService(kratosService, ketoService, jwtManager)

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
	// Create quizAttemptRepo early for quiz service
	quizAttemptRepo := repository.NewQuizAttemptRepository(cfg.DB)
	calendarRepo := repository.NewCalendarRepository(cfg.DB)
	departmentRepo := repository.NewDepartmentRepository(cfg.DB)

	var minioClient *storageclient.MinioClient
	storageBucket := "eduhub"
	storageEndpoint := "localhost:9000"
	storageUseSSL := false
	storageRegion := ""

	if cfg.StorageConfig != nil {
		if cfg.StorageConfig.Bucket != "" {
			storageBucket = cfg.StorageConfig.Bucket
		}
		if cfg.StorageConfig.Endpoint != "" {
			storageEndpoint = cfg.StorageConfig.Endpoint
		}
		storageUseSSL = cfg.StorageConfig.UseSSL
		storageRegion = cfg.StorageConfig.Region

		if cfg.StorageConfig.AccessKey != "" && cfg.StorageConfig.SecretKey != "" {
			client, err := storageclient.NewMinioClient(&storageclient.MinioConfig{
				Endpoint:  storageEndpoint,
				AccessKey: cfg.StorageConfig.AccessKey,
				SecretKey: cfg.StorageConfig.SecretKey,
				UseSSL:    storageUseSSL,
				Bucket:    storageBucket,
				Region:    storageRegion,
			})
			if err != nil {
				log.Printf("failed to initialize minio client: %v", err)
			} else {
				minioClient = client
			}
		}
	}

	assignmentRepo := repository.NewAssignmentRepository(cfg.DB, minioClient)
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
	quizService := quiz.NewQuizService(quizRepo, quizAttemptRepo, courseRepo, collegeRepo, enrollmentRepo)
	calendarService := calendar.NewCalendarService(calendarRepo)
	departmentService := department.NewDepartmentService(departmentRepo)
	assignmentService := assignment.NewAssignmentService(assignmentRepo, minioClient)
	userService := user.NewUserService(userRepo)
	announcementService := announcement.NewAnnouncementService(announcementRepo)
	profileService := profile.NewProfileService(profileRepo)

	// New services
	questionRepo := repository.NewQuestionRepository(cfg.DB)
	// quizAttemptRepo already created earlier
	studentAnswerRepo := repository.NewStudentAnswerRepository(cfg.DB)
	notificationRepo := repository.NewNotificationRepository(cfg.DB)
	webhookRepo := repository.NewWebhookRepository(cfg.DB)
	auditRepo := repository.NewAuditLogRepository(cfg.DB)

	answerOptionRepo := repository.NewAnswerOptionRepository(cfg.DB)
    questionService := quiz.NewSimpleQuestionService(questionRepo)
    // Auto-grading service for quiz attempts
    autoGradingService := quiz.NewAutoGradingService(
        questionRepo,
        studentAnswerRepo,
        quizAttemptRepo,
        answerOptionRepo,
    )
    quizAttemptService := quiz.NewSimpleQuizAttemptService(
        quizAttemptRepo,
        studentAnswerRepo,
        quizRepo,
        questionRepo,
        answerOptionRepo,
        autoGradingService,
    )
	fileRepo := repository.NewFileRepository(cfg.DB)
	var minioNative *minio.Client
	if minioClient != nil {
		minioNative = minioClient.Client()
	}
	storageService := storageservice.NewStorageService(
		minioNative,
		storageBucket,
		storageEndpoint,
		storageUseSSL,
	)
	fileService := file.NewFileService(fileRepo, storageService)
	websocketService := notification.NewWebSocketService(notificationRepo)
	notificationService := notification.NewNotificationService(notificationRepo, websocketService)
	analyticsService := analytics.NewAnalyticsService(studentRepo, attendanceRepo, gradeRepo, courseRepo, assignmentRepo, cfg.DB)
	advancedAnalyticsService := analytics.NewAdvancedAnalyticsService(cfg.DB, analyticsService)
	batchService := batch.NewBatchService(studentRepo, enrollmentRepo, gradeRepo)
	reportService := report.NewReportService(studentRepo, gradeRepo, attendanceRepo, enrollmentRepo, courseRepo)
	webhookService := webhook.NewWebhookService(webhookRepo)
	auditService := audit.NewAuditService(auditRepo)
	emailService := email.NewEmailService(
		cfg.EmailConfig.Host,
		cfg.EmailConfig.Port,
		cfg.EmailConfig.Username,
		cfg.EmailConfig.Password,
		cfg.EmailConfig.FromAddress,
	)

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
		FileService:         fileService,
		NotificationService: notificationService,
		WebSocketService:    websocketService,
		AnalyticsService:    analyticsService,
		AdvancedAnalyticsService: advancedAnalyticsService,
		BatchService:        batchService,
		ReportService:       reportService,
		WebhookService:      webhookService,
		AuditService:        auditService,
		EmailService:        emailService,
		DB:                  cfg.DB,
	}
}
