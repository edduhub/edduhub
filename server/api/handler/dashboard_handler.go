package handler

import (
	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/analytics"
	"eduhub/server/internal/services/announcement"
	"eduhub/server/internal/services/attendance"
	"eduhub/server/internal/services/audit"
	"eduhub/server/internal/services/calendar"
	"eduhub/server/internal/services/course"
	"eduhub/server/internal/services/student"

	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	studentService      student.StudentService
	courseService       course.CourseService
	attendanceService   attendance.AttendanceService
	announcementService announcement.AnnouncementService
	calendarService     calendar.CalendarService
	analyticsService    analytics.AnalyticsService
	auditService        audit.AuditService
}

func NewDashboardHandler(
	studentService student.StudentService,
	courseService course.CourseService,
	attendanceService attendance.AttendanceService,
	announcementService announcement.AnnouncementService,
	calendarService calendar.CalendarService,
	analyticsService analytics.AnalyticsService,
	auditService audit.AuditService,
) *DashboardHandler {
	return &DashboardHandler{
		studentService:      studentService,
		courseService:       courseService,
		attendanceService:   attendanceService,
		announcementService: announcementService,
		calendarService:     calendarService,
		analyticsService:    analyticsService,
		auditService:        auditService,
	}
}

// GetDashboard returns dashboard data based on user role
func (h *DashboardHandler) GetDashboard(c echo.Context) error {
	ctx := c.Request().Context()
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Get total students
	totalStudents := 0
	if students, err := h.studentService.ListStudents(ctx, collegeID, 1000, 0); err == nil {
		totalStudents = len(students)
	}

	// Get total courses
	totalCourses := 0
	if courses, err := h.courseService.FindAllCourses(ctx, collegeID, 1000, 0); err == nil {
		totalCourses = len(courses)
	}

	// Placeholder for attendance rate - proper implementation requires service review
	attendanceRate := 0

	// Build response with basic metrics
	response := map[string]interface{}{
		"metrics": map[string]interface{}{
			"totalStudents":  totalStudents,
			"totalCourses":   totalCourses,
			"attendanceRate": attendanceRate,
		},
		"announcements":  []map[string]interface{}{},
		"upcomingEvents": []map[string]interface{}{},
		"recentActivity": []map[string]interface{}{},
	}

	return helpers.Success(c, response, 200)
}
