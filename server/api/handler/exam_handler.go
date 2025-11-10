package handler

import (
	"strconv"
	"time"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/exam"

	"github.com/labstack/echo/v4"
)

type ExamHandler struct {
	examService exam.ExamService
}

func NewExamHandler(examService exam.ExamService) *ExamHandler {
	return &ExamHandler{
		examService: examService,
	}
}

// ===========================
// Exam Management Handlers
// ===========================

// CreateExam creates a new exam
// POST /api/v1/exams
func (h *ExamHandler) CreateExam(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateExamRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	exam := &models.Exam{
		CollegeID:          collegeID,
		CourseID:           req.CourseID,
		Title:              req.Title,
		Description:        req.Description,
		ExamType:           req.ExamType,
		StartTime:          req.StartTime,
		EndTime:            req.EndTime,
		Duration:           req.Duration,
		TotalMarks:         req.TotalMarks,
		PassingMarks:       req.PassingMarks,
		Instructions:       req.Instructions,
		AllowedMaterials:   req.AllowedMaterials,
		QuestionPaperSets:  req.QuestionPaperSets,
		Status:             "scheduled",
		CreatedBy:          userID,
	}

	if err := h.examService.CreateExam(c.Request().Context(), exam); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, exam, 201)
}

// GetExam retrieves an exam by ID
// GET /api/v1/exams/:examID
func (h *ExamHandler) GetExam(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	exam, err := h.examService.GetExam(c.Request().Context(), collegeID, examID)
	if err != nil {
		return helpers.Error(c, "exam not found", 404)
	}

	return helpers.Success(c, exam, 200)
}

// ListExams lists all exams with optional filters
// GET /api/v1/exams
func (h *ExamHandler) ListExams(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse query parameters
	filters := make(map[string]interface{})

	if courseID := c.QueryParam("course_id"); courseID != "" {
		if id, err := strconv.Atoi(courseID); err == nil {
			filters["course_id"] = id
		}
	}
	if status := c.QueryParam("status"); status != "" {
		filters["status"] = status
	}
	if examType := c.QueryParam("exam_type"); examType != "" {
		filters["exam_type"] = examType
	}

	limit := 50
	offset := 0
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	exams, err := h.examService.ListExams(c.Request().Context(), collegeID, filters, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, exams, 200)
}

// ListExamsByCourse lists exams for a specific course
// GET /api/v1/courses/:courseID/exams
func (h *ExamHandler) ListExamsByCourse(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	courseID, err := strconv.Atoi(c.Param("courseID"))
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	limit := 50
	offset := 0
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	exams, err := h.examService.ListExamsByCourse(c.Request().Context(), collegeID, courseID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, exams, 200)
}

// UpdateExam updates an exam
// PUT /api/v1/exams/:examID
func (h *ExamHandler) UpdateExam(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	var exam models.Exam
	if err := c.Bind(&exam); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	exam.ID = examID
	exam.CollegeID = collegeID

	if err := h.examService.UpdateExam(c.Request().Context(), &exam); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "exam updated successfully", 200)
}

// DeleteExam deletes an exam
// DELETE /api/v1/exams/:examID
func (h *ExamHandler) DeleteExam(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	if err := h.examService.DeleteExam(c.Request().Context(), collegeID, examID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "exam deleted successfully", 200)
}

// GetExamStats retrieves statistics for an exam
// GET /api/v1/exams/:examID/stats
func (h *ExamHandler) GetExamStats(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	stats, err := h.examService.GetExamStats(c.Request().Context(), collegeID, examID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}

// ===========================
// Enrollment Handlers
// ===========================

// EnrollStudent enrolls a student in an exam
// POST /api/v1/exams/:examID/enroll
func (h *ExamHandler) EnrollStudent(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	var req struct {
		StudentID int `json:"student_id"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	enrollment := &models.ExamEnrollment{
		ExamID:    examID,
		StudentID: req.StudentID,
		CollegeID: collegeID,
		Status:    "enrolled",
	}

	if err := h.examService.EnrollStudent(c.Request().Context(), enrollment); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, enrollment, 201)
}

// EnrollMultipleStudents enrolls multiple students in an exam
// POST /api/v1/exams/:examID/enroll-bulk
func (h *ExamHandler) EnrollMultipleStudents(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	var req struct {
		StudentIDs []int `json:"student_ids"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.examService.EnrollMultipleStudents(c.Request().Context(), examID, collegeID, req.StudentIDs); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "students enrolled successfully", 201)
}

// ListEnrollments lists all enrollments for an exam
// GET /api/v1/exams/:examID/enrollments
func (h *ExamHandler) ListEnrollments(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	enrollments, err := h.examService.ListEnrollments(c.Request().Context(), examID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, enrollments, 200)
}

// GetStudentEnrollments lists all exams a student is enrolled in
// GET /api/v1/students/:studentID/exam-enrollments
func (h *ExamHandler) GetStudentEnrollments(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	enrollments, err := h.examService.GetStudentEnrollments(c.Request().Context(), studentID, collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, enrollments, 200)
}

// UpdateEnrollment updates an enrollment
// PUT /api/v1/exams/:examID/enrollments/:studentID
func (h *ExamHandler) UpdateEnrollment(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	enrollment, err := h.examService.GetEnrollment(c.Request().Context(), examID, studentID)
	if err != nil {
		return helpers.Error(c, "enrollment not found", 404)
	}

	var update struct {
		SeatNumber       *string `json:"seat_number"`
		RoomNumber       *string `json:"room_number"`
		QuestionPaperSet *int    `json:"question_paper_set"`
		Status           *string `json:"status"`
	}
	if err := c.Bind(&update); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if update.SeatNumber != nil {
		enrollment.SeatNumber = update.SeatNumber
	}
	if update.RoomNumber != nil {
		enrollment.RoomNumber = update.RoomNumber
	}
	if update.QuestionPaperSet != nil {
		enrollment.QuestionPaperSet = update.QuestionPaperSet
	}
	if update.Status != nil {
		enrollment.Status = *update.Status
	}

	if err := h.examService.UpdateEnrollment(c.Request().Context(), enrollment); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "enrollment updated successfully", 200)
}

