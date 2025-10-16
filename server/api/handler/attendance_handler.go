package handler

import (
	"fmt"
	"net/http"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models" // Import models package

	// "eduhub/server/internal/middleware" // Assuming validator is setup via middleware or directly
	"eduhub/server/internal/services/attendance"

	"github.com/labstack/echo/v4"
)

type AttendanceHandler struct {
	attendanceService attendance.AttendanceService
}

// BulkAttendanceRequest defines the structure for the bulk attendance marking endpoint.
type BulkAttendanceRequest struct {
	Attendances []models.StudentAttendanceStatus `json:"attendances" validate:"required,dive"` // Use dive for validating nested structs
}

type QRCodeRequest struct {
	QRCodeData string `json:"qrcode_data"`
}

func NewAttendanceHandler(attendance attendance.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendance,
	}
}

func (a *AttendanceHandler) GenerateQRCode(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return err
	}
	lectureID, err := helpers.GetIDFromParam(c, "lectureID")
	if err != nil {
		return err
	}
	qrCode, err := a.attendanceService.GenerateQRCode(ctx, collegeID, courseID, lectureID)
	if err != nil {
		return helpers.Error(c, err, 400)
	}
	return helpers.Success(c, qrCode, 200)
}

func (a *AttendanceHandler) ProcessAttendance(c echo.Context) error {
	ctx := c.Request().Context()
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	studentId, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}
	var qrcodeData QRCodeRequest
	if err := c.Bind(&qrcodeData); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}
	err = a.attendanceService.ProcessQRCode(ctx, collegeID, studentId, qrcodeData.QRCodeData)
	return err
}

func (a *AttendanceHandler) MarkAttendance(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}

	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return err
	}
	lectureID, err := helpers.GetIDFromParam(c, "lectureID")
	if err != nil {
		return err
	}

	ok, err := a.attendanceService.MarkAttendance(ctx, collegeID, studentID, courseID, lectureID)
	if err != nil {
		return helpers.Error(c, err.Error(), http.StatusInternalServerError)
	}

	if !ok {
		return helpers.Error(c, "Failed to mark attendance", http.StatusInternalServerError)
	}

	return helpers.Success(c, "attendance marked", http.StatusOK)
}

func (a *AttendanceHandler) GetAttendanceByCourse(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return helpers.Error(c, "Invalid course ID", http.StatusBadRequest)
	}

	attendance, err := a.attendanceService.GetAttendanceByCourse(ctx, collegeID, courseID, 100, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return helpers.Success(c, attendance, http.StatusOK)
}

// use structs instead of maps while returning c.JSON
func (a *AttendanceHandler) GetAttendanceForStudent(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "invalid collegeID", 400)
	}
	studentID, err := helpers.GetIDFromParam(c, "studentID")
	if err != nil {
		return helpers.Error(c, "invalid studentID", 400)
	}

	attendance, err := a.attendanceService.GetAttendanceByStudent(ctx, collegeID, studentID, 100, 0)
	if err != nil {
		return helpers.Error(c, "unable to get attendance by student", http.StatusInternalServerError)
	}
	return helpers.Success(c, attendance, http.StatusOK)
}

func (a *AttendanceHandler) GetAttendanceByStudentAndCourse(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	studentID, err := helpers.GetIDFromParam(c, "studentID")
	if err != nil {
		return err
	}

	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return err
	}

	attendance, err := a.attendanceService.GetAttendanceByStudentAndCourse(ctx, collegeID, studentID, courseID, 100, 0)
	if err != nil {
		return helpers.Error(c, "unable to get attendance", http.StatusInternalServerError)
	}
	return helpers.Success(c, attendance, 200)
}

func (a *AttendanceHandler) UpdateAttendance(c echo.Context) error {
	ctx := c.Request().Context()
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "Invalid collegeID", 400)
	}
	lectureID, err := helpers.GetIDFromParam(c, "lectureID")
	if err != nil {
		return helpers.Error(c, "invalid lectureID", 400)
	}
	studentID, err := helpers.GetIDFromParam(c, "studentID")
	if err != nil {
		return helpers.Error(c, "invalid studentID", 400)
	}
	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return helpers.Error(c, "invalid courseID ", 400)
	}

	ok, err := a.attendanceService.UpdateAttendanceStatus(ctx, collegeID, studentID, courseID, lectureID, "Present")
	if !ok {
		return helpers.Error(c, "Unable update attendance", 500)
	}
	if err != nil {
		return helpers.Error(c, "error in update attendance", 502)
	}

	return helpers.Success(c, "Success", 200)
}

