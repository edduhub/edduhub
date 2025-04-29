package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAttendanceTest(t *testing.T) (pgxmock.PgxPoolIface, *DB, AttendanceRepository, context.Context) {
	// Create a new mock database connection
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	// Create a new DB instance with the mock connection
	db := &DB{
		Pool: mock,
		SQ:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	// Create a new attendance repository with the mock DB
	repo := NewAttendanceRepository(db)

	// Create a context for the tests
	ctx := context.Background()

	return mock, db, repo, ctx
}

func TestGetAttendanceByCourse(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	courseID := 2

	// Define expected rows
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, 101, 2, 1, time.Now(), "Present", time.Now(), 201).
		AddRow(2, 102, 2, 1, time.Now(), "Absent", time.Now(), 201)

	// Expect the query with specific arguments
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
		WithArgs(collegeID, courseID).
		WillReturnRows(rows)

	// Call the method
	attendances, err := repo.GetAttendanceByCourse(ctx, collegeID, courseID)

	// Assert no error occurred
	assert.NoError(t, err)
	assert.Len(t, attendances, 2)
	assert.Equal(t, 101, attendances[0].StudentID)
	assert.Equal(t, 102, attendances[1].StudentID)
	assert.Equal(t, "Present", attendances[0].Status)
	assert.Equal(t, "Absent", attendances[1].Status)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttendanceByCourse_Error(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	courseID := 2

	// Simulate a database error
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
		WithArgs(collegeID, courseID).
		WillReturnError(errors.New("database error"))

	// Call the method
	attendances, err := repo.GetAttendanceByCourse(ctx, collegeID, courseID)

	// Assert error occurred
	assert.Error(t, err)
	assert.Nil(t, attendances)
	assert.Contains(t, err.Error(), "failed to execute query or scan")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkAttendance(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2
	lectureID := 201

	// Expect the query with specific arguments (using a regex pattern since the exact SQL might vary)
	mock.ExpectExec(`INSERT INTO attendance`).
		WithArgs(studentID, courseID, collegeID, lectureID, pgxmock.AnyArg(), "Present", pgxmock.AnyArg()).
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

	// Simulate a database error
	mock.ExpectExec(`INSERT INTO attendance`).
		WithArgs(studentID, courseID, collegeID, lectureID, pgxmock.AnyArg(), "Present", pgxmock.AnyArg()).
		WillReturnError(errors.New("database error"))

	// Call the method
	success, err := repo.MarkAttendance(ctx, collegeID, studentID, courseID, lectureID)

	// Assert error occurred and operation failed
	assert.Error(t, err)
	assert.False(t, success)
	assert.Contains(t, err.Error(), "failed to execute query")

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

	// Expect the query with specific arguments
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE`).
		WithArgs(status, collegeID, studentID, courseID, lectureID).
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

	// Simulate no rows affected
	mock.ExpectExec(`UPDATE attendance SET status = \$1 WHERE`).
		WithArgs(status, collegeID, studentID, courseID, lectureID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	// Call the method
	err := repo.UpdateAttendance(ctx, collegeID, studentID, courseID, lectureID, status)

	// Assert error occurred
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "did not update attendance")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttendanceStudentInCourse(t *testing.T) {
	mock, _, repo, ctx := setupAttendanceTest(t)
	defer mock.Close()

	collegeID := 1
	studentID := 101
	courseID := 2

	// Define expected rows
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, studentID, courseID, collegeID, time.Now(), "Present", time.Now(), 201).
		AddRow(2, studentID, courseID, collegeID, time.Now().Add(24*time.Hour), "Absent", time.Now().Add(24*time.Hour), 202)

	// Expect the query with specific arguments
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
		WithArgs(collegeID, studentID, courseID).
		WillReturnRows(rows)

	// Call the method
	attendances, err := repo.GetAttendanceStudentInCourse(ctx, collegeID, studentID, courseID)

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

	// Define expected rows
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, studentID, 2, collegeID, time.Now(), "Present", time.Now(), 201).
		AddRow(2, studentID, 3, collegeID, time.Now(), "Absent", time.Now(), 301)

	// Expect the query with specific arguments
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
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

	// Define expected rows
	rows := pgxmock.NewRows([]string{
		"id", "student_id", "course_id", "college_id", "date", "status", "scanned_at", "lecture_id",
	}).
		AddRow(1, 101, courseID, collegeID, time.Now(), "Present", time.Now(), lectureID).
		AddRow(2, 102, courseID, collegeID, time.Now(), "Absent", time.Now(), lectureID)

	// Expect the query with specific arguments
	mock.ExpectQuery(`SELECT id, student_id, course_id, college_id, date, status, scanned_at, lecture_id FROM attendance WHERE`).
		WithArgs(collegeID, lectureID, courseID).
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
