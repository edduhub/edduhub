package handler

import (
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/analytics"
	"eduhub/server/internal/services/announcement"
	"eduhub/server/internal/services/assignment"
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
	assignmentService   assignment.AssignmentService
}

func NewDashboardHandler(
	studentService student.StudentService,
	courseService course.CourseService,
	attendanceService attendance.AttendanceService,
	announcementService announcement.AnnouncementService,
	calendarService calendar.CalendarService,
	analyticsService analytics.AnalyticsService,
	auditService audit.AuditService,
	assignmentService assignment.AssignmentService,
) *DashboardHandler {
	return &DashboardHandler{
		studentService:      studentService,
		courseService:       courseService,
		attendanceService:   attendanceService,
		announcementService: announcementService,
		calendarService:     calendarService,
		analyticsService:    analyticsService,
		auditService:        auditService,
		assignmentService:   assignmentService,
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

	// Get actual attendance rate from analytics service
	attendanceRate := 0.0
	totalFaculty := 0
	if dashboard, err := h.analyticsService.GetCollegeDashboard(ctx, collegeID); err == nil {
		attendanceRate = dashboard.AverageAttendance
		totalFaculty = dashboard.TotalFaculty
	}

	// Get recent announcements
	announcements := []map[string]interface{}{}
	isPublished := true
	announcementFilter := models.AnnouncementFilter{
		CollegeID:   &collegeID,
		IsPublished: &isPublished,
		Limit:       5,
		Offset:      0,
	}
	if announcementList, err := h.announcementService.GetAnnouncements(ctx, announcementFilter); err == nil {
		for _, a := range announcementList {
			announcements = append(announcements, map[string]interface{}{
				"id":       a.ID,
				"title":    a.Title,
				"content":  a.Content,
				"priority": a.Priority,
			})
		}
	}

	// Get upcoming calendar events
	upcomingEvents := []map[string]interface{}{}
	now := time.Now()
	calendarFilter := models.CalendarBlockFilter{
		CollegeID: &collegeID,
		StartDate: &now,
		Limit:     10,
		Offset:    0,
	}
	if events, err := h.calendarService.GetEvents(ctx, calendarFilter); err == nil {
		for _, event := range events {
			upcomingEvents = append(upcomingEvents, map[string]interface{}{
				"id":     event.ID,
				"title":  event.Title,
				"start":  event.Date,
				"course": event.Description,
			})
		}
	}

	// Get recent audit activity
	recentActivity := []map[string]interface{}{}
	if auditLogs, err := h.auditService.GetAuditLogs(ctx, collegeID, nil, "", "", 10, 0); err == nil {
		for _, log := range auditLogs {
			recentActivity = append(recentActivity, map[string]interface{}{
				"id":        log.ID,
				"entity":    log.EntityType,
				"message":   log.Action,
				"timestamp": log.Timestamp,
			})
		}
	}

	// Get pending submissions count
	pendingSubmissions := 0
	if count, err := h.assignmentService.CountPendingSubmissionsByCollege(ctx, collegeID); err == nil {
		pendingSubmissions = count
	}

	// Build response with real data
	response := map[string]interface{}{
		"metrics": map[string]interface{}{
			"totalStudents":      totalStudents,
			"totalCourses":       totalCourses,
			"totalFaculty":       totalFaculty,
			"attendanceRate":     attendanceRate,
			"pendingSubmissions": pendingSubmissions,
		},
		"announcements":  announcements,
		"upcomingEvents": upcomingEvents,
		"recentActivity": recentActivity,
	}

	return helpers.Success(c, response, 200)
}
