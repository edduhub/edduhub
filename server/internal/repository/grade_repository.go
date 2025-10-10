package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type GradeRepository interface {
	CreateGrade(ctx context.Context, grade *models.Grade) error
	GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error)
	UpdateGrade(ctx context.Context, grade *models.Grade) error
	UpdateGradePartial(ctx context.Context, collegeID int, gradeID int, req *models.UpdateGradeRequest) error
	DeleteGrade(ctx context.Context, gradeID int, collegeID int) error
	GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error)
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

	if grade.GradedAt == nil {
		grade.GradedAt = &now
	}

	if grade.TotalMarks > 0 && (grade.Percentage == 0 || math.IsNaN(grade.Percentage)) {
		grade.Percentage = roundToTwoDecimals(float64(grade.ObtainedMarks) / float64(grade.TotalMarks) * 100)
	}

	sql := `INSERT INTO grades (student_id, course_id, college_id, assessment_name, assessment_type, total_marks, obtained_marks, percentage, grade, remarks, graded_by, graded_at, created_at, updated_at)
            VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}

	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql,
		grade.StudentID,
		grade.CourseID,
		grade.CollegeID,
		grade.AssessmentName,
		grade.AssessmentType,
		grade.TotalMarks,
		grade.ObtainedMarks,
		grade.Percentage,
		grade.Grade,
		grade.Remarks,
		grade.GradedBy,
		grade.GradedAt,
		grade.CreatedAt,
		grade.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("CreateGrade: failed to execute query or scan ID: %w", err)
	}

	grade.ID = temp.ID
	return nil
}

func (r *gradeRepository) GetGradeByID(ctx context.Context, gradeID int, collegeID int) (*models.Grade, error) {
	grade := &models.Grade{}

	sql := `SELECT id, student_id, course_id, college_id, assessment_name, assessment_type, total_marks, obtained_marks, percentage, grade, remarks, graded_by, graded_at, created_at, updated_at
            FROM grades WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, grade, sql, gradeID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetGradeByID: grade with ID %d for college ID %d not found", gradeID, collegeID)
		}
		return nil, fmt.Errorf("GetGradeByID: failed to execute query: %w", err)
	}

	return grade, nil
}

func (r *gradeRepository) UpdateGrade(ctx context.Context, grade *models.Grade) error {
	grade.UpdatedAt = time.Now()

	if grade.TotalMarks > 0 && (grade.Percentage == 0 || math.IsNaN(grade.Percentage)) {
		grade.Percentage = roundToTwoDecimals(float64(grade.ObtainedMarks) / float64(grade.TotalMarks) * 100)
	}

	sql := `UPDATE grades SET student_id = $1, course_id = $2, assessment_name = $3, assessment_type = $4, total_marks = $5, obtained_marks = $6, percentage = $7, grade = $8, remarks = $9, graded_by = $10, graded_at = $11, updated_at = $12
            WHERE id = $13 AND college_id = $14`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql,
		grade.StudentID,
		grade.CourseID,
		grade.AssessmentName,
		grade.AssessmentType,
		grade.TotalMarks,
		grade.ObtainedMarks,
		grade.Percentage,
		grade.Grade,
		grade.Remarks,
		grade.GradedBy,
		grade.GradedAt,
		grade.UpdatedAt,
		grade.ID,
		grade.CollegeID,
	)
	if err != nil {
		return fmt.Errorf("UpdateGrade: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateGrade: no grade found with ID %d for college ID %d", grade.ID, grade.CollegeID)
	}

	return nil
}

