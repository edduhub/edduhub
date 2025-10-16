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
	students, err := h.studentService.ListStudents(ctx, collegeID, 1000, 0)
	if err != nil {
		return helpers.Error(c, "failed to fetch students", 500)
	}
	totalStudents := len(students)

	// Get total courses
	courses, err := h.courseService.FindAllCourses(ctx, collegeID, 1000, 0)
	if err != nil {
		return helpers.Error(c, "failed to fetch courses", 500)
	}
	totalCourses := len(courses)

	// Get attendance metrics - calculate average across all students
	var totalAttendanceRate float64
	if len(students) > 0 {
		for _, student := range students {
			records, err := h.attendanceService.GetAttendanceByStudent(ctx, collegeID, student.StudentID, 1000, 0)
			if err == nil && len(records) > 0 {
				present := 0
				for _, record := range records {
					if record.Status == "Present" {
						present++
					}
				}
				if len(records) > 0 {
					totalAttendanceRate += float64(present) / float64(len(records)) * 100
				}
			}
		}
		if len(students) > 0 {
			totalAttendanceRate = totalAttendanceRate / float64(len(students))
		}
	}

	// Get announcements count
	announcements, err := h.announcementService.GetAnnouncements(ctx, struct {
		CollegeID int
		Limit     uint64
		Offset    uint64
	}{
		CollegeID: collegeID,
		Limit:     100,
		Offset:    0,
	})
	if err != nil {
		announcements = []*struct{}{}
	}

	// Get upcoming events
	events, err := h.calendarService.GetEvents(ctx, collegeID, 10, 0)
	if err != nil {
		events = []struct {
			ID       int
			Title    string
			Start    string
			End      string
			Course   *string
			Location *string
		}{}
	}

	// Convert events to response format
	upcomingEvents := []map[string]interface{}{}
	for _, event := range events {
		eventMap := map[string]interface{}{
			"id":    event.ID,
			"title": event.Title,
			"start": event.Start,
		}
		if event.Course != nil {
			eventMap["course"] = *event.Course
		}
		upcomingEvents = append(upcomingEvents, eventMap)
	}

	// Get recent activity from audit logs
	recentActivity := []map[string]interface{}{}
	auditLogs, err := h.auditService.GetAuditLogs(ctx, collegeID, 10, 0, "", "")
	if err == nil {
		for _, log := range auditLogs {
			recentActivity = append(recentActivity, map[string]interface{}{
				"id":        log.ID,
				"entity":    log.EntityType,
				"message":   log.Action + " on " + log.EntityType,
				"timestamp": log.Timestamp,
			})
		}
	}

	// Build response
	response := map[string]interface{}{
		"metrics": map[string]interface{}{
			"totalStudents":  totalStudents,
			"totalCourses":   totalCourses,
			"attendanceRate": int(totalAttendanceRate),
			"announcements":  len(announcements),
		},
		"upcomingEvents":  upcomingEvents,
		"recentActivity":  recentActivity,
	}

	return helpers.Success(c, response, 200)
}