func (a *AttendanceHandler) FreezeAttendance(c echo.Context) error {
	ctx := c.Request().Context()
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "Invalid collegeID", 400)
	}
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "Invalid student ID", 400)
	}
	ok, _ := a.attendanceService.FreezeAttendance(ctx, collegeID, studentID)
	if !ok {
		return helpers.Error(c, "unable to freeze attendance", 500)
	}
	return helpers.Success(c, "Freezed Attendance", 200)
}

// GetMyAttendance gets attendance for the currently authenticated student
func (a *AttendanceHandler) GetMyAttendance(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "invalid collegeID", 400)
	}
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "invalid studentID", 400)
	}

	attendance, err := a.attendanceService.GetAttendanceByStudent(ctx, collegeID, studentID, 100, 0)
	if err != nil {
		return helpers.Error(c, "unable to get attendance by student", http.StatusInternalServerError)
	}
	return helpers.Success(c, attendance, http.StatusOK)
}

// GetMyCourseStats gets course-wise attendance stats for current student
func (a *AttendanceHandler) GetMyCourseStats(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "invalid collegeID", 400)
	}
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return helpers.Error(c, "invalid studentID", 400)
	}

	// Get all attendance records for the student
	attendance, err := a.attendanceService.GetAttendanceByStudent(ctx, collegeID, studentID, 1000, 0)
	if err != nil {
		return helpers.Error(c, "unable to get attendance", http.StatusInternalServerError)
	}

	// Group by course and calculate stats
	courseStats := make(map[int]struct {
		CourseName string
		CourseCode string
		Present    int
		Total      int
		Percentage float64
	})

	for _, record := range attendance {
		stats := courseStats[record.CourseID]
		stats.Total++
		if record.Status == "Present" {
			stats.Present++
		}
		// Note: CourseCode and CourseName would need to be fetched from course service
		// For now, using courseID as placeholder
		stats.CourseCode = fmt.Sprintf("COURSE-%d", record.CourseID)
		stats.CourseName = fmt.Sprintf("Course %d", record.CourseID)
		if stats.Total > 0 {
			stats.Percentage = float64(stats.Present) / float64(stats.Total) * 100
		}
		courseStats[record.CourseID] = stats
	}

	// Convert map to slice
	result := []map[string]interface{}{}
	for _, stats := range courseStats {
		result = append(result, map[string]interface{}{
			"courseName": stats.CourseName,
			"courseCode": stats.CourseCode,
			"present":    stats.Present,
			"total":      stats.Total,
			"percentage": stats.Percentage,
		})
	}

	return helpers.Success(c, result, http.StatusOK)
}

func (a *AttendanceHandler) MarkBulkAttendance(c echo.Context) error {
	ctx := c.Request().Context()

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return helpers.Error(c, "invalid collegeID", http.StatusBadRequest)
	}

	courseID, err := helpers.GetIDFromParam(c, "courseID")
	if err != nil {
		return err // GetIDFromParam already returns formatted error
	}
	lectureID, err := helpers.GetIDFromParam(c, "lectureID")
	if err != nil {
		return err // GetIDFromParam already returns formatted error
	}

	var req BulkAttendanceRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "Invalid request body: "+err.Error(), http.StatusBadRequest)
	}

	// Optional: Add validation using a validator library if you have one integrated
	// if err := c.Validate(&req); err != nil {
	// 	 return helpers.Error(c, "Validation failed: "+err.Error(), http.StatusBadRequest)
	// }

	err = a.attendanceService.MarkBulkAttendance(ctx, collegeID, courseID, lectureID, req.Attendances)
	if err != nil {
		// The service layer now returns a more descriptive error if needed
		return helpers.Error(c, "Failed to mark bulk attendance: "+err.Error(), http.StatusInternalServerError)
	}

	return helpers.Success(c, "Bulk attendance marked successfully", http.StatusOK)
}