// DeleteEnrollment removes a student from an exam
// DELETE /api/v1/exams/:examID/enrollments/:studentID
func (h *ExamHandler) DeleteEnrollment(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	if err := h.examService.DeleteEnrollment(c.Request().Context(), examID, studentID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "enrollment deleted successfully", 200)
}

// ===========================
// Seat Allocation & Hall Tickets
// ===========================

// AllocateSeats allocates seats for all enrolled students
// POST /api/v1/exams/:examID/allocate-seats
func (h *ExamHandler) AllocateSeats(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	if err := h.examService.AllocateSeats(c.Request().Context(), examID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "seats allocated successfully", 200)
}

// GenerateHallTicket generates hall ticket for a student
// GET /api/v1/exams/:examID/hall-ticket/:studentID
func (h *ExamHandler) GenerateHallTicket(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	hallTicket, err := h.examService.GenerateHallTicket(c.Request().Context(), examID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, hallTicket, 200)
}

// GenerateAllHallTickets generates hall tickets for all enrolled students
// POST /api/v1/exams/:examID/hall-tickets
func (h *ExamHandler) GenerateAllHallTickets(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	if err := h.examService.GenerateAllHallTickets(c.Request().Context(), examID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "hall tickets generated successfully", 200)
}

// ===========================
// Result Management Handlers
// ===========================

// CreateResult creates an exam result
// POST /api/v1/exams/:examID/results
func (h *ExamHandler) CreateResult(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	var req models.ExamResultRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	result := &models.ExamResult{
		ExamID:        examID,
		StudentID:     req.StudentID,
		CollegeID:     collegeID,
		MarksObtained: &req.MarksObtained,
		Remarks:       req.Remarks,
		EvaluatedBy:   &userID,
	}

	if err := h.examService.CreateResult(c.Request().Context(), result); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, result, 201)
}

// GetResult retrieves a specific exam result
// GET /api/v1/exams/:examID/results/:studentID
func (h *ExamHandler) GetResult(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	result, err := h.examService.GetResult(c.Request().Context(), examID, studentID)
	if err != nil {
		return helpers.Error(c, "result not found", 404)
	}

	return helpers.Success(c, result, 200)
}

// ListResults lists all results for an exam
// GET /api/v1/exams/:examID/results
func (h *ExamHandler) ListResults(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	results, err := h.examService.ListResults(c.Request().Context(), examID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, results, 200)
}

// GetStudentResults retrieves all exam results for a student
// GET /api/v1/students/:studentID/exam-results
func (h *ExamHandler) GetStudentResults(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	results, err := h.examService.GetStudentResults(c.Request().Context(), studentID, collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, results, 200)
}

// BulkGradeResults grades multiple exam results at once
// POST /api/v1/exams/:examID/bulk-grade
func (h *ExamHandler) BulkGradeResults(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	var req map[int]*exam.ResultInput
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.examService.BulkGradeResults(c.Request().Context(), examID, req); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "results graded successfully", 200)
}

