package repository

import (
	"context"
	"fmt" // Import fmt for better error wrapping
	"time"

	// Assuming models.Attendance uses time.Time
	"eduhub/server/internal/models" // Your models package
)

// Import the sqlc generated code
import (
	"eduhub/server/internal/repository/db"
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

	// Count methods (add corresponding count methods if needed)
	// ProcessQRCode(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int) (bool, error)
	// SetAttendanceStatus(ctx context.Context, collegeID int, studentID, courseID int, lectureID int, status string) error
}

const attendanceTable = "attendance"

type attendanceRepository struct {
	DB *DB // Assuming DB struct is accessible here
	*db.Queries
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

func NewAttendanceRepository(DB *DB) AttendanceRepository {
	return &attendanceRepository{
		DB: DB,
		Queries: db.New(DB.Pool),
	}
}

func (a *attendanceRepository) GetAttendanceByCourse(
	ctx context.Context,
	collegeID int,
	courseID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	// Convert parameters to int32 as expected by sqlc
	params := db.GetAttendanceByCourseParams{
		CollegeID: int32(collegeID),
		CourseID:  int32(courseID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	// Call the sqlc generated method
	rows, err := a.Queries.GetAttendanceByCourse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceByCourse: failed to execute query: %w", err)
	}

	// Convert sqlc models to our models
	attendances := make([]*models.Attendance, len(rows))
	for i, row := range rows {
		attendances[i] = &models.Attendance{
			ID:        int(row.ID),
			StudentID: int(row.StudentID),
			CourseID:  int(row.CourseID),
			CollegeID: int(row.CollegeID),
			Date:      row.Date,
			Status:    row.Status,
			ScannedAt: row.ScannedAt,
			LectureID: int(row.LectureID),
		}
	}

	return attendances, nil
}

func (a *attendanceRepository) MarkAttendance(ctx context.Context, collegeID int, studentID, courseID int, lectureID int) (bool, error) {
	now := time.Now()
	// Truncate date for the 'date' column if you only store the date part
	attendanceDate := now.Truncate(24 * time.Hour)

	// Convert parameters to int32 as expected by sqlc
	params := db.MarkAttendanceParams{
		StudentID: int32(studentID),
		CourseID:  int32(courseID),
		CollegeID: int32(collegeID),
		LectureID: int32(lectureID),
		Date:      attendanceDate,
		Status:    "Present", // Default status when marked
		ScannedAt: now,
	}

	// Call the sqlc generated method
	_, err := a.Queries.MarkAttendance(ctx, params)
	if err != nil {
		return false, fmt.Errorf("MarkAttendance: failed to execute query: %w", err)
	}

	// If no error occurred, the attendance was marked successfully
	return true, nil
}

func (a *attendanceRepository) UpdateAttendance(ctx context.Context, collegeID int, studentID int, courseID int, lectureID int, status string) error {
	// Convert parameters to int32 as expected by sqlc
	params := db.UpdateAttendanceParams{
		Status:    status,
		CollegeID: int32(collegeID),
		StudentID: int32(studentID),
		CourseID:  int32(courseID),
		LectureID: int32(lectureID),
	}

	// Call the sqlc generated method
	err := a.Queries.UpdateAttendance(ctx, params)
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
	// Convert parameters to int32 as expected by sqlc
	params := db.GetAttendanceStudentInCourseParams{
		CollegeID: int32(collegeID),
		StudentID: int32(studentID),
		CourseID:  int32(courseID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	// Call the sqlc generated method
	rows, err := a.Queries.GetAttendanceStudentInCourse(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceStudentInCourse: failed to execute query: %w", err)
	}

	// Convert sqlc models to our models
	attendances := make([]*models.Attendance, len(rows))
	for i, row := range rows {
		attendances[i] = &models.Attendance{
			ID:        int(row.ID),
			StudentID: int(row.StudentID),
			CourseID:  int(row.CourseID),
			CollegeID: int(row.CollegeID),
			Date:      row.Date,
			Status:    row.Status,
			ScannedAt: row.ScannedAt,
			LectureID: int(row.LectureID),
		}
	}

	return attendances, nil
}

func (a *attendanceRepository) GetAttendanceStudent(
	ctx context.Context,
	collegeID int,
	studentID int,
	limit, offset uint64, // Added pagination params
) ([]*models.Attendance, error) {
	// Convert parameters to int32 as expected by sqlc
	params := db.GetAttendanceStudentParams{
		CollegeID: int32(collegeID),
		StudentID: int32(studentID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	// Call the sqlc generated method
	rows, err := a.Queries.GetAttendanceStudent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceStudent: failed to execute query: %w", err)
	}

	// Convert sqlc models to our models
	attendances := make([]*models.Attendance, len(rows))
	for i, row := range rows {
		attendances[i] = &models.Attendance{
			ID:        int(row.ID),
			StudentID: int(row.StudentID),
			CourseID:  int(row.CourseID),
			CollegeID: int(row.CollegeID),
			Date:      row.Date,
			Status:    row.Status,
			ScannedAt: row.ScannedAt,
			LectureID: int(row.LectureID),
		}
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
	// Convert parameters to int32 as expected by sqlc
	params := db.GetAttendanceByLectureParams{
		CollegeID: int32(collegeID),
		LectureID: int32(lectureID),
		CourseID:  int32(courseID),
		Limit:     int32(limit),
		Offset:    int32(offset),
	}

	// Call the sqlc generated method
	rows, err := a.Queries.GetAttendanceByLecture(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("GetAttendanceByLecture: failed to execute query: %w", err)
	}

	// Convert sqlc models to our models
	attendances := make([]*models.Attendance, len(rows))
	for i, row := range rows {
		attendances[i] = &models.Attendance{
			ID:        int(row.ID),
			StudentID: int(row.StudentID),
			CourseID: int(row.CourseID),
			CollegeID: int(row.CollegeID),
			Date:      row.Date,
			Status:    row.Status,
			ScannedAt: row.ScannedAt,
			LectureID: int(row.LectureID),
		}
	}

	return attendances, nil
}

// FreezeAttendance updates the status of all attendance records for a specific student to "Frozen".
// This is a simple example; actual freezing logic might be more complex (e.g., only for past dates).
func (a *attendanceRepository) FreezeAttendance(ctx context.Context, collegeID int, studentID int) error {
	// Convert parameters to int32 as expected by sqlc
	params := db.FreezeAttendanceParams{
		Status:    "Frozen",
		CollegeID: int32(collegeID),
		StudentID: int32(studentID),
	}

	// Call the sqlc generated method
	err := a.Queries.FreezeAttendance(ctx, params)
	if err != nil {
		return fmt.Errorf("FreezeAttendance: failed to execute query: %w", err)
	}

	return nil
}

// UnFreezeAttendance updates the status of "Frozen" attendance records for a student back to a default (e.g., "Absent").
func (a *attendanceRepository) UnFreezeAttendance(ctx context.Context, collegeID int, studentID int) error {
	// Determine the status to revert to. "Absent" might be a safe default if the original status isn't stored.
	revertStatus := "Absent" // Or fetch original status if stored elsewhere

	// Convert parameters to int32 as expected by sqlc
	params := db.UnFreezeAttendanceParams{
		Status:    revertStatus,
		CollegeID: int32(collegeID),
		StudentID: int32(studentID),
		Status_2: "Frozen", // Only unfreeze records that are currently frozen
	}

	// Call the sqlc generated method
	err := a.Queries.UnFreezeAttendance(ctx, params)
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

	// Convert parameters to int32 as expected by sqlc
	params := db.SetAttendanceStatusParams{
		StudentID: int32(studentID),
		CourseID:  int32(courseID),
		CollegeID: int32(collegeID),
		LectureID: int32(lectureID),
		Status:    status,
		ScannedAt: now,
	}

	// Call the sqlc generated method
	err := a.Queries.SetAttendanceStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("SetAttendanceStatus: failed to execute query: %w", err)
	}

	return nil
}
