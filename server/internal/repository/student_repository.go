package repository

import (
	"context"
	sqlDriver "database/sql"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduhub/server/internal/models"
)

type StudentRepository interface {
	CreateStudent(ctx context.Context, student *models.Student) error
	GetStudentByRollNo(ctx context.Context, collegeID int, rollNo string) (*models.Student, error)
	GetStudentByID(ctx context.Context, collegeID int, studentID int) (*models.Student, error)
	UpdateStudent(ctx context.Context, model *models.Student) error
	FreezeStudent(ctx context.Context, rollNo string) error
	UnFreezeStudent(ctx context.Context, rollNo string) error
	FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error)
	DeleteStudent(ctx context.Context, collegeID int, studentID int) error

	// Find methods with pagination
	FindAllStudentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error)
	CountStudentsByCollege(ctx context.Context, collegeID int) (int, error)
	UpdateStudentPartial(ctx context.Context, collegeID int, studentID int, req *models.UpdateStudentRequest) error
}

type studentRepository struct {
	Pool *pgxpool.Pool
}

func NewStudentRepository(database *DB) StudentRepository {
	return &studentRepository{
		Pool: database.Pool,
	}
}

const studentTable = "students"

func (s *studentRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	// Set timestamps if they are zero-valued
	now := time.Now()
	if student.CreatedAt.IsZero() {
		student.CreatedAt = now
	}
	if student.UpdatedAt.IsZero() {
		student.UpdatedAt = now
	}

	sql := `INSERT INTO students (
    user_id,
    college_id,
    kratos_identity_id,
    enrollment_year,
    roll_no,
    is_active,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at`

	var enrollmentYear int32

	args := []any{int32(student.UserID), int32(student.CollegeID), student.KratosIdentityID, enrollmentYear, student.RollNo, student.IsActive, student.CreatedAt, student.UpdatedAt}
	err := pgxscan.Get(ctx, s.Pool, student, sql, args...)

	if err != nil {
		return fmt.Errorf("CreateStudent: failed to execute query: %w", err)
	}

	return nil
}

func (s *studentRepository) GetStudentByRollNo(ctx context.Context, collegeID int, rollNo string) (*models.Student, error) {
	sql := `SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE roll_no = $1 AND college_id = $2`

	var student models.Student
	var enrollmentYear sqlDriver.NullInt32
	err := pgxscan.Get(ctx, s.Pool, &student, sql, rollNo, int32(collegeID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("GetStudentByRollNo: student with rollNo %s not found in college %d", rollNo, collegeID)
		}
		return nil, fmt.Errorf("GetStudentByRollNo: failed to execute query or scan for college %d, rollNo %s: %w", collegeID, rollNo, err)
	}

	if enrollmentYear.Valid {
		student.EnrollmentYear = int(enrollmentYear.Int32)
	}

	return &student, nil
}

func (s *studentRepository) DeleteStudent(ctx context.Context, collegeID int, studentID int) error {
	sql := `DELETE FROM students
WHERE student_id = $1 AND college_id = $2`

	_, err := s.Pool.Exec(ctx, sql, int32(studentID), int32(collegeID))
	if err != nil {
		return fmt.Errorf("DeleteStudent: failed to execute query: %w", err)
	}

	return nil
}

func (s *studentRepository) FindAllStudentsByCollege(ctx context.Context, collegeID int, limit, offset uint64) ([]*models.Student, error) {
	sql := `SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE college_id = $1
ORDER BY roll_no ASC
LIMIT $2 OFFSET $3`
	var students []*models.Student
	err := pgxscan.Select(ctx, s.Pool, &students, sql, int32(collegeID), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("FindAllStudentsByCollege: failed to query students %w", err)
	}
	return students, nil
}

func (s *studentRepository) CountStudentsByCollege(ctx context.Context, collegeID int) (int, error) {
	sql := `SELECT COUNT(*) as count
FROM students
WHERE college_id = $1`

	temp := struct {
		Count int64 `db:"count"`
	}{}
	err := pgxscan.Get(ctx, s.Pool, &temp, sql, int32(collegeID))
	if err != nil {
		return 0, fmt.Errorf("CountStudentsByCollege: failed to execute query: %w", err)
	}

	return int(temp.Count), nil
}

