package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5" // For pgx.ErrNoRows
)

const calendarBlockTable = "calendar_events"

var calendarBlockQueryFields = []string{
	"id", "college_id", "title", "description", "event_type", "start_time", "end_time",
	"created_at", "updated_at",
}

type CalendarRepository interface {
	CreateCalendarBlock(ctx context.Context, block *models.CalendarBlock) error
	GetCalendarBlockByID(ctx context.Context, blockID int, collegeID int) (*models.CalendarBlock, error)
	UpdateCalendarBlock(ctx context.Context, block *models.CalendarBlock) error
	UpdateCalendarBlockPartial(ctx context.Context, collegeID int, calendarID int, req *models.UpdateCalendarRequest) error
	DeleteCalendarBlock(ctx context.Context, blockID int, collegeID int) error
	GetCalendarBlocks(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error)
	CountCalendarBlocks(ctx context.Context, filter models.CalendarBlockFilter) (int, error)
}

type calendarRepository struct {
	DB *DB
}

func NewCalendarRepository(db *DB) CalendarRepository {
	return &calendarRepository{DB: db}
}

func (r *calendarRepository) CreateCalendarBlock(ctx context.Context, block *models.CalendarBlock) error {
	now := time.Now()
	block.CreatedAt = now
	block.UpdatedAt = now

	sql := `INSERT INTO calendar_events (college_id, title, description, event_type, start_time, end_time, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	temp := struct {
		ID int `db:"id"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, block.CollegeID, block.Title, block.Description, block.EventType, block.StartTime, block.EndTime, block.CreatedAt, block.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateCalendarBlock: failed to execute query or scan ID: %w", err)
	}
	block.ID = temp.ID
	return nil
}

func (r *calendarRepository) GetCalendarBlockByID(ctx context.Context, blockID int, collegeID int) (*models.CalendarBlock, error) {
	block := &models.CalendarBlock{}
	sql := `SELECT id, college_id, title, description, event_type, start_time, end_time, created_at, updated_at FROM calendar_events WHERE id = $1 AND college_id = $2`

	err := pgxscan.Get(ctx, r.DB.Pool, block, sql, blockID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCalendarBlockByID: block with ID %d for college ID %d not found: %w", blockID, collegeID, err)
		}
		return nil, fmt.Errorf("GetCalendarBlockByID: failed to execute query or scan: %w", err)
	}
	return block, nil
}

func (r *calendarRepository) UpdateCalendarBlock(ctx context.Context, block *models.CalendarBlock) error {
	block.UpdatedAt = time.Now()

	sql := `UPDATE calendar_events SET title = $1, description = $2, event_type = $3, start_time = $4, end_time = $5, updated_at = $6 WHERE id = $7 AND college_id = $8`

	commandTag, err := r.DB.Pool.Exec(ctx, sql, block.Title, block.Description, block.EventType, block.StartTime, block.EndTime, block.UpdatedAt, block.ID, block.CollegeID)
	if err != nil {
		return fmt.Errorf("UpdateCalendarBlock: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCalendarBlock: no block found with ID %d for college ID %d, or no changes made", block.ID, block.CollegeID)
	}
	return nil
}

func (r *calendarRepository) UpdateCalendarBlockPartial(ctx context.Context, collegeID int, calendarID int, req *models.UpdateCalendarRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE calendar_events SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

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
	if req.EventType != nil {
		sql += fmt.Sprintf(`, event_type = $%d`, argIndex)
		args = append(args, *req.EventType)
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

	if len(args) == 0 {
		return errors.New("UpdateCalendarBlockPartial: no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d AND college_id = $%d`, argIndex, argIndex+1)
	args = append(args, calendarID, collegeID)

	commandTag, err := r.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateCalendarBlockPartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCalendarBlockPartial: calendar block with ID %d not found in college %d", calendarID, collegeID)
	}

	return nil
}

func (r *calendarRepository) DeleteCalendarBlock(ctx context.Context, blockID int, collegeID int) error {
	sql := `DELETE FROM calendar_events WHERE id = $1 AND college_id = $2`

	commandTag, err := r.DB.Pool.Exec(ctx, sql, blockID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteCalendarBlock: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteCalendarBlock: no block found with ID %d for college ID %d", blockID, collegeID)
	}
	return nil
}

func (r *calendarRepository) applyCalendarBlockFilter(filter models.CalendarBlockFilter) (string, []any) {
	clauses := []string{}
	args := []any{}

	if filter.CollegeID != nil {
		clauses = append(clauses, fmt.Sprintf("college_id = $%d", len(args)+1))
		args = append(args, *filter.CollegeID)
	}

	if filter.EventType != nil {
		clauses = append(clauses, fmt.Sprintf("event_type = $%d", len(args)+1))
		args = append(args, *filter.EventType)
	}
	if filter.StartDate != nil {
		clauses = append(clauses, fmt.Sprintf("start_time >= $%d", len(args)+1))
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		clauses = append(clauses, fmt.Sprintf("start_time <= $%d", len(args)+1))
		args = append(args, *filter.EndDate)
	}

	var whereClause string
	if len(clauses) > 0 {
		whereClause = "WHERE " + strings.Join(clauses, " AND ")
	}
	return whereClause, args
}

func (r *calendarRepository) GetCalendarBlocks(ctx context.Context, filter models.CalendarBlockFilter) ([]*models.CalendarBlock, error) {
	if filter.CollegeID == nil {
		return nil, errors.New("GetCalendarBlocks: CollegeID filter is required")
	}

	whereClause, args := r.applyCalendarBlockFilter(filter)
	baseSQL := "SELECT id, college_id, title, description, event_type, start_time, end_time, created_at, updated_at FROM calendar_events " + whereClause + " ORDER BY start_time ASC, created_at ASC"

	sql := baseSQL
	if filter.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	var blocks []*models.CalendarBlock
	err := pgxscan.Select(ctx, r.DB.Pool, &blocks, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.CalendarBlock{}, nil
		}
		return nil, fmt.Errorf("GetCalendarBlocks: failed to execute query or scan: %w", err)
	}
	return blocks, nil
}

func (r *calendarRepository) CountCalendarBlocks(ctx context.Context, filter models.CalendarBlockFilter) (int, error) {
	if filter.CollegeID == nil {
		return 0, errors.New("CountCalendarBlocks: CollegeID filter is required")
	}

	whereClause, args := r.applyCalendarBlockFilter(filter)
	sql := "SELECT COUNT(*) FROM calendar_events " + whereClause

	temp := struct {
		Count int `db:"count"`
	}{}
	err := pgxscan.Get(ctx, r.DB.Pool, &temp, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("CountCalendarBlocks: failed to execute query or scan: %w", err)
	}
	return temp.Count, nil
}