func (r *gradeRepository) UpdateGradePartial(ctx context.Context, collegeID int, gradeID int, req *models.UpdateGradeRequest) error {
	if gradeID <= 0 || collegeID <= 0 {
		return errors.New("UpdateGradePartial: gradeID and collegeID must be positive")
	}

	sql := "UPDATE grades SET updated_at = NOW()"
	args := []interface{}{}
	idx := 1
	changed := false

	if req.StudentID != nil {
		sql += fmt.Sprintf(", student_id = $%d", idx)
		args = append(args, *req.StudentID)
		idx++
		changed = true
	}

	if req.CourseID != nil {
		sql += fmt.Sprintf(", course_id = $%d", idx)
		args = append(args, *req.CourseID)
		idx++
		changed = true
	}

	if req.CollegeID != nil {
		sql += fmt.Sprintf(", college_id = $%d", idx)
		args = append(args, *req.CollegeID)
		idx++
		changed = true
	}

	if req.AssessmentName != nil {
		sql += fmt.Sprintf(", assessment_name = $%d", idx)
		args = append(args, *req.AssessmentName)
		idx++
		changed = true
	}

	if req.AssessmentType != nil {
		sql += fmt.Sprintf(", assessment_type = $%d", idx)
		args = append(args, *req.AssessmentType)
		idx++
		changed = true
	}

	if req.TotalMarks != nil {
		sql += fmt.Sprintf(", total_marks = $%d", idx)
		args = append(args, *req.TotalMarks)
		idx++
		changed = true
	}

	if req.ObtainedMarks != nil {
		sql += fmt.Sprintf(", obtained_marks = $%d", idx)
		args = append(args, *req.ObtainedMarks)
		idx++
		changed = true
	}

	if req.Percentage != nil {
		sql += fmt.Sprintf(", percentage = $%d", idx)
		args = append(args, *req.Percentage)
		idx++
		changed = true
	}

	if req.Grade != nil {
		sql += fmt.Sprintf(", grade = $%d", idx)
		args = append(args, *req.Grade)
		idx++
		changed = true
	}

	if req.Remarks != nil {
		sql += fmt.Sprintf(", remarks = $%d", idx)
		args = append(args, *req.Remarks)
		idx++
		changed = true
	}

	if req.GradedBy != nil {
		sql += fmt.Sprintf(", graded_by = $%d", idx)
		args = append(args, *req.GradedBy)
		idx++
		changed = true
	}

	if req.GradedAt != nil {
		sql += fmt.Sprintf(", graded_at = $%d", idx)
		args = append(args, *req.GradedAt)
		idx++
		changed = true
	}

	if !changed {
		return errors.New("UpdateGradePartial: no fields provided for update")
	}

	sql += fmt.Sprintf(" WHERE id = $%d AND college_id = $%d", idx, idx+1)
	args = append(args, gradeID, collegeID)

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateGradePartial: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateGradePartial: grade with ID %d not found for college ID %d", gradeID, collegeID)
	}

	return nil
}

func (r *gradeRepository) DeleteGrade(ctx context.Context, gradeID int, collegeID int) error {
	sql := `DELETE FROM grades WHERE id = $1 AND college_id = $2`

	cmdTag, err := r.DB.Pool.Exec(ctx, sql, gradeID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteGrade: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteGrade: no grade found with ID %d for college ID %d", gradeID, collegeID)
	}

	return nil
}

func (r *gradeRepository) GetGrades(ctx context.Context, filter models.GradeFilter) ([]*models.Grade, error) {
	if filter.CollegeID == nil {
		return nil, errors.New("GetGrades: CollegeID filter is required")
	}

	sql := `SELECT id, student_id, course_id, college_id, assessment_name, assessment_type, total_marks, obtained_marks, percentage, grade, remarks, graded_by, graded_at, created_at, updated_at
            FROM grades WHERE college_id = $1`
	args := []interface{}{*filter.CollegeID}
	idx := 2

	if filter.StudentID != nil {
		sql += fmt.Sprintf(" AND student_id = $%d", idx)
		args = append(args, *filter.StudentID)
		idx++
	}

	if filter.CourseID != nil {
		sql += fmt.Sprintf(" AND course_id = $%d", idx)
		args = append(args, *filter.CourseID)
		idx++
	}

	if filter.AssessmentType != nil && *filter.AssessmentType != "" {
		sql += fmt.Sprintf(" AND assessment_type = $%d", idx)
		args = append(args, *filter.AssessmentType)
		idx++
	}

	sql += " ORDER BY graded_at DESC NULLS LAST, updated_at DESC"

	if filter.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT $%d", idx)
		args = append(args, filter.Limit)
		idx++
	}

	if filter.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET $%d", idx)
		args = append(args, filter.Offset)
		idx++
	}

	grades := []*models.Grade{}
	if err := pgxscan.Select(ctx, r.DB.Pool, &grades, sql, args...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Grade{}, nil
		}
		return nil, fmt.Errorf("GetGrades: failed to execute query: %w", err)
	}

	return grades, nil
}

func (r *gradeRepository) GetGradesByCourse(ctx context.Context, collegeID int, courseID int) ([]*models.Grade, error) {
	filter := models.GradeFilter{
		CollegeID: &collegeID,
		CourseID:  &courseID,
	}
	return r.GetGrades(ctx, filter)
}

func (r *gradeRepository) GetGradesByStudent(ctx context.Context, collegeID int, studentID int) ([]*models.Grade, error) {
	filter := models.GradeFilter{
		CollegeID: &collegeID,
		StudentID: &studentID,
	}
	return r.GetGrades(ctx, filter)
}

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
