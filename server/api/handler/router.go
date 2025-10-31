package handler

import (
	"eduhub/server/internal/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, a *Handlers, m *middleware.AuthMiddleware) {
	// Public routes
	e.GET("/health", a.System.HealthCheck)
	e.GET("/ready", a.System.ReadinessCheck)
	e.GET("/alive", a.System.LivenessCheck)

	// Register Swagger routes - make sure these are registered correctly
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(302, "/docs/index.html")
	})
	e.GET("/docs/*", echoSwagger.WrapHandler)
	// Auth routes (public)
	auth := e.Group("/auth")
	auth.GET("/register", a.Auth.InitiateRegistration)
	auth.POST("/register/complete", a.Auth.HandleRegistration)
	auth.POST("/login", a.Auth.HandleLogin)
	auth.GET("/callback", a.Auth.HandleCallback, m.ValidateSession)

	// Auth routes (require authentication)
	auth.POST("/logout", a.Auth.HandleLogout, m.ValidateSession)
	auth.POST("/refresh", a.Auth.RefreshToken)

	// Password management (public)
	auth.POST("/password-reset", a.Auth.RequestPasswordReset)
	auth.POST("/password-reset/complete", a.Auth.CompletePasswordReset)

	// Email verification (public)
	auth.GET("/verify-email", a.Auth.VerifyEmail)
	auth.POST("/verify-email/initiate", a.Auth.InitiateEmailVerification, m.ValidateJWT)

	// Password change (authenticated)
	auth.POST("/change-password", a.Auth.ChangePassword, m.ValidateJWT)

	// Protected API routes with audit logging
	apiGroup := e.Group("/api", m.ValidateSession, m.RequireCollege)

	// Dashboard
	apiGroup.GET("/dashboard", a.Dashboard.GetDashboard)

	// Student Dashboard (student-specific comprehensive view)
	student := apiGroup.Group("/student", m.RequireRole(middleware.RoleStudent))
	student.GET("/dashboard", a.Dashboard.GetStudentDashboard)

	// User profile management
	profile := apiGroup.Group("/profile")
	profile.GET("", a.Profile.GetUserProfile)
	profile.PATCH("", a.Profile.UpdateUserProfile)
	profile.POST("/upload-image", a.Profile.UploadProfileImage)
	profile.GET("/history", a.Profile.GetProfileHistory)
	profile.GET("/:profileID", a.Profile.GetProfile, m.RequireRole(middleware.RoleAdmin))

	// College management
	college := apiGroup.Group("/college", m.RequireRole(middleware.RoleAdmin))
	college.GET("", a.College.GetCollegeDetails)
	college.PATCH("", a.College.UpdateCollegeDetails) // PATCH: Allows partial updates to college details
	college.GET("/stats", a.College.GetCollegeStats)

	// User management
	users := apiGroup.Group("/users", m.RequireRole(middleware.RoleAdmin))
	users.GET("", a.User.ListUsers)
	users.POST("", a.User.CreateUser)
	users.GET("/:userID", a.User.GetUser)
	users.PATCH("/:userID", a.User.UpdateUser)
	users.DELETE("/:userID", a.User.DeleteUser)
	users.PATCH("/:userID/role", a.User.UpdateUserRole)
	users.PATCH("/:userID/status", a.User.UpdateUserStatus)

	// Student management
	students := apiGroup.Group("/students", m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	students.GET("", a.Student.ListStudents)
	students.POST("", a.Student.CreateStudent, m.RequireRole(middleware.RoleAdmin))
	students.GET("/:studentID", a.Student.GetStudent)
	students.PATCH("/:studentID", a.Student.UpdateStudent, m.RequireRole(middleware.RoleAdmin)) // PATCH: Allows partial updates to student details
	students.DELETE("/:studentID", a.Student.DeleteStudent, m.RequireRole(middleware.RoleAdmin))
	students.PUT("/:studentID/freeze", a.Student.FreezeStudent, m.RequireRole(middleware.RoleAdmin))

	// Course management
	courses := apiGroup.Group("/courses")
	courses.GET("", a.Course.ListCourses)
	courses.POST("", a.Course.CreateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	courses.GET("/:courseID", a.Course.GetCourse)
	courses.PATCH("/:courseID", a.Course.UpdateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	courses.DELETE("/:courseID", a.Course.DeleteCourse, m.RequireRole(middleware.RoleAdmin))

	// Course enrollment
	courses.POST("/:courseID/enroll", a.Course.EnrollStudents, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	courses.DELETE("/:courseID/students/:studentID", a.Course.RemoveStudent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	courses.GET("/:courseID/students", a.Course.ListEnrolledStudents)

	// Lecture management
	lectures := apiGroup.Group("/courses/:courseID/lectures")
	lectures.GET("", a.Lecture.ListLectures)
	lectures.POST("", a.Lecture.CreateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	lectures.GET("/:lectureID", a.Lecture.GetLecture)
	lectures.PATCH("/:lectureID", a.Lecture.UpdateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty)) // PATCH: Allows partial updates to lecture details
	lectures.DELETE("/:lectureID", a.Lecture.DeleteLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Attendance management
	attendance := apiGroup.Group("/attendance")
	attendance.POST("/mark/course/:courseID/lecture/:lectureID", a.Attendance.MarkAttendance,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())
	attendance.POST("/mark/bulk/course/:courseID/lecture/:lectureID", a.Attendance.MarkBulkAttendance,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	attendance.GET("/course/:courseID/lecture/:lectureID/qrcode", a.Attendance.GenerateQRCode,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	attendance.GET("/course/:courseID", a.Attendance.GetAttendanceByCourse,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// Add convenience endpoint for current user
	attendance.GET("/student/me", a.Attendance.GetMyAttendance,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)
	attendance.GET("/student/:studentID", a.Attendance.GetAttendanceForStudent,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())
	attendance.GET("/student/:studentID/course/:courseID", a.Attendance.GetAttendanceByStudentAndCourse,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())
	attendance.PUT("/course/:courseID/lecture/:lectureID/student/:studentID", a.Attendance.UpdateAttendance,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty)) // PUT retained: Updates attendance status (full update, not partial update pattern)
	attendance.GET("/report/:studentID", a.Attendance.GetAttendanceForStudent, m.RequireRole(middleware.RoleAdmin, middleware.RoleStudent), m.VerifyStudentOwnership())
	attendance.POST("/process-qr", a.Attendance.ProcessAttendance, m.RequireRole(middleware.RoleStudent), m.LoadStudentProfile)
	// Add stats endpoint
	attendance.GET("/stats/courses", a.Attendance.GetMyCourseStats,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)
	// Grades/Assessment management
	grades := apiGroup.Group("/grades")
	// Add convenience endpoints for current user
	grades.GET("", a.Grade.GetMyGrades,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)
	grades.GET("/courses", a.Grade.GetMyCourseGrades,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)
	grades.GET("/course/:courseID", a.Grade.GetGradesByCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.POST("/course/:courseID", a.Grade.CreateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.PATCH("/course/:courseID/assessment/:assessmentID", a.Grade.UpdateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty)) // PATCH: Allows partial updates to assessment
	grades.DELETE("/course/:courseID/assessment/:assessmentID", a.Grade.DeleteAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.POST("/course/:courseID/assessment/:assessmentID/scores", a.Grade.SubmitScores, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.GET("/student/:studentID", a.Grade.GetStudentGrades,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())

	// Calendar/Schedule management
	calendar := apiGroup.Group("/calendar")
	calendar.GET("", a.Calendar.GetEvents)
	calendar.POST("", a.Calendar.CreateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	calendar.PATCH("/:eventID", a.Calendar.UpdateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	calendar.DELETE("/:eventID", a.Calendar.DeleteEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Department management
	departments := apiGroup.Group("/departments", m.RequireRole(middleware.RoleAdmin))
	departments.GET("", a.Department.GetDepartments)
	departments.POST("", a.Department.CreateDepartment)
	departments.GET("/:departmentID", a.Department.GetDepartment)
	departments.PATCH("/:departmentID", a.Department.UpdateDepartment)
	departments.DELETE("/:departmentID", a.Department.DeleteDepartment)

	// Assignment management
	assignments := apiGroup.Group("/courses/:courseID/assignments")
	assignments.GET("", a.Assignment.ListAssignments)
	assignments.POST("", a.Assignment.CreateAssignment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	assignments.GET("/:assignmentID", a.Assignment.GetAssignment)
	assignments.PATCH("/:assignmentID", a.Assignment.UpdateAssignment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	assignments.DELETE("/:assignmentID", a.Assignment.DeleteAssignment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	assignments.POST("/:assignmentID/submit", a.Assignment.SubmitAssignment, m.RequireRole(middleware.RoleStudent), m.LoadStudentProfile)
	assignments.POST("/submissions/:submissionID/grade", a.Assignment.GradeSubmission, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// Assignment grading enhancements
	assignments.GET("/:assignmentID/submissions", a.Assignment.ListSubmissionsByAssignment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	assignments.POST("/:assignmentID/submissions/bulk-grade", a.Assignment.BulkGradeSubmissions, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	assignments.GET("/:assignmentID/stats", a.Assignment.GetAssignmentGradingStats, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Convenience endpoint for all assignments (current user)
	assignmentsAll := apiGroup.Group("/assignments")
	assignmentsAll.GET("", a.Assignment.GetMyAssignments,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)

	// Quiz management
	quizzes := apiGroup.Group("/courses/:courseID/quizzes")
	quizzes.GET("", a.Quiz.ListQuizzes)
	quizzes.POST("", a.Quiz.CreateQuiz, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	quizzes.GET("/:quizID", a.Quiz.GetQuiz)
	quizzes.PATCH("/:quizID", a.Quiz.UpdateQuiz, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	quizzes.DELETE("/:quizID", a.Quiz.DeleteQuiz, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Convenience endpoint for all quizzes (current user)
	quizzesAll := apiGroup.Group("/quizzes")
	quizzesAll.GET("", a.Quiz.GetMyQuizzes,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)

	// Announcement management
	announcements := apiGroup.Group("/announcements")
	announcements.GET("", a.Announcement.ListAnnouncements)
	announcements.POST("", a.Announcement.CreateAnnouncement, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	announcements.GET("/:announcementID", a.Announcement.GetAnnouncement)
	announcements.PATCH("/:announcementID", a.Announcement.UpdateAnnouncement, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	announcements.DELETE("/:announcementID", a.Announcement.DeleteAnnouncement, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Question Bank management
	questions := apiGroup.Group("/quizzes/:quizID/questions")
	questions.GET("", a.Question.ListQuestions)
	questions.POST("", a.Question.CreateQuestion, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	questions.GET("/:questionID", a.Question.GetQuestion)
	questions.PATCH("/:questionID", a.Question.UpdateQuestion, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	questions.DELETE("/:questionID", a.Question.DeleteQuestion, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Quiz Attempts management
	quizAttempts := apiGroup.Group("/quizzes/:quizID/attempts")
	quizAttempts.POST("/start", a.QuizAttempt.StartQuizAttempt, m.RequireRole(middleware.RoleStudent))
	quizAttempts.GET("", a.QuizAttempt.ListQuizAttempts, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	attemptRoutes := apiGroup.Group("/attempts")
	attemptRoutes.GET("/:attemptID", a.QuizAttempt.GetQuizAttempt)
	attemptRoutes.POST("/:attemptID/submit", a.QuizAttempt.SubmitQuizAttempt, m.RequireRole(middleware.RoleStudent))
	attemptRoutes.GET("/student/:studentID", a.QuizAttempt.ListStudentAttempts)

	// File Upload management (legacy)
	files := apiGroup.Group("/files")
	files.POST("/upload", a.FileUpload.UploadFile)
	files.DELETE("", a.FileUpload.DeleteFile)
	files.GET("/url", a.FileUpload.GetFileURL)

	// Advanced File Management with versioning
	fileGroup := apiGroup.Group("/file-management")
	fileGroup.POST("/upload", a.File.UploadFile)
	fileGroup.GET("", a.File.ListFiles)
	fileGroup.GET("/:fileID", a.File.GetFile)
	fileGroup.PATCH("/:fileID", a.File.UpdateFile)
	fileGroup.DELETE("/:fileID", a.File.DeleteFile)
	fileGroup.GET("/:fileID/versions", a.File.GetFileVersions)
	fileGroup.POST("/:fileID/versions", a.File.UploadNewVersion)
	fileGroup.PATCH("/:fileID/versions/:versionID/current", a.File.SetCurrentVersion)
	fileGroup.GET("/:fileID/download", a.File.GetFileURL)

	// Folder management
	folders := apiGroup.Group("/folders")
	folders.POST("", a.File.CreateFolder)
	folders.GET("", a.File.ListFolders)
	folders.GET("/:folderID", a.File.GetFolder)
	folders.PATCH("/:folderID", a.File.UpdateFolder)
	folders.DELETE("/:folderID", a.File.DeleteFolder)

	// File search and tagging
	fileGroup.GET("/search", a.File.SearchFiles)
	fileGroup.GET("/tags", a.File.GetFilesByTags)

	// Notification management
	notifications := apiGroup.Group("/notifications")
	notifications.GET("", a.Notification.GetNotifications)
	notifications.POST("", a.Notification.SendNotification, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	notifications.GET("/unread/count", a.Notification.GetUnreadCount)
	notifications.PATCH("/:notificationID/read", a.Notification.MarkAsRead)
	notifications.POST("/mark-all-read", a.Notification.MarkAllAsRead)
	notifications.DELETE("/:notificationID", a.Notification.DeleteNotification)

	// WebSocket connection for real-time notifications
	notifications.GET("/ws", a.WebSocket.HandleWebSocket)

	// Analytics management
	analytics := apiGroup.Group("/analytics", m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	analytics.GET("/dashboard", a.Analytics.GetCollegeDashboard)
	analytics.GET("/students/:studentID/performance", a.Analytics.GetStudentPerformance)
	analytics.GET("/courses/:courseID/analytics", a.Analytics.GetCourseAnalytics)
	analytics.GET("/courses/:courseID/grades/distribution", a.Analytics.GetGradeDistribution)
	analytics.GET("/attendance/trends", a.Analytics.GetAttendanceTrends)

	advancedAnalytics := analytics.Group("/advanced")
	advancedAnalytics.GET("/students/:studentID/progression", a.AdvancedAnalytics.GetStudentProgression)
	advancedAnalytics.GET("/courses/:courseID/engagement", a.AdvancedAnalytics.GetCourseEngagement)
	advancedAnalytics.GET("/predictive-insights", a.AdvancedAnalytics.GetPredictiveInsights)
	advancedAnalytics.GET("/learning-analytics", a.AdvancedAnalytics.GetLearningAnalytics)
	advancedAnalytics.GET("/performance/:entityType/:entityID/trends", a.AdvancedAnalytics.GetPerformanceTrends)
	advancedAnalytics.GET("/courses/comparative", a.AdvancedAnalytics.GetComparativeAnalysis)

	// Batch Operations management
	batch := apiGroup.Group("/batch", m.RequireRole(middleware.RoleAdmin))
	batch.POST("/students/import", a.Batch.ImportStudents)
	batch.GET("/students/export", a.Batch.ExportStudents)
	batch.POST("/grades/import", a.Batch.ImportGrades)
	batch.GET("/grades/export", a.Batch.ExportGrades)
	batch.POST("/enroll", a.Batch.BulkEnroll)

	// Report Generation management
	reports := apiGroup.Group("/reports")
	// Student convenience endpoints (access own reports)
	reports.GET("/students/me/gradecard", a.Report.GenerateMyGradeCard, 
		m.RequireRole(middleware.RoleStudent), 
		m.LoadStudentProfile)
	reports.GET("/students/me/transcript", a.Report.GenerateMyTranscript,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile)
	// Admin/Faculty endpoints (access any student's reports)
	reports.GET("/students/:studentID/gradecard", a.Report.GenerateGradeCard,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	reports.GET("/students/:studentID/transcript", a.Report.GenerateTranscript,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	reports.GET("/courses/:courseID/attendance", a.Report.GenerateAttendanceReport,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	reports.GET("/courses/:courseID/report", a.Report.GenerateCourseReport,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Webhook management
	webhooks := apiGroup.Group("/webhooks", m.RequireRole(middleware.RoleAdmin))
	webhooks.GET("", a.Webhook.ListWebhooks)
	webhooks.POST("", a.Webhook.CreateWebhook)
	webhooks.GET("/:webhookID", a.Webhook.GetWebhook)
	webhooks.PATCH("/:webhookID", a.Webhook.UpdateWebhook)
	webhooks.DELETE("/:webhookID", a.Webhook.DeleteWebhook)
	webhooks.POST("/:webhookID/test", a.Webhook.TestWebhook)

	// Audit Logging management
	audit := apiGroup.Group("/audit", m.RequireRole(middleware.RoleAdmin))
	audit.GET("/logs", a.Audit.GetAuditLogs)
	audit.GET("/users/:userID/activity", a.Audit.GetUserActivity)
	audit.GET("/entities/:entityType/:entityID/history", a.Audit.GetEntityHistory)
	audit.GET("/stats", a.Audit.GetAuditStats)
}
