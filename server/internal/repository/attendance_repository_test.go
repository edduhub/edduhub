package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAttendanceTest(t *testing.T) (*pgxpool.Pool, AttendanceRepository, context.Context) {
	// Create a test database connection for pgx-based testing
	databaseURL := "postgres://your_db_user:your_db_password@localhost:5432/edduhub"

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)

	// Clean up after test
	t.Cleanup(func() {
		pool.Close()
	})

	// Create a new attendance repository with the pool
	repo := NewAttendanceRepository(pool)

	// Create a context for the tests
	ctx := context.Background()

	return pool, repo, ctx
}

func TestGetAttendanceByCourse(t *testing.T) {
	pool, repo, ctx := setupAttendanceTest(t)
	defer pool.Close()

	collegeID := 1
	courseID := 2

	// Call the method with pagination
	attendances, err := repo.GetAttendanceByCourse(ctx, collegeID, courseID, 10, 0)

	// Assert no error occurred
	assert.NoError(t, err)
	// Note: Actual assertions would depend on test data in the database
	// For this example, we'll assume the database has test data
	if len(attendances) > 0 {
		assert.Equal(t, collegeID, attendances[0].CollegeID)
		assert.Equal(t, courseID, attendances[0].CourseID)
	}
}

func TestGetAttendanceByCourse_Error(t *testing.T) {
	mock, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	courseID := 2

	// Simulate a database error
	mock.ExpectQuery(`SELECT  id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
		WithArgs(collegeID, courseID).
		WillReturnError(errors.New("database error"))

	// Call the method
	attendances, err := repo.GetAttendanceByCourse(ctx, collegeID, courseID, 10, 0)

	// Assert error occurred
	assert.Error(t, err)
	assert.Nil(t, attendances)
	assert.Contains(t, err.Error(), "failed to execute query or scan")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkAttendance(t *testing.T) {
	mock, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2
	lectureID := 201

	mock.ExpectExec(`INSERT INTO attendance \(student_id,course_id,college_id,lecture_id,date,status,scanned_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) ON CONFLICT \(student_id, course_id, lecture_id, date, college_id\) DO UPDATE SET scanned_at = EXCLUDED.scanned_at, status = EXCLUDED.status`).
		WithArgs(studentID, courseID, collegeID, lectureID, pgxmock.AnyArg(), "Present", pgxmock.AnyArg()). // 7 Args, matching the VALUES
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Call the method
	success, err := repo.MarkAttendance(ctx, collegeID, studentID, courseID, lectureID)

	// Assert no error occurred and operation was successful
	assert.NoError(t, err)
	assert.True(t, success)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkAttendance_Error(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2
	lectureID := 201

	// Simulate a database error using the same refined regex pattern
	mock.ExpectExec(`INSERT INTO attendance \(student_id,course_id,college_id,lecture_id,date,status,scanned_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) ON CONFLICT \(student_id, course_id, lecture_id, date, college_id\) DO UPDATE SET scanned_at = EXCLUDED.scanned_at, status = EXCLUDED.status`).
		WithArgs(studentID, courseID, collegeID, lectureID, pgxmock.AnyArg(), "Present", pgxmock.AnyArg()). // 7 Args
		WillReturnError(errors.New("database error"))

	// Call the method
	success, err := repo.MarkAttendance(ctx, collegeID, studentID, courseID, lectureID)

	// Assert error occurred and operation failed
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "failed to execute query")
	//
	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAttendance(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2
	lectureID := 201
	status := "Absent"

	// Expect the query with specific arguments in the correct order
	// Actual order seems to be: status, collegeID, courseID, lectureID, studentID
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE college_id = \$2 AND course_id = \$3 AND lecture_id = \$4 AND student_id = \$5`). // Match the actual WHERE clause order if known
																			WithArgs(status, collegeID, courseID, lectureID, studentID). // Correct argument order
																			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	// Call the method
	err := repo.UpdateAttendance(ctx, collegeID, studentID, courseID, lectureID, status)

	// Assert no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAttendance_NoRowsAffected(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2
	lectureID := 201
	status := "Absent"

	// Simulate no rows affected with correct argument order
	// Actual order seems to be: status, collegeID, courseID, lectureID, studentID
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE college_id = \$2 AND course_id = \$3 AND lecture_id = \$4 AND student_id = \$5`). // Match the actual WHERE clause order if known
																			WithArgs(status, collegeID, courseID, lectureID, studentID). // Correct argument order
																			WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Call the method
	err := repo.UpdateAttendance(ctx, collegeID, studentID, courseID, lectureID, status)

	// Assert error occurred
	assert.Error(t, err)
	// Update the expected error message if the underlying function changed it
	assert.Contains(t, err.Error(), "did not update attendance") // Or the actual error message from repo

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttendanceStudentInCourse(t *testing.T) {
	mock, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2

	// Define expected rows (without 'id')
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, studentID, courseID, collegeID, time.Now(), "Present", time.Now(), 201).
		AddRow(2, studentID, courseID, collegeID, time.Now().Add(24*time.Hour), "Absent", time.Now().Add(24*time.Hour), 202)

	// Expect the query matching the actual WHERE clause order and argument order
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE college_id = \$1 AND course_id = \$2 AND student_id = \$3 ORDER BY scanned_at ASC`). // Correct WHERE clause order
																											WithArgs(collegeID, courseID, studentID). // Correct argument order
																											WillReturnRows(rows)

	// Call the method
	attendances, err := repo.GetAttendanceStudentInCourse(ctx, collegeID, studentID, courseID, 10, 0)

	// Assert no error occurred
	assert.NoError(t, err)
	assert.Len(t, attendances, 2)
	assert.Equal(t, studentID, attendances[0].StudentID)
	assert.Equal(t, courseID, attendances[0].CourseID)
	assert.Equal(t, "Present", attendances[0].Status)
	assert.Equal(t, "Absent", attendances[1].Status)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttendanceStudent(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101

	// Define expected rows (without 'id')
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, studentID, 2, collegeID, time.Now(), "Present", time.Now(), 201).
		AddRow(2, studentID, 3, collegeID, time.Now(), "Absent", time.Now(), 301)

	// Expect the query without 'id' and with ordering
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE college_id = \$1 AND student_id = \$2 ORDER BY date ASC, course_id ASC, scanned_at ASC`).
		WithArgs(collegeID, studentID).
		WillReturnRows(rows)

	// Call the method
	attendances, err := repo.GetAttendanceStudent(ctx, collegeID, studentID)

	// Assert no error occurred
	assert.NoError(t, err)
	assert.Len(t, attendances, 2)
	assert.Equal(t, studentID, attendances[0].StudentID)
	assert.Equal(t, 2, attendances[0].CourseID)
	assert.Equal(t, 3, attendances[1].CourseID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttendanceByLecture(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	lectureID := 201
	courseID := 2

	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, 101, courseID, collegeID, time.Now(), "Present", time.Now(), lectureID).
		AddRow(2, 102, courseID, collegeID, time.Now(), "Absent", time.Now(), lectureID)

	// Expect the query matching the actual WHERE clause order and argument order
	mock.ExpectQuery(`SELECT id,  student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE college_id = \$1 AND course_id = \$2 AND lecture_id = \$3 ORDER BY student_id ASC, scanned_at ASC`). // Correct WHERE clause order
																														WithArgs(collegeID, courseID, lectureID). // Correct argument order
																														WillReturnRows(rows)

	// Call the method
	attendances, err := repo.GetAttendanceByLecture(ctx, collegeID, lectureID, courseID)

	// Assert no error occurred
	assert.NoError(t, err)
	assert.Len(t, attendances, 2)
	assert.Equal(t, lectureID, attendances[0].LectureID)
	assert.Equal(t, lectureID, attendances[1].LectureID)
	assert.Equal(t, 101, attendances[0].StudentID)
	assert.Equal(t, 102, attendances[1].StudentID)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFreezeAttendance(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101

	// Expect the query with specific arguments
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE`).
		WithArgs("Frozen", collegeID, studentID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 2))

	// Call the method
	err := repo.FreezeAttendance(ctx, collegeID, studentID)

	// Assert no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFreezeAttendance_NoRowsAffected(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101

	// Simulate no rows affected
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE`).
		WithArgs("Frozen", collegeID, studentID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Call the method
	err := repo.FreezeAttendance(ctx, collegeID, studentID)

	// Assert error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unabel to freeze attendance")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFreezeAttendance_Error(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101

	// Simulate a database error
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE`).
		WithArgs("Frozen", collegeID, studentID).
		WillReturnError(errors.New("database error"))

	// Call the method
	err := repo.FreezeAttendance(ctx, collegeID, studentID)

	// Assert error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
