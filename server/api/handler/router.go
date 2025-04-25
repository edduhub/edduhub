package handler

import (
	"eduhub/server/internal/middleware"

	
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, a *Handlers, m *middleware.AuthMiddleware) {
	// Public routes
	e.GET("/health", a.System.HealthCheck)
	e.GET("/docs/*", a.System.SwaggerDocs)
	

	// Register Swagger routes - make sure these are registered correctly
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/docs", func(c echo.Context) error {
		return c.Redirect(302, "/docs/index.html")
	})
	e.GET("/docs/*", echoSwagger.WrapHandler)
	// Auth routes
	auth := e.Group("/auth")
	auth.GET("/register", a.Auth.InitiateRegistration)
	auth.POST("/register/complete", a.Auth.HandleRegistration)
	auth.POST("/login", a.Auth.HandleLogin)
	auth.GET("/callback", a.Auth.HandleCallback)
	// auth.POST("/logout", a.Auth.HandleLogout)
	// auth.POST("/refresh", a.Auth.RefreshToken)
	// auth.POST("/password/reset/request", a.Auth.RequestPasswordReset)
	// auth.POST("/password/reset/complete", a.Auth.CompletePasswordReset)

	// Protected API routes
	apiGroup := e.Group("/api", m.ValidateSession, m.RequireCollege)

	// // User profile management
	// profile := apiGroup.Group("/profile")
	// profile.GET("", a.User.GetProfile)
	// profile.PUT("", a.User.UpdateProfile)
	// profile.PUT("/password", a.User.ChangePassword)

	// // College management
	// college := apiGroup.Group("/college", m.RequireRole(middleware.RoleAdmin))
	// college.GET("", a.College.GetCollegeDetails)
	// college.PUT("", a.College.UpdateCollegeDetails)
	// college.GET("/stats", a.College.GetCollegeStats)

	// // User management
	// users := apiGroup.Group("/users", m.RequireRole(middleware.RoleAdmin))
	// users.GET("", a.User.ListUsers)
	// users.POST("", a.User.CreateUser)
	// users.GET("/:userID", a.User.GetUser)
	// users.PUT("/:userID", a.User.UpdateUser)
	// users.DELETE("/:userID", a.User.DeleteUser)
	// users.PUT("/:userID/role", a.User.UpdateUserRole)
	// users.PUT("/:userID/status", a.User.UpdateUserStatus)

	// // Student management
	// students := apiGroup.Group("/students", m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// students.GET("", a.Student.ListStudents)
	// students.POST("", a.Student.CreateStudent, m.RequireRole(middleware.RoleAdmin))
	// students.GET("/:studentID", a.Student.GetStudent)
	// students.PUT("/:studentID", a.Student.UpdateStudent, m.RequireRole(middleware.RoleAdmin))
	// students.DELETE("/:studentID", a.Student.DeleteStudent, m.RequireRole(middleware.RoleAdmin))
	// students.PUT("/:studentID/freeze", a.Student.FreezeStudent, m.RequireRole(middleware.RoleAdmin))

	// // Course management
	// courses := apiGroup.Group("/courses")
	// courses.GET("", a.Course.ListCourses)
	// courses.POST("", a.Course.CreateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.GET("/:courseID", a.Course.GetCourse)
	// courses.PUT("/:courseID", a.Course.UpdateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.DELETE("/:courseID", a.Course.DeleteCourse, m.RequireRole(middleware.RoleAdmin))

	// // Course enrollment
	// courses.POST("/:courseID/enroll", a.Course.EnrollStudents, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.DELETE("/:courseID/students/:studentID", a.Course.RemoveStudent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.GET("/:courseID/students", a.Course.ListEnrolledStudents)

	// // Lecture management
	// lectures := apiGroup.Group("/courses/:courseID/lectures")
	// lectures.GET("", a.Lecture.ListLectures)
	// lectures.POST("", a.Lecture.CreateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// lectures.GET("/:lectureID", a.Lecture.GetLecture)
	// lectures.PUT("/:lectureID", a.Lecture.UpdateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// lectures.DELETE("/:lectureID", a.Lecture.DeleteLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))

	// Attendance management
	attendance := apiGroup.Group("/attendance")
	attendance.POST("/mark/course/:courseID/lecture/:lectureID", a.Attendance.MarkAttendance,
		m.RequireRole(middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership)
	// attendance.POST("/mark/bulk/course/:courseID/lecture/:lectureID", a.Attendance.MarkBulkAttendance,
	// 	m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	attendance.GET("/course/:courseID/lecture/:lectureID/qrcode", a.Attendance.GenerateQRCode,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	attendance.GET("/course/:courseID", a.Attendance.GetAttendanceByCourse,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	attendance.GET("/student/:studentID", a.Attendance.GetAttendanceForStudent,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership)
	attendance.GET("/student/:studentID/course/:courseID", a.Attendance.GetAttendanceByStudentAndCourse,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership)
	// attendance.PUT("/course/:courseID/lecture/:lectureID/student/:studentID", a.Attendance.UpdateAttendance,
	// 	m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// attendance.GET("/reports/course/:courseID", a.Attendance.GetCourseAttendanceReport,
	// 	m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// attendance.GET("/reports/student/:studentID", a.Attendance.GetStudentAttendanceReport,
	// 	m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
	// 	m.LoadStudentProfile,
	// 	m.VerifyStudentOwnership)

	// 	// Grades/Assessment management
	// 	grades := apiGroup.Group("/grades")
	// 	grades.GET("/course/:courseID", a.Grade.GetGradesByCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	grades.POST("/course/:courseID", a.Grade.CreateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	grades.PUT("/course/:courseID/assessment/:assessmentID", a.Grade.UpdateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	grades.DELETE("/course/:courseID/assessment/:assessmentID", a.Grade.DeleteAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	grades.POST("/course/:courseID/assessment/:assessmentID/scores", a.Grade.SubmitScores, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	grades.GET("/student/:studentID", a.Grade.GetStudentGrades,
	// 		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
	// 		m.LoadStudentProfile,
	// 		m.VerifyStudentOwnership)

	// 	// Calendar/Schedule management
	// 	calendar := apiGroup.Group("/calendar")
	// 	calendar.GET("", a.Calendar.GetEvents)
	// 	calendar.POST("", a.Calendar.CreateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	calendar.PUT("/:eventID", a.Calendar.UpdateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// 	calendar.DELETE("/:eventID", a.Calendar.DeleteEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
}
