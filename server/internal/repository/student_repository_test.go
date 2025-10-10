//go:build integration
// +build integration

package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStudentTest(t *testing.T) (pgxmock.PgxPoolIface, *DB, StudentRepository, context.Context) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})

	db := &DB{
		Pool: mock,
	}

	repo := NewStudentRepository(db)
	ctx := context.Background()

	return mock, db, repo, ctx
}

func TestCreateStudent(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	student := &models.Student{
		UserID:           10,
		CollegeID:        1,
		KratosIdentityID: "kratos-test-id",
		EnrollmentYear:   2023,
		RollNo:           "TEST001",
		IsActive:         true,
	}
	expectedID := 5

	sqlRegex := `INSERT INTO students \(user_id,college_id,kratos_identity_id,enrollment_year,roll_no,is_active,created_at,updated_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING id`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(
			student.UserID,
			student.CollegeID,
			student.KratosIdentityID,
			student.EnrollmentYear,
			student.RollNo,
			student.IsActive,
			pgxmock.AnyArg(), // CreatedAt
			pgxmock.AnyArg(), // UpdatedAt
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID))

	err := repo.CreateStudent(ctx, student)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, student.StudentID) // Ensure the ID field name matches your struct
	assert.False(t, student.CreatedAt.IsZero())
	assert.False(t, student.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateStudent_Error(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	student := &models.Student{RollNo: "FAIL001"} // Minimal data for error case
	dbError := errors.New("insert failed")
	sqlRegex := `INSERT INTO students \(user_id,college_id,kratos_identity_id,enrollment_year,roll_no,is_active,created_at,updated_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\) RETURNING id`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(student.UserID, student.CollegeID, student.KratosIdentityID, student.EnrollmentYear, student.RollNo, student.IsActive, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(dbError)

	err := repo.CreateStudent(ctx, student)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query or scan ID")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Test GetStudentByRollNo ---

func TestGetStudentByRollNo(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "TEST001"
	expectedStudent := &models.Student{
		StudentID:        5,
		UserID:           10,
		CollegeID:        1,
		KratosIdentityID: "kratos-test-id",
		EnrollmentYear:   2023,
		RollNo:           rollNo,
		IsActive:         true,
		CreatedAt:        time.Now().Add(-time.Hour),
		UpdatedAt:        time.Now(),
	}

	sqlRegex := `SELECT id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at FROM students WHERE roll_no = \$1`
	rows := pgxmock.NewRows([]string{"id", "user_id", "college_id", "kratos_identity_id", "enrollment_year", "roll_no", "is_active", "created_at", "updated_at"}).
		AddRow(expectedStudent.StudentID, expectedStudent.UserID, expectedStudent.CollegeID, expectedStudent.KratosIdentityID, expectedStudent.EnrollmentYear, expectedStudent.RollNo, expectedStudent.IsActive, expectedStudent.CreatedAt, expectedStudent.UpdatedAt)

	// pgxscan.Get uses QueryRow internally
	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(rollNo).
		WillReturnRows(rows)

	student, err := repo.GetStudentByRollNo(ctx, rollNo)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedStudent, student) // Use EqualValues for struct comparison
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetStudentByRollNo_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "NOTFOUND"
	sqlRegex := `SELECT id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at FROM students WHERE roll_no = \$1`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(rollNo).
		WillReturnError(pgx.ErrNoRows)

	student, err := repo.GetStudentByRollNo(ctx, rollNo)

	assert.NoError(t, err) // Implementation returns nil, nil for not found
	assert.Nil(t, student)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Test GetStudentByID ---
// Similar tests for GetStudentByID, filtering by ID and CollegeID

// --- Test UpdateStudent ---

func TestUpdateStudent(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	studentToUpdate := &models.Student{
		StudentID:        5,
		UserID:           10,
		CollegeID:        1,
		KratosIdentityID: "kratos-updated-id", // Changed
		EnrollmentYear:   2023,
		RollNo:           "TEST001-UPDATED", // Changed
		IsActive:         false,             // Changed
		// UpdatedAt will be set by the method
	}

	sqlRegex := `UPDATE students SET user_id = \$1, college_id = \$2, kratos_identity_id = \$3, enrollment_year = \$4, roll_no = \$5, is_active = \$6, updated_at = \$7 WHERE id = \$8`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(
			studentToUpdate.UserID,
			studentToUpdate.CollegeID,
			studentToUpdate.KratosIdentityID,
			studentToUpdate.EnrollmentYear,
			studentToUpdate.RollNo,
			studentToUpdate.IsActive,
			pgxmock.AnyArg(), // UpdatedAt
			studentToUpdate.StudentID,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateStudent(ctx, studentToUpdate)

	assert.NoError(t, err)
	assert.False(t, studentToUpdate.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateStudent_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	studentToUpdate := &models.Student{StudentID: 999, RollNo: "NOTFOUND"} // Non-existent ID
	sqlRegex := `UPDATE students SET user_id = \$1, college_id = \$2, kratos_identity_id = \$3, enrollment_year = \$4, roll_no = \$5, is_active = \$6, updated_at = \$7 WHERE id = \$8`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(studentToUpdate.UserID, studentToUpdate.CollegeID, studentToUpdate.KratosIdentityID, studentToUpdate.EnrollmentYear, studentToUpdate.RollNo, studentToUpdate.IsActive, pgxmock.AnyArg(), studentToUpdate.StudentID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0)) // Simulate 0 rows affected

	err := repo.UpdateStudent(ctx, studentToUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no row updated")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Test FreezeStudent ---

func TestFreezeStudent(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "TEST001"
	sqlRegex := `UPDATE students SET is_active = \$1, updated_at = \$2 WHERE roll_no = \$3`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(false, pgxmock.AnyArg(), rollNo).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.FreezeStudent(ctx, rollNo)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFreezeStudent_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "NOTFOUND"
	sqlRegex := `UPDATE students SET is_active = \$1, updated_at = \$2 WHERE roll_no = \$3`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(false, pgxmock.AnyArg(), rollNo).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.FreezeStudent(ctx, rollNo)

	// Current implementation does NOT return an error here. Adjust if implementation changes.
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Test UnFreezeStudent ---

func TestUnFreezeStudent(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "TEST001"
	sqlRegex := `UPDATE students SET is_active = \$1, updated_at = \$2 WHERE roll_no = \$3`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(true, pgxmock.AnyArg(), rollNo).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UnFreezeStudent(ctx, rollNo)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnFreezeStudent_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	rollNo := "NOTFOUND"
	sqlRegex := `UPDATE students SET is_active = \$1, updated_at = \$2 WHERE roll_no = \$3`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(true, pgxmock.AnyArg(), rollNo).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.UnFreezeStudent(ctx, rollNo)

	// Current implementation does NOT return an error here. Adjust if implementation changes.
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Test FindByKratosID ---

func TestFindByKratosID(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	kratosID := "kratos-test-id"
	expectedStudent := &models.Student{ /* ... fill expected data ... */ RollNo: "TEST001", KratosIdentityID: kratosID}

	sqlRegex := `SELECT id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at FROM students WHERE kratos_identity_id = \$1`
	rows := pgxmock.NewRows([]string{"id", "user_id", "college_id", "kratos_identity_id", "enrollment_year", "roll_no", "is_active", "created_at", "updated_at"}).
		AddRow(expectedStudent.StudentID, expectedStudent.UserID, expectedStudent.CollegeID, expectedStudent.KratosIdentityID, expectedStudent.EnrollmentYear, expectedStudent.RollNo, expectedStudent.IsActive, expectedStudent.CreatedAt, expectedStudent.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(kratosID).
		WillReturnRows(rows)

	student, err := repo.FindByKratosID(ctx, kratosID)

	assert.NoError(t, err)
	assert.EqualValues(t, expectedStudent, student)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByKratosID_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupStudentTest(t)
	defer mock.Close()

	kratosID := "kratos-not-found"
	sqlRegex := `SELECT id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at FROM students WHERE kratos_identity_id = \$1`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(kratosID).
		WillReturnError(pgx.ErrNoRows)

	student, err := repo.FindByKratosID(ctx, kratosID)

	assert.NoError(t, err) // Implementation returns nil, nil
	assert.Nil(t, student)
	assert.NoError(t, mock.ExpectationsWereMet())
}
