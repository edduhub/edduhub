package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"eduhub/server/internal/models"

	"github.com/jackc/pgx/v5"
)

type FacultyToolsRepository interface {
	ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error)

	CreateRubric(ctx context.Context, rubric *models.GradingRubric) error
	UpdateRubric(ctx context.Context, rubric *models.GradingRubric) error
	DeleteRubric(ctx context.Context, collegeID, rubricID int) error
	GetRubricByID(ctx context.Context, collegeID, rubricID int) (*models.GradingRubric, error)
	ListRubrics(ctx context.Context, collegeID int, facultyID *int) ([]*models.GradingRubric, error)

	CreateOfficeHour(ctx context.Context, slot *models.OfficeHourSlot) error
	UpdateOfficeHour(ctx context.Context, slot *models.OfficeHourSlot) error
	DeleteOfficeHour(ctx context.Context, collegeID, officeHourID int) error
	GetOfficeHourByID(ctx context.Context, collegeID, officeHourID int) (*models.OfficeHourSlot, error)
	ListOfficeHours(ctx context.Context, collegeID int, facultyID *int, activeOnly bool) ([]*models.OfficeHourSlot, error)

	CreateBooking(ctx context.Context, booking *models.OfficeHourBooking) error
	GetBookingByID(ctx context.Context, collegeID, bookingID int) (*models.OfficeHourBooking, error)
	ListBookings(ctx context.Context, collegeID int, officeHourID, studentID, facultyID *int) ([]*models.OfficeHourBooking, error)
	UpdateBookingStatus(ctx context.Context, collegeID, bookingID int, status string, notes *string) (*models.OfficeHourBooking, error)
}

type facultyToolsRepository struct {
	DB *DB
}

func NewFacultyToolsRepository(db *DB) FacultyToolsRepository {
	return &facultyToolsRepository{DB: db}
}

