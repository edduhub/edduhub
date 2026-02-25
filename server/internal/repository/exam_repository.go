package repository

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
)

type ExamRepository interface {
	// Exam CRUD
	CreateExam(ctx context.Context, exam *models.Exam) error
	GetExamByID(ctx context.Context, collegeID, examID int) (*models.Exam, error)
	ListExams(ctx context.Context, collegeID int, filters map[string]any, limit, offset int) ([]*models.Exam, error)
	UpdateExam(ctx context.Context, exam *models.Exam) error
	DeleteExam(ctx context.Context, collegeID, examID int) error
	ListExamsByCourse(ctx context.Context, collegeID, courseID int, limit, offset int) ([]*models.Exam, error)

	// Exam Enrollment
	EnrollStudent(ctx context.Context, enrollment *models.ExamEnrollment) error
	GetEnrollment(ctx context.Context, examID, studentID int) (*models.ExamEnrollment, error)
	ListEnrollments(ctx context.Context, examID int) ([]*models.ExamEnrollment, error)
	UpdateEnrollment(ctx context.Context, enrollment *models.ExamEnrollment) error
	DeleteEnrollment(ctx context.Context, examID, studentID int) error
	GetStudentEnrollments(ctx context.Context, studentID, collegeID int) ([]*models.ExamEnrollment, error)

	// Exam Results
	CreateResult(ctx context.Context, result *models.ExamResult) error
	GetResult(ctx context.Context, examID, studentID int) (*models.ExamResult, error)
	GetResultByID(ctx context.Context, resultID int) (*models.ExamResult, error)
	ListResults(ctx context.Context, examID int) ([]*models.ExamResult, error)
	UpdateResult(ctx context.Context, result *models.ExamResult) error
	GetStudentResults(ctx context.Context, studentID, collegeID int) ([]*models.ExamResult, error)

	// Revaluation Requests
	CreateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error
	GetRevaluationRequest(ctx context.Context, requestID int) (*models.RevaluationRequest, error)
	ListRevaluationRequests(ctx context.Context, collegeID int, filters map[string]any) ([]*models.RevaluationRequest, error)
	UpdateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error

	// Exam Rooms
	CreateRoom(ctx context.Context, room *models.ExamRoom) error
	GetRoomByID(ctx context.Context, collegeID, roomID int) (*models.ExamRoom, error)
	ListRooms(ctx context.Context, collegeID int, activeOnly bool) ([]*models.ExamRoom, error)
	UpdateRoom(ctx context.Context, room *models.ExamRoom) error
	DeleteRoom(ctx context.Context, collegeID, roomID int) error
	CheckRoomAvailability(ctx context.Context, roomID int, startTime, endTime string) (bool, error)
}

type examRepository struct {
	db *DB
}

func NewExamRepository(db *DB) ExamRepository {
	return &examRepository{db: db}
}

