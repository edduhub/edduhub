package handler

import (
	"eduhub/server/internal/services"
)

type Handlers struct {
	Auth              *AuthHandler
	Dashboard         *DashboardHandler
	Attendance        *AttendanceHandler
	Student           *StudentHandler
	College           *CollegeHandler
	Course            *CourseHandler
	CourseMaterial    *CourseMaterialHandler
	Lecture           *LectureHandler
	Quiz              *QuizHandler
	Grade             *GradeHandler
	Calendar          *CalendarHandler
	Department        *DepartmentHandler
	Assignment        *AssignmentHandler
	User              *UserHandler
	Announcement      *AnnouncementHandler
	Profile           *ProfileHandler
	System            *SystemHandler
	Question          *QuestionHandler
	QuizAttempt       *QuizAttemptHandler
	FileUpload        *FileUploadHandler
	File              *FileHandler
	Notification      *NotificationHandler
	WebSocket         *WebSocketHandler
	Analytics         *AnalyticsHandler
	AdvancedAnalytics *AdvancedAnalyticsHandler
	Batch             *BatchHandler
	Report            *ReportHandler
	Webhook           *WebhookHandler
	Audit             *AuditHandler
	Role              *RoleHandler
	Fee               *FeeHandler
	Timetable         *TimetableHandler
	Exam              *ExamHandler
	Placement         *PlacementHandler
	Forum             *ForumHandler
	Parent            *ParentHandler
	SelfService       *SelfServiceHandler
}

func NewHandlers(services *services.Services) *Handlers {
	return &Handlers{
		Auth: NewAuthHandler(services.Auth),
		Dashboard: NewDashboardHandler(
			services.StudentService,
			services.CourseService,
			services.Attendance,
			services.AnnouncementService,
			services.CalendarService,
			services.AnalyticsService,
			services.AuditService,
			services.AssignmentService,
			services.EnrollmentService,
			services.GradeService,
		),
		Attendance:        NewAttendanceHandler(services.Attendance, services.CourseService),
		Student:           NewStudentHandler(services.StudentService),
		College:           NewCollegeHandler(services.CollegeService),
		Course:            NewCourseHandler(services.CourseService, services.EnrollmentService, services.StudentService),
		CourseMaterial:    NewCourseMaterialHandler(services.CourseMaterialService),
		Lecture:           NewLectureHandler(services.LectureService),
		Quiz:              NewQuizHandler(services.QuizService, services.EnrollmentService, services.CourseService),
		Grade:             NewGradeHandler(services.GradeService, services.CourseService),
		Calendar:          NewCalendarHandler(services.CalendarService),
		Department:        NewDepartmentHandler(services.DepartmentService),
		Assignment:        NewAssignmentHandler(services.AssignmentService, services.EnrollmentService, services.CourseService),
		User:              NewUserHandler(services.UserService),
		Announcement:      NewAnnouncementHandler(services.AnnouncementService),
		Profile:           NewProfileHandler(services.ProfileService, services.AuditService, services.StorageService),
		System:            NewSystemHandler(services.DB),
		Question:          NewQuestionHandler(services.QuestionService),
		QuizAttempt:       NewQuizAttemptHandler(services.QuizAttemptService),
		FileUpload:        NewFileUploadHandler(services.StorageService),
		File:              NewFileHandler(services.FileService),
		Notification:      NewNotificationHandler(services.NotificationService),
		WebSocket:         NewWebSocketHandler(services.WebSocketService),
		Analytics:         NewAnalyticsHandler(services.AnalyticsService),
		AdvancedAnalytics: NewAdvancedAnalyticsHandler(services.AdvancedAnalyticsService),
		Batch:             NewBatchHandler(services.BatchService),
		Report:            NewReportHandler(services.ReportService),
		Webhook:           NewWebhookHandler(services.WebhookService),
		Audit:             NewAuditHandler(services.AuditService),
		Role:              NewRoleHandler(services.RoleService),
		Fee:               NewFeeHandler(services.FeeService),
		Timetable:         NewTimetableHandler(services.TimetableService),
		Exam:              NewExamHandler(services.ExamService),
		Placement:         NewPlacementHandler(services.PlacementService),
		Forum:             NewForumHandler(services.ForumService),
		Parent: NewParentHandler(
			services.StudentService,
			services.Attendance,
			services.GradeService,
			services.AssignmentService,
		),
		SelfService: NewSelfServiceHandler(),
	}
}
