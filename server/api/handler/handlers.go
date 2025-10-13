package handler

import (
	"eduhub/server/internal/services"
)

type Handlers struct {
	Auth         *AuthHandler
	Attendance   *AttendanceHandler
	Student      *StudentHandler
	College      *CollegeHandler
	Course       *CourseHandler
	Lecture      *LectureHandler
	Quiz         *QuizHandler
	Grade        *GradeHandler
	Calendar     *CalendarHandler
	Department   *DepartmentHandler
	Assignment   *AssignmentHandler
	User         *UserHandler
	Announcement *AnnouncementHandler
	Profile      *ProfileHandler
	System       *SystemHandler
	Question     *QuestionHandler
	QuizAttempt  *QuizAttemptHandler
	FileUpload   *FileUploadHandler
	File         *FileHandler
	Notification *NotificationHandler
	WebSocket    *WebSocketHandler
	Analytics    *AnalyticsHandler
	Batch        *BatchHandler
	Report       *ReportHandler
	Webhook      *WebhookHandler
	Audit        *AuditHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Auth:         NewAuthHandler(services.Auth),
		Attendance:   NewAttendanceHandler(services.Attendance),
		Student:      NewStudentHandler(services.StudentService),
		College:      NewCollegeHandler(services.CollegeService),
		Course:       NewCourseHandler(services.CourseService, services.EnrollmentService, services.StudentService),
		Lecture:      NewLectureHandler(services.LectureService),
		Quiz:         NewQuizHandler(services.QuizService),
		Grade:        NewGradeHandler(services.GradeService),
		Calendar:     NewCalendarHandler(services.CalendarService),
		Department:   NewDepartmentHandler(services.DepartmentService),
		Assignment:   NewAssignmentHandler(services.AssignmentService),
		User:         NewUserHandler(services.UserService),
		Announcement: NewAnnouncementHandler(services.AnnouncementService),
		Profile:      NewProfileHandler(services.ProfileService, services.AuditService, services.StorageService),
		System:       NewSystemHandler(services.DB),
		Question:     NewQuestionHandler(services.QuestionService),
		QuizAttempt:  NewQuizAttemptHandler(services.QuizAttemptService),
		FileUpload:   NewFileUploadHandler(services.StorageService),
		File:         NewFileHandler(services.FileService),
		Notification: NewNotificationHandler(services.NotificationService),
		WebSocket:    NewWebSocketHandler(services.WebSocketService),
		Analytics:    NewAnalyticsHandler(services.AnalyticsService),
		Batch:        NewBatchHandler(services.BatchService),
		Report:       NewReportHandler(services.ReportService),
		Webhook:      NewWebhookHandler(services.WebhookService),
		Audit:        NewAuditHandler(services.AuditService),
	}
}