// CreateExam creates a new exam
func (r *examRepository) CreateExam(ctx context.Context, exam *models.Exam) error {
	sql := `
		INSERT INTO exams (college_id, course_id, title, description, exam_type, start_time,
			end_time, duration, total_marks, passing_marks, room_id, status, instructions,
			allowed_materials, question_paper_sets, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, sql,
		exam.CollegeID, exam.CourseID, exam.Title, exam.Description, exam.ExamType,
		exam.StartTime, exam.EndTime, exam.Duration, exam.TotalMarks, exam.PassingMarks,
		exam.RoomID, exam.Status, exam.Instructions, exam.AllowedMaterials,
		exam.QuestionPaperSets, exam.CreatedBy,
	).Scan(&exam.ID, &exam.CreatedAt, &exam.UpdatedAt)
}

// GetExamByID retrieves an exam by ID
func (r *examRepository) GetExamByID(ctx context.Context, collegeID, examID int) (*models.Exam, error) {
	sql := `SELECT id, college_id, course_id, title, description, exam_type, start_time,
			end_time, duration, total_marks, passing_marks, room_id, status, instructions,
			allowed_materials, question_paper_sets, created_by, created_at, updated_at
			FROM exams WHERE id = $1 AND college_id = $2`

	exam := &models.Exam{}
	err := r.db.Pool.QueryRow(ctx, sql, examID, collegeID).Scan(
		&exam.ID, &exam.CollegeID, &exam.CourseID, &exam.Title, &exam.Description,
		&exam.ExamType, &exam.StartTime, &exam.EndTime, &exam.Duration, &exam.TotalMarks,
		&exam.PassingMarks, &exam.RoomID, &exam.Status, &exam.Instructions,
		&exam.AllowedMaterials, &exam.QuestionPaperSets, &exam.CreatedBy,
		&exam.CreatedAt, &exam.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("exam not found: %w", err)
	}
	return exam, nil
}

// ListExams retrieves exams with optional filters
func (r *examRepository) ListExams(ctx context.Context, collegeID int, filters map[string]any, limit, offset int) ([]*models.Exam, error) {
	sql := `SELECT id, college_id, course_id, title, description, exam_type, start_time,
			end_time, duration, total_marks, passing_marks, room_id, status, instructions,
			allowed_materials, question_paper_sets, created_by, created_at, updated_at
			FROM exams WHERE college_id = $1`
	args := []any{collegeID}
	argCount := 1

	// Add optional filters
	if courseID, ok := filters["course_id"]; ok {
		argCount++
		sql += fmt.Sprintf(" AND course_id = $%d", argCount)
		args = append(args, courseID)
	}
	if status, ok := filters["status"]; ok {
		argCount++
		sql += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	if examType, ok := filters["exam_type"]; ok {
		argCount++
		sql += fmt.Sprintf(" AND exam_type = $%d", argCount)
		args = append(args, examType)
	}

	sql += " ORDER BY start_time DESC LIMIT $" + fmt.Sprintf("%d", argCount+1) + " OFFSET $" + fmt.Sprintf("%d", argCount+2)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exams []*models.Exam
	for rows.Next() {
		exam := &models.Exam{}
		err := rows.Scan(
			&exam.ID, &exam.CollegeID, &exam.CourseID, &exam.Title, &exam.Description,
			&exam.ExamType, &exam.StartTime, &exam.EndTime, &exam.Duration, &exam.TotalMarks,
			&exam.PassingMarks, &exam.RoomID, &exam.Status, &exam.Instructions,
			&exam.AllowedMaterials, &exam.QuestionPaperSets, &exam.CreatedBy,
			&exam.CreatedAt, &exam.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		exams = append(exams, exam)
	}
	return exams, nil
}

// UpdateExam updates an exam
func (r *examRepository) UpdateExam(ctx context.Context, exam *models.Exam) error {
	sql := `UPDATE exams SET title = $1, description = $2, exam_type = $3, start_time = $4,
			end_time = $5, duration = $6, total_marks = $7, passing_marks = $8, room_id = $9,
			status = $10, instructions = $11, allowed_materials = $12, question_paper_sets = $13
			WHERE id = $14 AND college_id = $15`

	result, err := r.db.Pool.Exec(ctx, sql,
		exam.Title, exam.Description, exam.ExamType, exam.StartTime, exam.EndTime,
		exam.Duration, exam.TotalMarks, exam.PassingMarks, exam.RoomID, exam.Status,
		exam.Instructions, exam.AllowedMaterials, exam.QuestionPaperSets,
		exam.ID, exam.CollegeID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("exam not found")
	}
	return nil
}

// DeleteExam deletes an exam
func (r *examRepository) DeleteExam(ctx context.Context, collegeID, examID int) error {
	sql := `DELETE FROM exams WHERE id = $1 AND college_id = $2`
	result, err := r.db.Pool.Exec(ctx, sql, examID, collegeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("exam not found")
	}
	return nil
}

// ListExamsByCourse retrieves exams for a specific course
func (r *examRepository) ListExamsByCourse(ctx context.Context, collegeID, courseID int, limit, offset int) ([]*models.Exam, error) {
	return r.ListExams(ctx, collegeID, map[string]any{"course_id": courseID}, limit, offset)
}

// EnrollStudent enrolls a student in an exam
func (r *examRepository) EnrollStudent(ctx context.Context, enrollment *models.ExamEnrollment) error {
	sql := `INSERT INTO exam_enrollments (exam_id, student_id, college_id, seat_number,
			room_number, question_paper_set, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, enrollment_date, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, sql,
		enrollment.ExamID, enrollment.StudentID, enrollment.CollegeID, enrollment.SeatNumber,
		enrollment.RoomNumber, enrollment.QuestionPaperSet, enrollment.Status,
	).Scan(&enrollment.ID, &enrollment.EnrollmentDate, &enrollment.CreatedAt, &enrollment.UpdatedAt)
}