// GetResultStats retrieves statistics for exam results
// GET /api/v1/exams/:examID/result-stats
func (h *ExamHandler) GetResultStats(c echo.Context) error {
	examID, err := strconv.Atoi(c.Param("examID"))
	if err != nil {
		return helpers.Error(c, "invalid exam ID", 400)
	}

	stats, err := h.examService.GetResultStats(c.Request().Context(), examID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}

// ===========================
// Revaluation Handlers
// ===========================

// CreateRevaluationRequest creates a revaluation request
// POST /api/v1/revaluation-requests
func (h *ExamHandler) CreateRevaluationRequest(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req struct {
		ExamResultID  int     `json:"exam_result_id"`
		StudentID     int     `json:"student_id"`
		Reason        string  `json:"reason"`
		PreviousMarks float64 `json:"previous_marks"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	request := &models.RevaluationRequest{
		ExamResultID:  req.ExamResultID,
		StudentID:     req.StudentID,
		CollegeID:     collegeID,
		Reason:        req.Reason,
		PreviousMarks: req.PreviousMarks,
		Status:        "pending",
		RequestedAt:   time.Now(),
	}

	if err := h.examService.CreateRevaluationRequest(c.Request().Context(), request); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, request, 201)
}

// ListRevaluationRequests lists revaluation requests
// GET /api/v1/revaluation-requests
func (h *ExamHandler) ListRevaluationRequests(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	filters := make(map[string]interface{})
	if status := c.QueryParam("status"); status != "" {
		filters["status"] = status
	}
	if studentID := c.QueryParam("student_id"); studentID != "" {
		if id, err := strconv.Atoi(studentID); err == nil {
			filters["student_id"] = id
		}
	}

	requests, err := h.examService.ListRevaluationRequests(c.Request().Context(), collegeID, filters)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, requests, 200)
}

// ApproveRevaluationRequest approves a revaluation request
// PUT /api/v1/revaluation-requests/:requestID/approve
func (h *ExamHandler) ApproveRevaluationRequest(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "invalid request ID", 400)
	}

	var req struct {
		RevisedMarks float64 `json:"revised_marks"`
		Comments     string  `json:"comments"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.examService.ApproveRevaluationRequest(c.Request().Context(), requestID, userID, req.RevisedMarks, req.Comments); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "revaluation request approved", 200)
}

// RejectRevaluationRequest rejects a revaluation request
// PUT /api/v1/revaluation-requests/:requestID/reject
func (h *ExamHandler) RejectRevaluationRequest(c echo.Context) error {
	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "invalid request ID", 400)
	}

	var req struct {
		Comments string `json:"comments"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.examService.RejectRevaluationRequest(c.Request().Context(), requestID, userID, req.Comments); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "revaluation request rejected", 200)
}

// ===========================
// Room Management Handlers
// ===========================

// CreateRoom creates a new exam room
// POST /api/v1/exam-rooms
func (h *ExamHandler) CreateRoom(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var room models.ExamRoom
	if err := c.Bind(&room); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	room.CollegeID = collegeID

	if err := h.examService.CreateRoom(c.Request().Context(), &room); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, room, 201)
}

// GetRoom retrieves a room by ID
// GET /api/v1/exam-rooms/:roomID
func (h *ExamHandler) GetRoom(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		return helpers.Error(c, "invalid room ID", 400)
	}

	room, err := h.examService.GetRoom(c.Request().Context(), collegeID, roomID)
	if err != nil {
		return helpers.Error(c, "room not found", 404)
	}

	return helpers.Success(c, room, 200)
}

// ListRooms lists all exam rooms
// GET /api/v1/exam-rooms
func (h *ExamHandler) ListRooms(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	activeOnly := c.QueryParam("active_only") == "true"

	rooms, err := h.examService.ListRooms(c.Request().Context(), collegeID, activeOnly)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, rooms, 200)
}

// UpdateRoom updates a room
// PUT /api/v1/exam-rooms/:roomID
func (h *ExamHandler) UpdateRoom(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		return helpers.Error(c, "invalid room ID", 400)
	}

	var room models.ExamRoom
	if err := c.Bind(&room); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	room.ID = roomID
	room.CollegeID = collegeID

	if err := h.examService.UpdateRoom(c.Request().Context(), &room); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "room updated successfully", 200)
}

// DeleteRoom deletes a room
// DELETE /api/v1/exam-rooms/:roomID
func (h *ExamHandler) DeleteRoom(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		return helpers.Error(c, "invalid room ID", 400)
	}

	if err := h.examService.DeleteRoom(c.Request().Context(), collegeID, roomID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "room deleted successfully", 200)
}

// CheckRoomAvailability checks if a room is available
// GET /api/v1/exam-rooms/:roomID/availability
func (h *ExamHandler) CheckRoomAvailability(c echo.Context) error {
	roomID, err := strconv.Atoi(c.Param("roomID"))
	if err != nil {
		return helpers.Error(c, "invalid room ID", 400)
	}

	startTime := c.QueryParam("start_time")
	endTime := c.QueryParam("end_time")

	if startTime == "" || endTime == "" {
		return helpers.Error(c, "start_time and end_time are required", 400)
	}

	available, err := h.examService.CheckRoomAvailability(c.Request().Context(), roomID, startTime, endTime)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]bool{"available": available}, 200)
}
