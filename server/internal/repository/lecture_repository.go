package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type LectureRepository interface {
	CreateLecture(ctx context.Context, lecture *models.Lecture) error
	GetLectureByID(ctx context.Context, collegeID int, lectureID int) (*models.Lecture, error)
	UpdateLecture(ctx context.Context, lecture *models.Lecture) error
	UpdateLecturePartial(ctx context.Context, collegeID int, lectureID int, req *models.UpdateLectureRequest) error
	DeleteLecture(ctx context.Context, collegeID int, lectureID int) error

	// Finder methods
	FindLecturesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Lecture, error)
	CountLecturesByCourse(ctx context.Context, collegeID int, courseID int) (int, error)
	// Add more finders as needed, e.g., FindLecturesByDateRange, FindLecturesByInstructor (if lectures are directly linked to instructors)
}

type lectureRepository struct {
	DB *DB
}

func NewLectureRepository(db *DB) LectureRepository {
	return &lectureRepository{DB: db}
}

const lectureTable = "lectures"

func (r *lectureRepository) CreateLecture(ctx context.Context, lecture *models.Lecture) error {
	now := time.Now()
	lecture.CreatedAt = now
	lecture.UpdatedAt = now

	sql := `INSERT INTO lectures (course_id, college_id, title, description, start_time, end_time, meeting_link, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, course_id, college_id, title, description, start_time, end_time, meeting_link, created_at, updated_at`
	var result models.Lecture
	err := pgxscan.Get(ctx, r.DB.Pool, &result, sql, lecture.CourseID, lecture.CollegeID, lecture.Title, lecture.Description, lecture.StartTime, lecture.EndTime, lecture.MeetingLink, lecture.CreatedAt, lecture.UpdatedAt)
	if err != nil {
		// Consider checking for specific DB errors like foreign key violations
		return fmt.Errorf("CreateLecture: failed to execute query or scan: %w", err)
	}
	*lecture = result
	return nil
}

func (r *lectureRepository) GetLectureByID(ctx context.Context, collegeID int, lectureID int) (*models.Lecture, error) {
	sql := `SELECT id, course_id, college_id, title, description, start_time, end_time, meeting_link, created_at, updated_at FROM lectures WHERE id = $1 AND college_id = $2`
	lecture := &models.Lecture{}
	err := pgxscan.Get(ctx, r.DB.Pool, lecture, sql, lectureID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetLectureByID: lecture with ID %d not found for college ID %d", lectureID, collegeID)
		}
		return nil, fmt.Errorf("GetLectureByID: failed to execute query or scan: %w", err)
	}
	return lecture, nil
}

func (r *lectureRepository) UpdateLecture(ctx context.Context, lecture *models.Lecture) error {
	lecture.UpdatedAt = time.Now()

	sql := `UPDATE lectures SET title = $1, description = $2, start_time = $3, end_time = $4, meeting_link = $5, course_id = $6, updated_at = $7 WHERE id = $8 AND college_id = $9`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, lecture.Title, lecture.Description, lecture.StartTime, lecture.EndTime, lecture.MeetingLink, lecture.CourseID, lecture.UpdatedAt, lecture.ID, lecture.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdateLecture: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateLecture: no lecture found with ID %d for college ID %d, or no changes made", lecture.ID, lecture.CollegeID)
	}
	return nil
}

func (r *lectureRepository) DeleteLecture(ctx context.Context, collegeID int, lectureID int) error {
	sql := `DELETE FROM lectures WHERE id = $1 AND college_id = $2`
	cmdTag, err := r.DB.Pool.Exec(ctx, sql, lectureID, collegeID)
	if err != nil {
		// Consider foreign key constraint errors (e.g., if attendance records exist)
		// These should ideally be handled at the service layer (e.g., prevent deletion or cascade)
		return fmt.Errorf("DeleteLecture: failed to execute query: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteLecture: no lecture found with ID %d for college ID %d, or already deleted", lectureID, collegeID)
	}
	return nil
}

func (r *lectureRepository) FindLecturesByCourse(ctx context.Context, collegeID int, courseID int, limit, offset uint64) ([]*models.Lecture, error) {
	sql := `SELECT id, course_id, college_id, title, description, start_time, end_time, meeting_link, created_at, updated_at FROM lectures WHERE college_id = $1 AND course_id = $2 ORDER BY start_time ASC LIMIT $3 OFFSET $4`
	lectures := []*models.Lecture{}
	err := pgxscan.Select(ctx, r.DB.Pool, &lectures, sql, collegeID, courseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FindLecturesByCourse: failed to execute query or scan: %w", err)
	}
	return lectures, nil
}

func (r *lectureRepository) CountLecturesByCourse(ctx context.Context, collegeID int, courseID int) (int, error) {
	sql := `SELECT COUNT(*) as count FROM lectures WHERE college_id = $1 AND course_id = $2`
	var result struct {
		Count int `db:"count"`
	}
	err := pgxscan.Get(ctx, r.DB.Pool, &result, sql, collegeID, courseID)
	if err != nil {
		return 0, fmt.Errorf("CountLecturesByCourse: failed to execute query or scan: %w", err)
	}
	return result.Count, nil
}

func (r *lectureRepository) UpdateLecturePartial(ctx context.Context, collegeID int, lectureID int, req *models.UpdateLectureRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE lectures SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if req.CourseID != nil {
		sql += fmt.Sprintf(`, course_id = $%d`, argIndex)
		args = append(args, int32(*req.CourseID))
		argIndex++
	}
	if req.CollegeID != nil {
		sql += fmt.Sprintf(`, college_id = $%d`, argIndex)
		args = append(args, int32(*req.CollegeID))
		argIndex++
	}
	if req.Title != nil {
		sql += fmt.Sprintf(`, title = $%d`, argIndex)
		args = append(args, *req.Title)
		argIndex++
	}
	if req.Description != nil {
		sql += fmt.Sprintf(`, description = $%d`, argIndex)
		args = append(args, *req.Description)
		argIndex++
	}
	if req.StartTime != nil {
		sql += fmt.Sprintf(`, start_time = $%d`, argIndex)
		args = append(args, *req.StartTime)
		argIndex++
	}
	if req.EndTime != nil {
		sql += fmt.Sprintf(`, end_time = $%d`, argIndex)
		args = append(args, *req.EndTime)
		argIndex++
	}
	if req.MeetingLink != nil {
		sql += fmt.Sprintf(`, meeting_link = $%d`, argIndex)
		args = append(args, *req.MeetingLink)
		argIndex++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, int32(lectureID), int32(collegeID))

	commandTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateLecturePartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateLecturePartial: lecture with ID %d not found in college %d", lectureID, collegeID)
	}

	return nil
}
