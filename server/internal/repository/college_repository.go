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

type CollegeRepository interface {
	CreateCollege(ctx context.Context, college *models.College) error
	GetCollegeByID(ctx context.Context, id int) (*models.College, error)
	GetCollegeByName(ctx context.Context, name string) (*models.College, error)
	UpdateCollege(ctx context.Context, college *models.College) error
	DeleteCollege(ctx context.Context, id int) error
	ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error)
}

type collegeRepository struct {
	DB *DB
}

const collegeTable = "college"

func NewCollegeRepository(DB *DB) CollegeRepository {
	return &collegeRepository{
		DB: DB,
	}
}

// type College struct {
// 	ID        int       `db:"id" json:"id"`
// 	Name      string    `db:"name" json:"name"`
// 	Address   string    `db:"address" json:"address"`
// 	City      string    `db:"city" json:"city"`
// 	State     string    `db:"state" json:"state"`
// 	Country   string    `db:"country" json:"country"`
// 	CreatedAt time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

// 	// Relations - not stored in DB
// 	Students []*Student `db:"-" json:"students,omitempty"`
// }

func (c *collegeRepository) CreateCollege(ctx context.Context, college *models.College) error {
	now := time.Now()
	college.CreatedAt = now
	college.UpdatedAt = now

	sql := `INSERT INTO college (name, address, city, state, country, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := c.DB.Pool.QueryRow(ctx, sql, college.Name, college.Address, college.City, college.State, college.Country, college.CreatedAt, college.UpdatedAt).Scan(&college.ID)
	if err != nil {
		return fmt.Errorf("CreateCollege: failed to execute query: %w", err)
	}
	return nil
}

func (c *collegeRepository) GetCollegeByID(ctx context.Context, id int) (*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE id = $1`
	college := &models.College{}
	findErr := pgxscan.Get(ctx, c.DB.Pool, college, sql, id)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCollegeByID: college with id %d not found", id)
		}
		return nil, fmt.Errorf("GetCollegeByID: failed to execute query: %w", findErr)
	}
	return college, nil
}

func (c *collegeRepository) GetCollegeByName(ctx context.Context, name string) (*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE name = $1`
	college := &models.College{}
	findErr := pgxscan.Get(ctx, c.DB.Pool, college, sql, name)
	if findErr != nil {
		if errors.Is(findErr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetCollegeByName: college with name '%s' not found", name)
		}
		return nil, fmt.Errorf("GetCollegeByName: failed to execute query: %w", findErr)
	}
	return college, nil
}

// type College struct {
// 	ID        int       `db:"id" json:"id"`
// 	Name      string    `db:"name" json:"name"`
// 	Address   string    `db:"address" json:"address"`
// 	City      string    `db:"city" json:"city"`
// 	State     string    `db:"state" json:"state"`
// 	Country   string    `db:"country" json:"country"`
// 	CreatedAt time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

// 	// Relations - not stored in DB
// 	Students []*Student `db:"-" json:"students,omitempty"`
// }

func (c *collegeRepository) UpdateCollege(ctx context.Context, college *models.College) error {
	college.UpdatedAt = time.Now()
	sql := `UPDATE college SET name = $1, address = $2, city = $3, state = $4, country = $5, updated_at = $6 WHERE id = $7`
	commandTag, err := c.DB.Pool.Exec(ctx, sql, college.Name, college.Address, college.City, college.State, college.Country, college.UpdatedAt, college.ID)
	if err != nil {
		return fmt.Errorf("UpdateCollege: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateCollege: no college found with ID %d", college.ID)
	}

	return nil
}

func (c *collegeRepository) DeleteCollege(ctx context.Context, id int) error {
	sql := `DELETE FROM college WHERE id = $1`
	commandTag, err := c.DB.Pool.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("DeleteCollege: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteCollege: no college found with ID %d", id)
	}
	return nil
}

func (c *collegeRepository) ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
	sql := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college ORDER BY name ASC LIMIT $1 OFFSET $2`
	colleges := []*models.College{}
	err := pgxscan.Select(ctx, c.DB.Pool, &colleges, sql, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("ListColleges: failed to execute query: %w", err)
	}
	return colleges, nil
}
