package handler

import (
	"eduhub/server/internal/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func SetupRoutes(e *echo.Echo, a *Handlers, m *middleware.AuthMiddleware) {
	// Public routes
	// e.GET("/health", a.System.HealthCheck)
	// e.GET("/docs/*", a.System.SwaggerDocs)

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
	auth.GET("/callback", a.Auth.HandleCallback, m.ValidateJWT)
	// auth.POST("/logout", a.Auth.HandleLogout)
	// auth.POST("/refresh", a.Auth.RefreshToken)
	// auth.POST("/password/reset/request", a.Auth.RequestPasswordReset)
	// auth.POST("/password/reset/complete", a.Auth.CompletePasswordReset)

	// Protected API routes
	apiGroup := e.Group("/api", m.ValidateJWT, m.RequireCollege)

	// User profile management
	profile := apiGroup.Group("/profile")
	profile.GET("", a.User.GetProfile)
	profile.PATCH("", a.User.UpdateProfile)    // PATCH: Allows partial updates to user profile
	profile.PATCH("/password", a.User.ChangePassword)    // PATCH: Allows partial updates to password

	// College management
	college := apiGroup.Group("/college", m.RequireRole(middleware.RoleAdmin))
	college.GET("", a.College.GetCollegeDetails)
	college.PATCH("", a.College.UpdateCollegeDetails)    // PATCH: Allows partial updates to college details
	college.GET("/stats", a.College.GetCollegeStats)

	// User management
	users := apiGroup.Group("/users", m.RequireRole(middleware.RoleAdmin))
	users.GET("", a.User.ListUsers)
	users.POST("", a.User.CreateUser)
	users.GET("/:userID", a.User.GetUser)
	users.PATCH("/:userID", a.User.UpdateUser)    // PATCH: Allows partial updates to user details
	users.DELETE("/:userID", a.User.DeleteUser)
	users.PATCH("/:userID/role", a.User.UpdateUserRole)    // PATCH: Allows partial updates to user role
	users.PATCH("/:userID/status", a.User.UpdateUserStatus)    // PATCH: Allows partial updates to user status

	// Student management
	students := apiGroup.Group("/students", m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	students.GET("", a.Student.ListStudents)
	students.POST("", a.Student.CreateStudent, m.RequireRole(middleware.RoleAdmin))
	students.GET("/:studentID", a.Student.GetStudent)
	students.PATCH("/:studentID", a.Student.UpdateStudent, m.RequireRole(middleware.RoleAdmin))    // PATCH: Allows partial updates to student details
	students.DELETE("/:studentID", a.Student.DeleteStudent, m.RequireRole(middleware.RoleAdmin))
	students.PUT("/:studentID/freeze", a.Student.FreezeStudent, m.RequireRole(middleware.RoleAdmin))

	// Course management
	courses := apiGroup.Group("/courses")
	// courses.GET("", a.Course.ListCourses)
	// courses.POST("", a.Course.CreateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.GET("/:courseID", a.Course.GetCourse)
	courses.PATCH("/:courseID", a.Course.UpdateCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))    // PATCH: Allows partial updates to course details
	// courses.DELETE("/:courseID", a.Course.DeleteCourse, m.RequireRole(middleware.RoleAdmin))

	// // Course enrollment
	// courses.POST("/:courseID/enroll", a.Course.EnrollStudents, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.DELETE("/:courseID/students/:studentID", a.Course.RemoveStudent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	// courses.GET("/:courseID/students", a.Course.ListEnrolledStudents)

	// Lecture management
	lectures := apiGroup.Group("/courses/:courseID/lectures")
	lectures.GET("", a.Lecture.ListLectures)
	lectures.POST("", a.Lecture.CreateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	lectures.GET("/:lectureID", a.Lecture.GetLecture)
	lectures.PATCH("/:lectureID", a.Lecture.UpdateLecture, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))    // PATCH: Allows partial updates to lecture details
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
	attendance.GET("/student/:studentID", a.Attendance.GetAttendanceForStudent,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())
	attendance.GET("/student/:studentID/course/:courseID", a.Attendance.GetAttendanceByStudentAndCourse,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership())
	attendance.PUT("/course/:courseID/lecture/:lectureID/student/:studentID", a.Attendance.UpdateAttendance,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))    // PUT retained: Updates attendance status (full update, not partial update pattern)
	attendance.GET("/report/:studentID", a.Attendance.GetAttendanceForStudent, m.RequireRole(middleware.RoleAdmin, middleware.RoleStudent), m.VerifyStudentOwnership())
	attendance.POST("/process-qr", a.Attendance.ProcessAttendance, m.RequireRole(middleware.RoleStudent), m.LoadStudentProfile)
	// Grades/Assessment management
	grades := apiGroup.Group("/grades")
	grades.GET("/course/:courseID", a.Grade.GetGradesByCourse, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.POST("/course/:courseID", a.Grade.CreateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.PATCH("/course/:courseID/assessment/:assessmentID", a.Grade.UpdateAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))    // PATCH: Allows partial updates to assessment
	grades.DELETE("/course/:courseID/assessment/:assessmentID", a.Grade.DeleteAssessment, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.POST("/course/:courseID/assessment/:assessmentID/scores", a.Grade.SubmitScores, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	grades.GET("/student/:studentID", a.Grade.GetStudentGrades,
		m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty, middleware.RoleStudent),
		m.LoadStudentProfile,
		m.VerifyStudentOwnership)

	// Calendar/Schedule management
	calendar := apiGroup.Group("/calendar")
	calendar.GET("", a.Calendar.GetEvents)
	calendar.POST("", a.Calendar.CreateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
	calendar.PATCH("/:eventID", a.Calendar.UpdateEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))    // PATCH: Allows partial updates to calendar event
	calendar.DELETE("/:eventID", a.Calendar.DeleteEvent, m.RequireRole(middleware.RoleAdmin, middleware.RoleFaculty))
}