// GetEnrollment retrieves an enrollment
func (r *examRepository) GetEnrollment(ctx context.Context, examID, studentID int) (*models.ExamEnrollment, error) {
	sql := `SELECT id, exam_id, student_id, college_id, enrollment_date, seat_number,
			room_number, question_paper_set, status, hall_ticket_generated, created_at, updated_at
			FROM exam_enrollments WHERE exam_id = $1 AND student_id = $2`

	enrollment := &models.ExamEnrollment{}
	err := r.db.Pool.QueryRow(ctx, sql, examID, studentID).Scan(
		&enrollment.ID, &enrollment.ExamID, &enrollment.StudentID, &enrollment.CollegeID,
		&enrollment.EnrollmentDate, &enrollment.SeatNumber, &enrollment.RoomNumber,
		&enrollment.QuestionPaperSet, &enrollment.Status, &enrollment.HallTicketGenerated,
		&enrollment.CreatedAt, &enrollment.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("enrollment not found: %w", err)
	}
	return enrollment, nil
}

// ListEnrollments retrieves all enrollments for an exam
func (r *examRepository) ListEnrollments(ctx context.Context, examID int) ([]*models.ExamEnrollment, error) {
	sql := `SELECT id, exam_id, student_id, college_id, enrollment_date, seat_number,
			room_number, question_paper_set, status, hall_ticket_generated, created_at, updated_at
			FROM exam_enrollments WHERE exam_id = $1 ORDER BY seat_number`

	rows, err := r.db.Pool.Query(ctx, sql, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []*models.ExamEnrollment
	for rows.Next() {
		enrollment := &models.ExamEnrollment{}
		err := rows.Scan(
			&enrollment.ID, &enrollment.ExamID, &enrollment.StudentID, &enrollment.CollegeID,
			&enrollment.EnrollmentDate, &enrollment.SeatNumber, &enrollment.RoomNumber,
			&enrollment.QuestionPaperSet, &enrollment.Status, &enrollment.HallTicketGenerated,
			&enrollment.CreatedAt, &enrollment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, nil
}

// UpdateEnrollment updates an enrollment
func (r *examRepository) UpdateEnrollment(ctx context.Context, enrollment *models.ExamEnrollment) error {
	sql := `UPDATE exam_enrollments SET seat_number = $1, room_number = $2,
			question_paper_set = $3, status = $4, hall_ticket_generated = $5
			WHERE id = $6`

	result, err := r.db.Pool.Exec(ctx, sql,
		enrollment.SeatNumber, enrollment.RoomNumber, enrollment.QuestionPaperSet,
		enrollment.Status, enrollment.HallTicketGenerated, enrollment.ID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("enrollment not found")
	}
	return nil
}

// DeleteEnrollment deletes an enrollment
func (r *examRepository) DeleteEnrollment(ctx context.Context, examID, studentID int) error {
	sql := `DELETE FROM exam_enrollments WHERE exam_id = $1 AND student_id = $2`
	result, err := r.db.Pool.Exec(ctx, sql, examID, studentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("enrollment not found")
	}
	return nil
}

// GetStudentEnrollments retrieves all enrollments for a student
func (r *examRepository) GetStudentEnrollments(ctx context.Context, studentID, collegeID int) ([]*models.ExamEnrollment, error) {
	sql := `SELECT id, exam_id, student_id, college_id, enrollment_date, seat_number,
			room_number, question_paper_set, status, hall_ticket_generated, created_at, updated_at
			FROM exam_enrollments WHERE student_id = $1 AND college_id = $2
			ORDER BY enrollment_date DESC`

	rows, err := r.db.Pool.Query(ctx, sql, studentID, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []*models.ExamEnrollment
	for rows.Next() {
		enrollment := &models.ExamEnrollment{}
		err := rows.Scan(
			&enrollment.ID, &enrollment.ExamID, &enrollment.StudentID, &enrollment.CollegeID,
			&enrollment.EnrollmentDate, &enrollment.SeatNumber, &enrollment.RoomNumber,
			&enrollment.QuestionPaperSet, &enrollment.Status, &enrollment.HallTicketGenerated,
			&enrollment.CreatedAt, &enrollment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, nil
}

// CreateResult creates an exam result
func (r *examRepository) CreateResult(ctx context.Context, result *models.ExamResult) error {
	sql := `INSERT INTO exam_results (exam_id, student_id, college_id, marks_obtained,
			grade, percentage, result, remarks, evaluated_by, evaluated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, sql,
		result.ExamID, result.StudentID, result.CollegeID, result.MarksObtained,
		result.Grade, result.Percentage, result.Result, result.Remarks,
		result.EvaluatedBy, result.EvaluatedAt,
	).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)
}

// GetResult retrieves a result
func (r *examRepository) GetResult(ctx context.Context, examID, studentID int) (*models.ExamResult, error) {
	sql := `SELECT id, exam_id, student_id, college_id, marks_obtained, grade, percentage,
			result, remarks, evaluated_by, evaluated_at, revaluation_status, created_at, updated_at
			FROM exam_results WHERE exam_id = $1 AND student_id = $2`

	res := &models.ExamResult{}
	err := r.db.Pool.QueryRow(ctx, sql, examID, studentID).Scan(
		&res.ID, &res.ExamID, &res.StudentID, &res.CollegeID, &res.MarksObtained,
		&res.Grade, &res.Percentage, &res.Result, &res.Remarks, &res.EvaluatedBy,
		&res.EvaluatedAt, &res.RevaluationStatus, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("result not found: %w", err)
	}
	return res, nil
}

// GetResultByID retrieves a result by its ID
func (r *examRepository) GetResultByID(ctx context.Context, resultID int) (*models.ExamResult, error) {
	sql := `SELECT id, exam_id, student_id, college_id, marks_obtained, grade, percentage,
			result, remarks, evaluated_by, evaluated_at, revaluation_status, created_at, updated_at
			FROM exam_results WHERE id = $1`

	res := &models.ExamResult{}
	err := r.db.Pool.QueryRow(ctx, sql, resultID).Scan(
		&res.ID, &res.ExamID, &res.StudentID, &res.CollegeID, &res.MarksObtained,
		&res.Grade, &res.Percentage, &res.Result, &res.Remarks, &res.EvaluatedBy,
		&res.EvaluatedAt, &res.RevaluationStatus, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("result not found: %w", err)
	}
	return res, nil
}

// ListResults retrieves all results for an exam
func (r *examRepository) ListResults(ctx context.Context, examID int) ([]*models.ExamResult, error) {
	sql := `SELECT id, exam_id, student_id, college_id, marks_obtained, grade, percentage,
			result, remarks, evaluated_by, evaluated_at, revaluation_status, created_at, updated_at
			FROM exam_results WHERE exam_id = $1 ORDER BY student_id`

	rows, err := r.db.Pool.Query(ctx, sql, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.ExamResult
	for rows.Next() {
		res := &models.ExamResult{}
		err := rows.Scan(
			&res.ID, &res.ExamID, &res.StudentID, &res.CollegeID, &res.MarksObtained,
			&res.Grade, &res.Percentage, &res.Result, &res.Remarks, &res.EvaluatedBy,
			&res.EvaluatedAt, &res.RevaluationStatus, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

// UpdateResult updates a result
func (r *examRepository) UpdateResult(ctx context.Context, result *models.ExamResult) error {
	sql := `UPDATE exam_results SET marks_obtained = $1, grade = $2, percentage = $3,
			result = $4, remarks = $5, evaluated_by = $6, evaluated_at = $7,
			revaluation_status = $8 WHERE id = $9`

	res, err := r.db.Pool.Exec(ctx, sql,
		result.MarksObtained, result.Grade, result.Percentage, result.Result,
		result.Remarks, result.EvaluatedBy, result.EvaluatedAt,
		result.RevaluationStatus, result.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("result not found")
	}
	return nil
}

// GetStudentResults retrieves all results for a student
func (r *examRepository) GetStudentResults(ctx context.Context, studentID, collegeID int) ([]*models.ExamResult, error) {
	sql := `SELECT id, exam_id, student_id, college_id, marks_obtained, grade, percentage,
			result, remarks, evaluated_by, evaluated_at, revaluation_status, created_at, updated_at
			FROM exam_results WHERE student_id = $1 AND college_id = $2 ORDER BY created_at DESC`

	rows, err := r.db.Pool.Query(ctx, sql, studentID, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.ExamResult
	for rows.Next() {
		res := &models.ExamResult{}
		err := rows.Scan(
			&res.ID, &res.ExamID, &res.StudentID, &res.CollegeID, &res.MarksObtained,
			&res.Grade, &res.Percentage, &res.Result, &res.Remarks, &res.EvaluatedBy,
			&res.EvaluatedAt, &res.RevaluationStatus, &res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

// CreateRevaluationRequest creates a revaluation request
func (r *examRepository) CreateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error {
	sql := `INSERT INTO revaluation_requests (exam_result_id, student_id, college_id,
			reason, previous_marks, requested_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			RETURNING id, status, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, sql,
		request.ExamResultID, request.StudentID, request.CollegeID,
		request.Reason, request.PreviousMarks,
	).Scan(&request.ID, &request.Status, &request.CreatedAt, &request.UpdatedAt)
}

// GetRevaluationRequest retrieves a revaluation request
func (r *examRepository) GetRevaluationRequest(ctx context.Context, requestID int) (*models.RevaluationRequest, error) {
	sql := `SELECT id, exam_result_id, student_id, college_id, reason, status,
			previous_marks, revised_marks, reviewed_by, review_comments,
			requested_at, reviewed_at, created_at, updated_at
			FROM revaluation_requests WHERE id = $1`

	req := &models.RevaluationRequest{}
	err := r.db.Pool.QueryRow(ctx, sql, requestID).Scan(
		&req.ID, &req.ExamResultID, &req.StudentID, &req.CollegeID, &req.Reason,
		&req.Status, &req.PreviousMarks, &req.RevisedMarks, &req.ReviewedBy,
		&req.ReviewComments, &req.RequestedAt, &req.ReviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}
	return req, nil
}

// ListRevaluationRequests retrieves revaluation requests with filters
func (r *examRepository) ListRevaluationRequests(ctx context.Context, collegeID int, filters map[string]any) ([]*models.RevaluationRequest, error) {
	sql := `SELECT id, exam_result_id, student_id, college_id, reason, status,
			previous_marks, revised_marks, reviewed_by, review_comments,
			requested_at, reviewed_at, created_at, updated_at
			FROM revaluation_requests WHERE college_id = $1`
	args := []any{collegeID}
	argCount := 1

	if status, ok := filters["status"]; ok {
		argCount++
		sql += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	if studentID, ok := filters["student_id"]; ok {
		argCount++
		sql += fmt.Sprintf(" AND student_id = $%d", argCount)
		args = append(args, studentID)
	}

	sql += " ORDER BY requested_at DESC"

	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.RevaluationRequest
	for rows.Next() {
		req := &models.RevaluationRequest{}
		err := rows.Scan(
			&req.ID, &req.ExamResultID, &req.StudentID, &req.CollegeID, &req.Reason,
			&req.Status, &req.PreviousMarks, &req.RevisedMarks, &req.ReviewedBy,
			&req.ReviewComments, &req.RequestedAt, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

// UpdateRevaluationRequest updates a revaluation request
func (r *examRepository) UpdateRevaluationRequest(ctx context.Context, request *models.RevaluationRequest) error {
	sql := `UPDATE revaluation_requests SET status = $1, revised_marks = $2,
			reviewed_by = $3, review_comments = $4, reviewed_at = $5 WHERE id = $6`

	res, err := r.db.Pool.Exec(ctx, sql,
		request.Status, request.RevisedMarks, request.ReviewedBy,
		request.ReviewComments, request.ReviewedAt, request.ID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("request not found")
	}
	return nil
}

// CreateRoom creates an exam room
func (r *examRepository) CreateRoom(ctx context.Context, room *models.ExamRoom) error {
	sql := `INSERT INTO exam_rooms (college_id, room_number, room_name, capacity,
			location, facilities, is_active)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`

	return r.db.Pool.QueryRow(ctx, sql,
		room.CollegeID, room.RoomNumber, room.RoomName, room.Capacity,
		room.Location, room.Facilities, room.IsActive,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
}

// GetRoomByID retrieves a room by ID
func (r *examRepository) GetRoomByID(ctx context.Context, collegeID, roomID int) (*models.ExamRoom, error) {
	sql := `SELECT id, college_id, room_number, room_name, capacity, location,
			facilities, is_active, created_at, updated_at
			FROM exam_rooms WHERE id = $1 AND college_id = $2`

	room := &models.ExamRoom{}
	err := r.db.Pool.QueryRow(ctx, sql, roomID, collegeID).Scan(
		&room.ID, &room.CollegeID, &room.RoomNumber, &room.RoomName, &room.Capacity,
		&room.Location, &room.Facilities, &room.IsActive, &room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("room not found: %w", err)
	}
	return room, nil
}

// ListRooms retrieves all rooms
func (r *examRepository) ListRooms(ctx context.Context, collegeID int, activeOnly bool) ([]*models.ExamRoom, error) {
	sql := `SELECT id, college_id, room_number, room_name, capacity, location,
			facilities, is_active, created_at, updated_at
			FROM exam_rooms WHERE college_id = $1`

	if activeOnly {
		sql += " AND is_active = true"
	}
	sql += " ORDER BY room_number"

	rows, err := r.db.Pool.Query(ctx, sql, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*models.ExamRoom
	for rows.Next() {
		room := &models.ExamRoom{}
		err := rows.Scan(
			&room.ID, &room.CollegeID, &room.RoomNumber, &room.RoomName, &room.Capacity,
			&room.Location, &room.Facilities, &room.IsActive, &room.CreatedAt, &room.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// UpdateRoom updates a room
func (r *examRepository) UpdateRoom(ctx context.Context, room *models.ExamRoom) error {
	sql := `UPDATE exam_rooms SET room_number = $1, room_name = $2, capacity = $3,
			location = $4, facilities = $5, is_active = $6 WHERE id = $7 AND college_id = $8`

	res, err := r.db.Pool.Exec(ctx, sql,
		room.RoomNumber, room.RoomName, room.Capacity, room.Location,
		room.Facilities, room.IsActive, room.ID, room.CollegeID,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("room not found")
	}
	return nil
}

// DeleteRoom deletes a room
func (r *examRepository) DeleteRoom(ctx context.Context, collegeID, roomID int) error {
	sql := `DELETE FROM exam_rooms WHERE id = $1 AND college_id = $2`
	res, err := r.db.Pool.Exec(ctx, sql, roomID, collegeID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("room not found")
	}
	return nil
}

// CheckRoomAvailability checks if a room is available for a time slot
func (r *examRepository) CheckRoomAvailability(ctx context.Context, roomID int, startTime, endTime string) (bool, error) {
	sql := `SELECT COUNT(*) FROM exams
			WHERE room_id = $1
			AND status NOT IN ('cancelled', 'completed')
			AND (
				(start_time <= $2 AND end_time >= $2) OR
				(start_time <= $3 AND end_time >= $3) OR
				(start_time >= $2 AND end_time <= $3)
			)`

	var count int
	err := r.db.Pool.QueryRow(ctx, sql, roomID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}
