package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduhub/server/internal/models"
)

type AttendanceRepository interface {
	MarkAttendance(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int) (bool, error)
	UpdateAttendance(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int, status string) error
	SetAttendanceStatus(ctx context.Context, collegeID int, studentID, courseID int, lectureID int, status string) error
	FreezeAttendance(ctx context.Context, collegeID int, studentID int) error
	UnFreezeAttendance(ctx context.Context, collegeID int, studentID int) error

	// Get methods with pagination
	GetAttendanceByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceStudentInCourse(ctx context.Context, collegeID int, studentID int, courseID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceStudent(ctx context.Context, collegeID int, studentID int, limit, offset uint64) ([]*models.Attendance, error)
	GetAttendanceByLecture(ctx context.Context, collegeID int, lectureID int, courseID int, limit, offset uint64) ([]*models.Attendance, error)
}

const attendanceTable = "attendance"

type attendanceRepository struct {
	Pool *pgxpool.Pool // Direct pgxpool connection
}

// Assuming models.Attendance struct with db tags is defined in models package:
// type Attendance struct {
//     ID        int       `db:"id"`
//     StudentID int       `db:"student_id"`
//     CourseID  int       `db:"course_id"`
//     CollegeID int       `db:"college_id"`
//     Date      time.Time `db:"date"`
//     Status    string    `db:"status"`
//     ScannedAt time.Time `db:"scanned_at"`
//     LectureID int       `db:"lecture_id"`
// }

func NewAttendanceRepository(pool *pgxpool.Pool) AttendanceRepository {
	return &attendanceRepository{
		Pool: pool,
	}
}

func (a *attendanceRepository) GetAttendanceByCourse(
	ctx context.Context,
	collegeID int,
	courseID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	sql := `SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND course_id = $2
ORDER BY date DESC, student_id ASC
LIMIT $3 OFFSET $4`

	attendances := make([]*models.Attendance, 0)
	err := pgxscan.Select(ctx, a.Pool, &attendances, sql, int32(collegeID), int32(courseID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceByCourse: failed to scan: %w", err)
	}

	return attendances, nil
}

func (a *attendanceRepository) MarkAttendance(ctx context.Context, collegeID int, studentID, courseID int, lectureID int) (bool, error) {
	now := time.Now()
	// Truncate date for the 'date' column if you only store the date part
	attendanceDate := now.Truncate(24 * time.Hour)

	sql := `INSERT INTO attendance (
    	student_id,
    	course_id,
    	college_id,
    	lecture_id,
    	date,
    	status,
    	scanned_at
	) VALUES (
    	$1, $2, $3, $4, $5, $6, $7
	) ON CONFLICT (student_id, course_id, lecture_id, date, college_id)
	DO UPDATE SET scanned_at = EXCLUDED.scanned_at, status = EXCLUDED.status
	RETURNING *`

	var result models.Attendance
	err := pgxscan.Get(ctx, a.Pool, &result, sql, int32(studentID), int32(courseID), int32(collegeID), int32(lectureID), attendanceDate, "Present", now)
	if err != nil {
		return false, fmt.Errorf("MarkAttendance: failed to execute query: %w", err)
	}

	return true, nil
}

func (a *attendanceRepository) UpdateAttendance(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int, status string) error {
	sql := `UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3 AND course_id = $4 AND lecture_id = $5`

	_, err := a.Pool.Exec(ctx, sql, status, int32(collegeID), int32(studentID), int32(courseID), int32(lectureID))
	if err != nil {
		return fmt.Errorf("UpdateAttendance: failed to execute query: %w", err)
	}

	return nil
}

func (a *attendanceRepository) GetAttendanceStudentInCourse(
	ctx context.Context,
	collegeID int,
	studentID int,
	courseID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	sql := `SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND student_id = $2 AND course_id = $3
ORDER BY date DESC, scanned_at DESC
LIMIT $4 OFFSET $5`

	attendances := make([]*models.Attendance, 0)
	err := pgxscan.Select(ctx, a.Pool, &attendances, sql, int32(collegeID), int32(studentID), int32(courseID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceStudentInCourse: failed to scan: %w", err)
	}

	return attendances, nil
}

func (a *attendanceRepository) GetAttendanceStudent(
	ctx context.Context,
	collegeID int,
	studentID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	sql := `SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND student_id = $2
ORDER BY date DESC, course_id ASC, scanned_at DESC
LIMIT $3 OFFSET $4`

	attendances := make([]*models.Attendance, 0)
	err := pgxscan.Select(ctx, a.Pool, &attendances, sql, int32(collegeID), int32(studentID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceStudent: failed to scan: %w", err)
	}

	return attendances, nil
}

// GetAttendanceByLecture retrieves attendance records for a specific lecture.
func (a *attendanceRepository) GetAttendanceByLecture(
	ctx context.Context,
	collegeID int,
	lectureID int,
	courseID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	sql := `SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id
FROM attendance
WHERE college_id = $1 AND lecture_id = $2 AND course_id = $3
ORDER BY student_id ASC, scanned_at ASC
LIMIT $4 OFFSET $5`

	attendances := make([]*models.Attendance, 0)
	err := pgxscan.Select(ctx, a.Pool, &attendances, sql, int32(collegeID), int32(lectureID), int32(courseID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceByLecture: failed to scan: %w", err)
	}

	return attendances, nil
}

// FreezeAttendance updates the status of all attendance records for a specific student to "Frozen".
// This is a simple example; actual freezing logic might be more complex (e.g., only for past dates).
func (a *attendanceRepository) FreezeAttendance(ctx context.Context, collegeID int, studentID int) error {
	sql := `UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3`

	_, err := a.Pool.Exec(ctx, sql, "Frozen", int32(collegeID), int32(studentID))
	if err != nil {
		return fmt.Errorf("FreezeAttendance: failed to execute query: %w", err)
	}

	return nil
}

// UnFreezeAttendance updates the status of "Frozen" attendance records for a student back to a default (e.g., "Absent").
func (a *attendanceRepository) UnFreezeAttendance(ctx context.Context, collegeID int, studentID int) error {
	// Determine the status to revert to. "Absent" might be a safe default if the original status isn't stored.
	revertStatus := "Absent" // Or fetch original status if stored elsewhere

	sql := `UPDATE attendance
SET status = $1
WHERE college_id = $2 AND student_id = $3 AND status = $4`

	_, err := a.Pool.Exec(ctx, sql, revertStatus, int32(collegeID), int32(studentID), "Frozen")
	if err != nil {
		return fmt.Errorf("UnFreezeAttendance: failed to execute query: %w", err)
	}

	return nil
}

// type Attendance struct {
// 	ID        int       `json:"ID"`
// 	StudentID int       `json:"studentID"`
// 	CourseID  int       `json:"courseId"`
// 	CollegeID int       `json:"collegeID"`
// 	Date      time.Time `json:"date"`
// 	Status    string    `json:"status"`
// 	ScannedAt time.Time `json:"scannedAt"`
// 	LectureID int       `json:"lectureID"`
// }

func (a *attendanceRepository) SetAttendanceStatus(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int, status string) error {
	now := time.Now()

	sql := `INSERT INTO attendance (
    student_id,
    course_id,
    college_id,
    lecture_id,
    status,
    scanned_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) ON CONFLICT (student_id, course_id, college_id, lecture_id)
DO UPDATE SET status = EXCLUDED.status, scanned_at = EXCLUDED.scanned_at`

	_, err := a.Pool.Exec(ctx, sql, int32(studentID), int32(courseID), int32(collegeID), int32(lectureID), status, now)
	if err != nil {
		return fmt.Errorf("SetAttendanceStatus: failed to execute query: %w", err)
	}

	return nil
}
