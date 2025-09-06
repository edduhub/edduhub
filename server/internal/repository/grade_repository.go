package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
)

const gradeTable = "grades"

var gradeQueryFields = []string{
	"id", "student_id", "course_id", "college_id", "marks_obtained", "total_marks",
	"grade_letter", "semester", "academic_year", "exam_type", "graded_at",
	"comments", "created_at", "updated_at",
}

type GradeRepository interface {
	CreateGrade(ctx context.Context, grade *models.Grade) error
	GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error)
	UpdateGrade(ctx context.Context, grade *models.Grade) error
	DeleteGrade(ctx context.Context, gradeID int, collegeID int) error
	GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error)
	// GetStudentProgress and GenerateStudentReport might be higher-level service methods
	// or more complex queries. For now, GetGrades with filters can serve many needs.
	GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error)
	GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error)
}

type gradeRepository struct {
	DB *DB
}

func NewGradeRepository(db *DB) GradeRepository {
	return &gradeRepository{DB: db}
}

func (r *gradeRepository) CreateGrade(ctx context.Context, grade *models.Grade) error {
	now := time.Now()
	grade.CreatedAt = now
	grade.UpdatedAt = now
	if grade.GradedAt.IsZero() { // Default GradedAt to now if not provided
		grade.GradedAt = now
	}

	sql := `INSERT INTO grades (student_id, course_id, college_id, marks_obtained, total_marks, grade_letter, semester, academic_year, exam_type, graded_at, comments, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`
	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, grade.StudentID, grade.CourseID, grade.CollegeID, grade.MarksObtained, grade.TotalMarks, grade.GradeLetter, grade.Semester, grade.AcademicYear, grade.ExamType, grade.GradedAt, grade.Comments, grade.CreatedAt, grade.UpdatedAt)
	if err != nil {
		// Consider specific error handling for duplicate entries or foreign key violations
		return fmt.Errorf("CreateGrade: failed to execute query or scan ID: %w", err)
	}
	grade.ID = temp.ID
	return nil
}

func (r *gradeRepository) GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error) {
	grade := &models.Grade{}
	sql := `SELECT id, student_id, course_id, college_id, marks_obtained, total_marks, grade_letter, semester, academic_year, exam_type, graded_at, comments, created_at, updated_at FROM grades WHERE id = $1 AND college_id = $2`
	err := pgxscan.Get(ctx, r.DB.Pool, grade, sql, gradeID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetGradeByID: grade with ID %d for college ID %d not found: %w", gradeID, collegeID, err)
		}
		return nil, fmt.Errorf("GetGradeByID: failed to execute query or scan: %w", err)
	}
	return grade, nil
}

func (r *gradeRepository) UpdateGrade(ctx context.Context, grade *models.Grade) error {
	grade.UpdatedAt = time.Now()

	sql := `UPDATE grades SET student_id = $1, course_id = $2, marks_obtained = $3, total_marks = $4, grade_letter = $5, semester = $6, academic_year = $7, exam_type = $8, graded_at = $9, comments = $10, updated_at = $11 WHERE id = $12 AND college_id = $13`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, grade.StudentID, grade.CourseID, grade.MarksObtained, grade.TotalMarks, grade.GradeLetter, grade.Semester, grade.AcademicYear, grade.ExamType, grade.GradedAt, grade.Comments, grade.UpdatedAt, grade.ID, grade.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdateGrade: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateGrade: no grade found with ID %d for college ID %d, or no changes made", grade.ID, grade.CollegeID)
	}
	return nil
}

func (r *gradeRepository) DeleteGrade(ctx context.Context, gradeID int, collegeID int) error {
	sql := `DELETE FROM grades WHERE id = $1 AND college_id = $2`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, gradeID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteGrade: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteGrade: no grade found with ID %d for college ID %d", gradeID, collegeID)
	}
	return nil
}

func (r *gradeRepository) GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error) {
	if filter.CollegeID == nil {
		return nil, errors.New("GetGrades: CollegeID filter is required")
	}

	sql := `SELECT id, student_id, course_id, college_id, marks_obtained, total_marks, grade_letter, semester, academic_year, exam_type, graded_at, comments, created_at, updated_at FROM grades WHERE college_id = $1`
	args := []interface{}{*filter.CollegeID}
	placeholderCount := 1

	if filter.StudentID != nil {
		placeholderCount++
		sql += fmt.Sprintf(" AND student_id = $%d", placeholderCount)
		args = append(args, *filter.StudentID)
	}
	if filter.CourseID != nil {
		placeholderCount++
		sql += fmt.Sprintf(" AND course_id = $%d", placeholderCount)
		args = append(args, *filter.CourseID)
	}
	if filter.Semester != nil {
		placeholderCount++
		sql += fmt.Sprintf(" AND semester = $%d", placeholderCount)
		args = append(args, *filter.Semester)
	}
	if filter.AcademicYear != nil {
		placeholderCount++
		sql += fmt.Sprintf(" AND academic_year = $%d", placeholderCount)
		args = append(args, *filter.AcademicYear)
	}
	if filter.ExamType != "" {
		placeholderCount++
		sql += fmt.Sprintf(" AND exam_type = $%d", placeholderCount)
		args = append(args, filter.ExamType)
	}

	sql += " ORDER BY academic_year ASC, semester ASC, graded_at ASC"

	if filter.Limit > 0 {
		placeholderCount++
		sql += fmt.Sprintf(" LIMIT $%d", placeholderCount)
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		placeholderCount++
		sql += fmt.Sprintf(" OFFSET $%d", placeholderCount)
		args = append(args, filter.Offset)
	}

	var grades []*models.Grade
	err := pgxscan.Select(ctx, r.DB.Pool, &grades, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Grade{}, nil // Return empty slice if no rows found
		}
		return nil, fmt.Errorf("GetGrades: failed to execute query or scan: %w", err)
	}
	return grades, nil
}

func (r *gradeRepository) GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error) {
	studentIDStr := string(studentID)
	filter := models.GradeFilter{
		StudentID: &studentIDStr,
		CollegeID: &collegeID,
	}
	return r.GetGrades(ctx, filter)
}

func (r *gradeRepository) GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error) {
	// courseIDStr := string(courseID)
	filter := models.GradeFilter{
		CourseID:  &courseID,
		CollegeID: &collegeID,
	}
	return r.GetGrades(ctx, filter)
}
