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

const timeTableBlockTable = "timetable_blocks"

var timeTableBlockQueryFields = []string{
	"id", "college_id", "department_id", "course_id", "class_id",
	"day_of_week", "start_time", "end_time", "room_number", "faculty_id",
	"created_at", "updated_at",
}

type TimeTableRepository interface {
	CreateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error
	GetTimeTableBlockByID(ctx context.Context, blockID int, collegeID int) (*models.TimeTableBlock, error)
	UpdateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error
	DeleteTimeTableBlock(ctx context.Context, blockID int, collegeID int) error
	GetTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) ([]*models.TimeTableBlock, error)
	CountTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) (int, error)
}

type timetableRepository struct {
	DB *DB
}

func NewTimeTableRepository(db *DB) TimeTableRepository {
	return &timetableRepository{DB: db}
}

func (r *timetableRepository) CreateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error {
	now := time.Now()
	block.CreatedAt = now
	block.UpdatedAt = now

	sql := `INSERT INTO timetable_blocks (college_id, department_id, course_id, class_id, day_of_week, start_time, end_time, room_number, faculty_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`
	err := r.DB.Pool.QueryRow(ctx, sql, block.CollegeID, block.DepartmentID, block.CourseID, block.ClassID, block.DayOfWeek, block.StartTime, block.EndTime, block.RoomNumber, block.FacultyID, block.CreatedAt, block.UpdatedAt).Scan(&block.ID)
	if err != nil {
		return fmt.Errorf("CreateTimeTableBlock: failed to execute query or scan ID: %w", err)
	}
	return nil
}

func (r *timetableRepository) GetTimeTableBlockByID(ctx context.Context, blockID int, collegeID int) (*models.TimeTableBlock, error) {
	sql := `SELECT id, college_id, department_id, course_id, class_id, day_of_week, start_time, end_time, room_number, faculty_id, created_at, updated_at FROM timetable_blocks WHERE id = $1 AND college_id = $2`
	block := &models.TimeTableBlock{}
	err := pgxscan.Get(ctx, r.DB.Pool, block, sql, blockID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetTimeTableBlockByID: block with ID %d for college ID %d not found: %w", blockID, collegeID, err)
		}
		return nil, fmt.Errorf("GetTimeTableBlockByID: failed to execute query or scan: %w", err)
	}
	return block, nil
}

func (r *timetableRepository) UpdateTimeTableBlock(ctx context.Context, block *models.TimeTableBlock) error {
	block.UpdatedAt = time.Now()

	sql := `UPDATE timetable_blocks SET department_id = $1, course_id = $2, class_id = $3, day_of_week = $4, start_time = $5, end_time = $6, room_number = $7, faculty_id = $8, updated_at = $9 WHERE id = $10 AND college_id = $11`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, block.DepartmentID, block.CourseID, block.ClassID, block.DayOfWeek, block.StartTime, block.EndTime, block.RoomNumber, block.FacultyID, block.UpdatedAt, block.ID, block.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdateTimeTableBlock: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateTimeTableBlock: no block found with ID %d for college ID %d, or no changes made", block.ID, block.CollegeID)
	}
	return nil
}

func (r *timetableRepository) DeleteTimeTableBlock(ctx context.Context, blockID int, collegeID int) error {
	sql := `DELETE FROM timetable_blocks WHERE id = $1 AND college_id = $2`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, blockID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteTimeTableBlock: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteTimeTableBlock: no block found with ID %d for college ID %d", blockID, collegeID)
	}
	return nil
}


func (r *timetableRepository) GetTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) ([]*models.TimeTableBlock, error) {
	if filter.CollegeID == 0 { // Or handle as pointer and check for nil
		return nil, errors.New("GetTimeTableBlocks: CollegeID filter is required")
	}

	sql := "SELECT id, college_id, department_id, course_id, class_id, day_of_week, start_time, end_time, room_number, faculty_id, created_at, updated_at FROM timetable_blocks WHERE college_id = $1"
	args := []interface{}{filter.CollegeID}
	paramCount := 1

	if filter.DepartmentID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND department_id = $%d", paramCount)
		args = append(args, *filter.DepartmentID)
	}
	if filter.CourseID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND course_id = $%d", paramCount)
		args = append(args, *filter.CourseID)
	}
	if filter.ClassID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND class_id = $%d", paramCount)
		args = append(args, *filter.ClassID)
	}
	if filter.DayOfWeek != nil {
		paramCount++
		sql += fmt.Sprintf(" AND day_of_week = $%d", paramCount)
		args = append(args, *filter.DayOfWeek)
	}
	if filter.InstructorID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND faculty_id = $%d", paramCount)
		args = append(args, *filter.InstructorID)
	}
	if filter.StartTime != nil {
		paramCount++
		sql += fmt.Sprintf(" AND start_time = $%d", paramCount)
		args = append(args, *filter.StartTime)
	}
	if filter.EndTime != nil {
		paramCount++
		sql += fmt.Sprintf(" AND end_time = $%d", paramCount)
		args = append(args, *filter.EndTime)
	}

	sql += " ORDER BY day_of_week ASC, start_time ASC"

	if filter.Limit > 0 {
		paramCount++
		sql += fmt.Sprintf(" LIMIT $%d", paramCount)
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		paramCount++
		sql += fmt.Sprintf(" OFFSET $%d", paramCount)
		args = append(args, filter.Offset)
	}

	var blocks []*models.TimeTableBlock
	err := pgxscan.Select(ctx, r.DB.Pool, &blocks, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.TimeTableBlock{}, nil
		}
		return nil, fmt.Errorf("GetTimeTableBlocks: failed to execute query or scan: %w", err)
	}
	return blocks, nil
}

func (r *timetableRepository) CountTimeTableBlocks(ctx context.Context, filter models.TimeTableBlockFilter) (int, error) {
	if filter.CollegeID == 0 {
		return 0, errors.New("CountTimeTableBlocks: CollegeID filter is required")
	}

	sql := "SELECT COUNT(*) FROM timetable_blocks WHERE college_id = $1"
	args := []interface{}{filter.CollegeID}
	paramCount := 1

	if filter.DepartmentID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND department_id = $%d", paramCount)
		args = append(args, *filter.DepartmentID)
	}
	if filter.CourseID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND course_id = $%d", paramCount)
		args = append(args, *filter.CourseID)
	}
	if filter.ClassID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND class_id = $%d", paramCount)
		args = append(args, *filter.ClassID)
	}
	if filter.DayOfWeek != nil {
		paramCount++
		sql += fmt.Sprintf(" AND day_of_week = $%d", paramCount)
		args = append(args, *filter.DayOfWeek)
	}
	if filter.InstructorID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND faculty_id = $%d", paramCount)
		args = append(args, *filter.InstructorID)
	}
	if filter.StartTime != nil {
		paramCount++
		sql += fmt.Sprintf(" AND start_time = $%d", paramCount)
		args = append(args, *filter.StartTime)
	}
	if filter.EndTime != nil {
		paramCount++
		sql += fmt.Sprintf(" AND end_time = $%d", paramCount)
		args = append(args, *filter.EndTime)
	}

	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountTimeTableBlocks: failed to execute query or scan: %w", err)
	}
	return count, nil
}
