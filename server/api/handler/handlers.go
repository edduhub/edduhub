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
	Notification *NotificationHandler
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
		Profile:      NewProfileHandler(services.ProfileService),
		System:       NewSystemHandler(services.DB),
		Question:     NewQuestionHandler(services.QuestionService),
		QuizAttempt:  NewQuizAttemptHandler(services.QuizAttemptService),
		FileUpload:   NewFileUploadHandler(services.StorageService),
		Notification: NewNotificationHandler(services.NotificationService),
		Analytics:    NewAnalyticsHandler(services.AnalyticsService),
		Batch:        NewBatchHandler(services.BatchService),
		Report:       NewReportHandler(services.ReportService),
		Webhook:      NewWebhookHandler(services.WebhookService),
		Audit:        NewAuditHandler(services.AuditService),
	}
}