func (s *studentRepository) GetStudentByID(ctx context.Context, collegeID int, studentID int) (*models.Student, error) {
	sql := `SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE student_id = $1 AND college_id = $2`

	var student models.Student
	var enrollmentYear sqlDriver.NullInt32

	err := pgxscan.Get(ctx, s.Pool, &student, sql, studentID, collegeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetStudentByID: failed to execute query: %w", err)
	}
	if enrollmentYear.Valid {
		student.EnrollmentYear = int(enrollmentYear.Int32)
	}

	return &student, nil
}

func (s *studentRepository) UpdateStudent(ctx context.Context, model *models.Student) error {
	// Update the UpdatedAt timestamp
	model.UpdatedAt = time.Now()

	sql := `UPDATE students
SET user_id = $1,
    college_id = $2,
    kratos_identity_id = $3,
    enrollment_year = $4,
    roll_no = $5,
    is_active = $6,
    updated_at = $7
WHERE student_id = $8`

	_, err := s.Pool.Exec(ctx, sql,
		int32(model.UserID),
		int32(model.CollegeID),
		model.KratosIdentityID,
		int32(model.EnrollmentYear),
		model.RollNo,
		model.IsActive,
		model.UpdatedAt,
		int32(model.StudentID),
	)

	if err != nil {
		return fmt.Errorf("UpdateStudent: failed to execute query: %w", err)
	}

	return nil
}

func (s *studentRepository) FreezeStudent(ctx context.Context, rollNo string) error {
	sql := `UPDATE students
SET is_active = false,
    updated_at = NOW()
WHERE roll_no = $1`

	_, err := s.Pool.Exec(ctx, sql, rollNo)
	if err != nil {
		return fmt.Errorf("FreezeStudent: failed to execute query: %w", err)
	}

	return nil
}

func (s *studentRepository) UnFreezeStudent(ctx context.Context, rollNo string) error {
	sql := `UPDATE students
SET is_active = true,
    updated_at = NOW()
WHERE roll_no = $1`

	_, err := s.Pool.Exec(ctx, sql, rollNo)
	if err != nil {
		return fmt.Errorf("UnFreezeStudent: failed to execute query: %w", err)
	}

	return nil
}

func (s *studentRepository) FindByKratosID(ctx context.Context, kratosID string) (*models.Student, error) {
	sql := `SELECT student_id, user_id, college_id, kratos_identity_id, enrollment_year, roll_no, is_active, created_at, updated_at
FROM students
WHERE kratos_identity_id = $1`

	var student models.Student
	err := pgxscan.Get(ctx, s.Pool, &student, sql, kratosID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindByKratosID: failed to execute query: %w", err)
	}

	return &student, nil
}
func (s *studentRepository) UpdateStudentPartial(ctx context.Context, collegeID int, studentID int, req *models.UpdateStudentRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE students SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.UserID != nil {
		sql += fmt.Sprintf(`, user_id = $%d`, argIndex)
		args = append(args, int32(*req.UserID))
		argIndex++
	}
	if req.CollegeID != nil {
		sql += fmt.Sprintf(`, college_id = $%d`, argIndex)
		args = append(args, int32(*req.CollegeID))
		argIndex++
	}
	if req.EnrollmentYear != nil {
		sql += fmt.Sprintf(`, enrollment_year = $%d`, argIndex)
		args = append(args, int32(*req.EnrollmentYear))
		argIndex++
	}
	if req.RollNo != nil {
		sql += fmt.Sprintf(`, roll_no = $%d`, argIndex)
		args = append(args, *req.RollNo)
		argIndex++
	}
	if req.IsActive != nil {
		sql += fmt.Sprintf(`, is_active = $%d`, argIndex)
		args = append(args, *req.IsActive)
		argIndex++
	}

	sql += fmt.Sprintf(` WHERE student_id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, int32(studentID), int32(collegeID))

	_, err := s.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateStudentPartial: failed to execute query: %w", err)
	}

	return nil
}