func (r *facultyToolsRepository) ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error) {
	var userID int
	err := r.DB.Pool.QueryRow(ctx, `SELECT id FROM users WHERE kratos_identity_id = $1 AND is_active = TRUE`, kratosID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return userID, nil
}

func (r *facultyToolsRepository) CreateRubric(ctx context.Context, rubric *models.GradingRubric) error {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	insertRubric := `
		INSERT INTO grading_rubrics (
			faculty_id, college_id, name, description, course_id, is_template, is_active, max_score
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(
		ctx,
		insertRubric,
		rubric.FacultyID,
		rubric.CollegeID,
		rubric.Name,
		rubric.Description,
		rubric.CourseID,
		rubric.IsTemplate,
		rubric.IsActive,
		rubric.MaxScore,
	).Scan(&rubric.ID, &rubric.CreatedAt, &rubric.UpdatedAt)
	if err != nil {
		return err
	}

	for idx := range rubric.Criteria {
		criterion := &rubric.Criteria[idx]
		if criterion.SortOrder == 0 {
			criterion.SortOrder = idx + 1
		}
		insertCriterion := `
			INSERT INTO rubric_criteria (
				rubric_id, name, description, weight, max_score, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at`
		err = tx.QueryRow(
			ctx,
			insertCriterion,
			rubric.ID,
			criterion.Name,
			criterion.Description,
			criterion.Weight,
			criterion.MaxScore,
			criterion.SortOrder,
		).Scan(&criterion.ID, &criterion.CreatedAt, &criterion.UpdatedAt)
		if err != nil {
			return err
		}
		criterion.RubricID = rubric.ID
	}

	return tx.Commit(ctx)
}

func (r *facultyToolsRepository) UpdateRubric(ctx context.Context, rubric *models.GradingRubric) error {
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	updateRubric := `
		UPDATE grading_rubrics
		SET name = $1,
			description = $2,
			course_id = $3,
			is_template = $4,
			is_active = $5,
			max_score = $6,
			updated_at = NOW()
		WHERE id = $7 AND college_id = $8
		RETURNING updated_at`
	if err := tx.QueryRow(
		ctx,
		updateRubric,
		rubric.Name,
		rubric.Description,
		rubric.CourseID,
		rubric.IsTemplate,
		rubric.IsActive,
		rubric.MaxScore,
		rubric.ID,
		rubric.CollegeID,
	).Scan(&rubric.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("rubric not found")
		}
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM rubric_criteria WHERE rubric_id = $1`, rubric.ID); err != nil {
		return err
	}

	for idx := range rubric.Criteria {
		criterion := &rubric.Criteria[idx]
		if criterion.SortOrder == 0 {
			criterion.SortOrder = idx + 1
		}
		insertCriterion := `
			INSERT INTO rubric_criteria (
				rubric_id, name, description, weight, max_score, sort_order
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, created_at, updated_at`
		if err := tx.QueryRow(
			ctx,
			insertCriterion,
			rubric.ID,
			criterion.Name,
			criterion.Description,
			criterion.Weight,
			criterion.MaxScore,
			criterion.SortOrder,
		).Scan(&criterion.ID, &criterion.CreatedAt, &criterion.UpdatedAt); err != nil {
			return err
		}
		criterion.RubricID = rubric.ID
	}

	return tx.Commit(ctx)
}

func (r *facultyToolsRepository) DeleteRubric(ctx context.Context, collegeID, rubricID int) error {
	result, err := r.DB.Pool.Exec(ctx, `DELETE FROM grading_rubrics WHERE id = $1 AND college_id = $2`, rubricID, collegeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("rubric not found")
	}
	return nil
}

func (r *facultyToolsRepository) GetRubricByID(ctx context.Context, collegeID, rubricID int) (*models.GradingRubric, error) {
	query := `
		SELECT id, faculty_id, college_id, name, description, course_id,
			is_template, is_active, max_score, created_at, updated_at
		FROM grading_rubrics
		WHERE id = $1 AND college_id = $2`

	rubric := &models.GradingRubric{}
	err := r.DB.Pool.QueryRow(ctx, query, rubricID, collegeID).Scan(
		&rubric.ID,
		&rubric.FacultyID,
		&rubric.CollegeID,
		&rubric.Name,
		&rubric.Description,
		&rubric.CourseID,
		&rubric.IsTemplate,
		&rubric.IsActive,
		&rubric.MaxScore,
		&rubric.CreatedAt,
		&rubric.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	criteria, err := r.listRubricCriteria(ctx, rubric.ID)
	if err != nil {
		return nil, err
	}
	rubric.Criteria = criteria
	return rubric, nil
}

func (r *facultyToolsRepository) ListRubrics(ctx context.Context, collegeID int, facultyID *int) ([]*models.GradingRubric, error) {
	query := `
		SELECT id, faculty_id, college_id, name, description, course_id,
			is_template, is_active, max_score, created_at, updated_at
		FROM grading_rubrics
		WHERE college_id = $1`
	args := []any{collegeID}
	if facultyID != nil {
		query += ` AND faculty_id = $2`
		args = append(args, *facultyID)
	}
	query += ` ORDER BY updated_at DESC, id DESC`

	rows, err := r.DB.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rubrics := make([]*models.GradingRubric, 0)
	for rows.Next() {
		item := &models.GradingRubric{}
		if err := rows.Scan(
			&item.ID,
			&item.FacultyID,
			&item.CollegeID,
			&item.Name,
			&item.Description,
			&item.CourseID,
			&item.IsTemplate,
			&item.IsActive,
			&item.MaxScore,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		criteria, err := r.listRubricCriteria(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Criteria = criteria
		rubrics = append(rubrics, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rubrics, nil
}

func (r *facultyToolsRepository) listRubricCriteria(ctx context.Context, rubricID int) ([]models.RubricCriterion, error) {
	query := `
		SELECT id, rubric_id, name, description, weight, max_score, sort_order, created_at, updated_at
		FROM rubric_criteria
		WHERE rubric_id = $1
		ORDER BY sort_order ASC, id ASC`

	rows, err := r.DB.Pool.Query(ctx, query, rubricID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	criteria := make([]models.RubricCriterion, 0)
	for rows.Next() {
		var c models.RubricCriterion
		if err := rows.Scan(
			&c.ID,
			&c.RubricID,
			&c.Name,
			&c.Description,
			&c.Weight,
			&c.MaxScore,
			&c.SortOrder,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		criteria = append(criteria, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return criteria, nil
}

func (r *facultyToolsRepository) CreateOfficeHour(ctx context.Context, slot *models.OfficeHourSlot) error {
	query := `
		INSERT INTO faculty_office_hours (
			faculty_id, college_id, day_of_week, start_time, end_time,
			location, is_virtual, virtual_link, max_students, is_active
		) VALUES ($1, $2, $3, $4::time, $5::time, $6, $7, $8, $9, $10)
		RETURNING id, to_char(start_time, 'HH24:MI') AS start_time,
			to_char(end_time, 'HH24:MI') AS end_time, created_at, updated_at`

	if err := r.DB.Pool.QueryRow(
		ctx,
		query,
		slot.FacultyID,
		slot.CollegeID,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.Location,
		slot.IsVirtual,
		slot.VirtualLink,
		slot.MaxStudents,
		slot.IsActive,
	).Scan(&slot.ID, &slot.StartTime, &slot.EndTime, &slot.CreatedAt, &slot.UpdatedAt); err != nil {
		return err
	}
	return nil
}

func (r *facultyToolsRepository) UpdateOfficeHour(ctx context.Context, slot *models.OfficeHourSlot) error {
	query := `
		UPDATE faculty_office_hours
		SET day_of_week = $1,
			start_time = $2::time,
			end_time = $3::time,
			location = $4,
			is_virtual = $5,
			virtual_link = $6,
			max_students = $7,
			is_active = $8,
			updated_at = NOW()
		WHERE id = $9 AND college_id = $10
		RETURNING to_char(start_time, 'HH24:MI') AS start_time,
			to_char(end_time, 'HH24:MI') AS end_time, updated_at`
	if err := r.DB.Pool.QueryRow(
		ctx,
		query,
		slot.DayOfWeek,
		slot.StartTime,
		slot.EndTime,
		slot.Location,
		slot.IsVirtual,
		slot.VirtualLink,
		slot.MaxStudents,
		slot.IsActive,
		slot.ID,
		slot.CollegeID,
	).Scan(&slot.StartTime, &slot.EndTime, &slot.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("office hour not found")
		}
		return err
	}
	return nil
}

func (r *facultyToolsRepository) DeleteOfficeHour(ctx context.Context, collegeID, officeHourID int) error {
	result, err := r.DB.Pool.Exec(ctx, `DELETE FROM faculty_office_hours WHERE id = $1 AND college_id = $2`, officeHourID, collegeID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("office hour not found")
	}
	return nil
}

func (r *facultyToolsRepository) GetOfficeHourByID(ctx context.Context, collegeID, officeHourID int) (*models.OfficeHourSlot, error) {
	query := `
		SELECT oh.id, oh.faculty_id, oh.college_id, oh.day_of_week,
			to_char(oh.start_time, 'HH24:MI') AS start_time,
			to_char(oh.end_time, 'HH24:MI') AS end_time,
			oh.location, oh.is_virtual, oh.virtual_link, oh.max_students,
			oh.is_active, oh.created_at, oh.updated_at,
			u.name as faculty_name
		FROM faculty_office_hours oh
		LEFT JOIN users u ON u.id = oh.faculty_id
		WHERE oh.id = $1 AND oh.college_id = $2`

	slot := &models.OfficeHourSlot{}
	err := r.DB.Pool.QueryRow(ctx, query, officeHourID, collegeID).Scan(
		&slot.ID,
		&slot.FacultyID,
		&slot.CollegeID,
		&slot.DayOfWeek,
		&slot.StartTime,
		&slot.EndTime,
		&slot.Location,
		&slot.IsVirtual,
		&slot.VirtualLink,
		&slot.MaxStudents,
		&slot.IsActive,
		&slot.CreatedAt,
		&slot.UpdatedAt,
		&slot.FacultyName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return slot, nil
}

func (r *facultyToolsRepository) ListOfficeHours(ctx context.Context, collegeID int, facultyID *int, activeOnly bool) ([]*models.OfficeHourSlot, error) {
	query := `
		SELECT oh.id, oh.faculty_id, oh.college_id, oh.day_of_week,
			to_char(oh.start_time, 'HH24:MI') AS start_time,
			to_char(oh.end_time, 'HH24:MI') AS end_time,
			oh.location, oh.is_virtual, oh.virtual_link, oh.max_students,
			oh.is_active, oh.created_at, oh.updated_at,
			u.name as faculty_name
		FROM faculty_office_hours oh
		LEFT JOIN users u ON u.id = oh.faculty_id
		WHERE oh.college_id = $1`
	args := []any{collegeID}
	argPos := 2
	if facultyID != nil {
		query += fmt.Sprintf(` AND oh.faculty_id = $%d`, argPos)
		args = append(args, *facultyID)
		argPos++
	}
	if activeOnly {
		query += fmt.Sprintf(` AND oh.is_active = $%d`, argPos)
		args = append(args, true)
		argPos++
	}
	query += ` ORDER BY oh.day_of_week ASC, oh.start_time ASC`

	rows, err := r.DB.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]*models.OfficeHourSlot, 0)
	for rows.Next() {
		slot := &models.OfficeHourSlot{}
		if err := rows.Scan(
			&slot.ID,
			&slot.FacultyID,
			&slot.CollegeID,
			&slot.DayOfWeek,
			&slot.StartTime,
			&slot.EndTime,
			&slot.Location,
			&slot.IsVirtual,
			&slot.VirtualLink,
			&slot.MaxStudents,
			&slot.IsActive,
			&slot.CreatedAt,
			&slot.UpdatedAt,
			&slot.FacultyName,
		); err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return slots, nil
}

func (r *facultyToolsRepository) CreateBooking(ctx context.Context, booking *models.OfficeHourBooking) error {
	query := `
		INSERT INTO office_hour_bookings (
			office_hour_id, student_id, booking_date, start_time, end_time, purpose, status, notes
		) VALUES ($1, $2, $3, $4::time, $5::time, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	if err := r.DB.Pool.QueryRow(
		ctx,
		query,
		booking.OfficeHourID,
		booking.StudentID,
		booking.BookingDate,
		booking.StartTime,
		booking.EndTime,
		booking.Purpose,
		booking.Status,
		booking.Notes,
	).Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (r *facultyToolsRepository) GetBookingByID(ctx context.Context, collegeID, bookingID int) (*models.OfficeHourBooking, error) {
	query := `
		SELECT b.id, b.office_hour_id, b.student_id, oh.college_id, b.booking_date,
			to_char(b.start_time, 'HH24:MI') AS start_time,
			to_char(b.end_time, 'HH24:MI') AS end_time,
			b.purpose, b.status, b.notes, b.created_at, b.updated_at
		FROM office_hour_bookings b
		JOIN faculty_office_hours oh ON oh.id = b.office_hour_id
		WHERE b.id = $1 AND oh.college_id = $2`

	item := &models.OfficeHourBooking{}
	err := r.DB.Pool.QueryRow(ctx, query, bookingID, collegeID).Scan(
		&item.ID,
		&item.OfficeHourID,
		&item.StudentID,
		&item.CollegeID,
		&item.BookingDate,
		&item.StartTime,
		&item.EndTime,
		&item.Purpose,
		&item.Status,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func (r *facultyToolsRepository) ListBookings(ctx context.Context, collegeID int, officeHourID, studentID, facultyID *int) ([]*models.OfficeHourBooking, error) {
	query := `
		SELECT b.id, b.office_hour_id, b.student_id, oh.college_id, b.booking_date,
			to_char(b.start_time, 'HH24:MI') AS start_time,
			to_char(b.end_time, 'HH24:MI') AS end_time,
			b.purpose, b.status, b.notes, b.created_at, b.updated_at
		FROM office_hour_bookings b
		JOIN faculty_office_hours oh ON oh.id = b.office_hour_id
		WHERE oh.college_id = $1`
	args := []any{collegeID}
	argPos := 2
	if officeHourID != nil {
		query += fmt.Sprintf(` AND b.office_hour_id = $%d`, argPos)
		args = append(args, *officeHourID)
		argPos++
	}
	if studentID != nil {
		query += fmt.Sprintf(` AND b.student_id = $%d`, argPos)
		args = append(args, *studentID)
		argPos++
	}
	if facultyID != nil {
		query += fmt.Sprintf(` AND oh.faculty_id = $%d`, argPos)
		args = append(args, *facultyID)
		argPos++
	}
	query += ` ORDER BY b.booking_date DESC, b.start_time ASC, b.id DESC`

	rows, err := r.DB.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*models.OfficeHourBooking, 0)
	for rows.Next() {
		item := &models.OfficeHourBooking{}
		if err := rows.Scan(
			&item.ID,
			&item.OfficeHourID,
			&item.StudentID,
			&item.CollegeID,
			&item.BookingDate,
			&item.StartTime,
			&item.EndTime,
			&item.Purpose,
			&item.Status,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *facultyToolsRepository) UpdateBookingStatus(ctx context.Context, collegeID, bookingID int, status string, notes *string) (*models.OfficeHourBooking, error) {
	query := `
		UPDATE office_hour_bookings b
		SET status = $1,
			notes = $2,
			updated_at = NOW()
		FROM faculty_office_hours oh
		WHERE b.office_hour_id = oh.id
		  AND b.id = $3
		  AND oh.college_id = $4
		RETURNING b.id, b.office_hour_id, b.student_id, oh.college_id, b.booking_date,
			to_char(b.start_time, 'HH24:MI') AS start_time,
			to_char(b.end_time, 'HH24:MI') AS end_time,
			b.purpose, b.status, b.notes, b.created_at, b.updated_at`

	item := &models.OfficeHourBooking{}
	err := r.DB.Pool.QueryRow(ctx, query, strings.ToLower(status), notes, bookingID, collegeID).Scan(
		&item.ID,
		&item.OfficeHourID,
		&item.StudentID,
		&item.CollegeID,
		&item.BookingDate,
		&item.StartTime,
		&item.EndTime,
		&item.Purpose,
		&item.Status,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

func parseTimeHHMM(in string) (string, error) {
	if in == "" {
		return "", fmt.Errorf("time is required")
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(layout, in); err == nil {
			return t.Format("15:04"), nil
		}
	}
	return "", fmt.Errorf("invalid time format")
}
