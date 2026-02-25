//go:build integration

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

func setupCollegeTest(t *testing.T) (pgxmock.PgxPoolIface, *DB, CollegeRepository, context.Context) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})

	db := &DB{
		Pool: mock,
	}

	repo := NewCollegeRepository(db)
	ctx := context.Background()

	return mock, db, repo, ctx
}

func TestCreateCollege(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	college := &models.College{
		Name:    "Test College",
		Address: "123 Test St",
		City:    "Test City",
		State:   "Test State",
		Country: "Test Country",
	}
	expectedID := 1

	sqlRegex := `INSERT INTO college \(name, address, city, state, country, created_at, updated_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING id, name, address, city, state, country, created_at, updated_at`

	rows := pgxmock.NewRows([]string{"id", "name", "address", "city", "state", "country", "created_at", "updated_at"}).
		AddRow(expectedID, college.Name, college.Address, college.City, college.State, college.Country, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(college.Name, college.Address, college.City, college.State, college.Country, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(rows)

	err := repo.CreateCollege(ctx, college)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, college.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateCollege_Error(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	college := &models.College{Name: "Fail College"}
	dbError := errors.New("insert failed")

	sqlRegex := `INSERT INTO college \(name, address, city, state, country, created_at, updated_at\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7\) RETURNING id, name, address, city, state, country, created_at, updated_at`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(college.Name, college.Address, college.City, college.State, college.Country, pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(dbError)

	err := repo.CreateCollege(ctx, college)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCollegeByID(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 1
	expectedCollege := &models.College{
		ID:        collegeID,
		Name:      "Test College",
		Address:   "123 Test St",
		City:      "Test City",
		State:     "Test State",
		Country:   "Test Country",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	sqlRegex := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE id = \$1`
	rows := pgxmock.NewRows([]string{"id", "name", "address", "city", "state", "country", "created_at", "updated_at"}).
		AddRow(expectedCollege.ID, expectedCollege.Name, expectedCollege.Address, expectedCollege.City, expectedCollege.State, expectedCollege.Country, expectedCollege.CreatedAt, expectedCollege.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(collegeID).
		WillReturnRows(rows)

	college, err := repo.GetCollegeByID(ctx, collegeID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCollege.ID, college.ID)
	assert.Equal(t, expectedCollege.Name, college.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCollegeByID_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 999
	sqlRegex := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE id = \$1`

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(collegeID).
		WillReturnError(pgx.ErrNoRows)

	college, err := repo.GetCollegeByID(ctx, collegeID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Nil(t, college)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetCollegeByName(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeName := "Test College"
	expectedCollege := &models.College{
		ID:        1,
		Name:      collegeName,
		Address:   "123 Test St",
		City:      "Test City",
		State:     "Test State",
		Country:   "Test Country",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sqlRegex := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college WHERE name = \$1`
	rows := pgxmock.NewRows([]string{"id", "name", "address", "city", "state", "country", "created_at", "updated_at"}).
		AddRow(expectedCollege.ID, expectedCollege.Name, expectedCollege.Address, expectedCollege.City, expectedCollege.State, expectedCollege.Country, expectedCollege.CreatedAt, expectedCollege.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(collegeName).
		WillReturnRows(rows)

	college, err := repo.GetCollegeByName(ctx, collegeName)

	assert.NoError(t, err)
	assert.Equal(t, expectedCollege.Name, college.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCollege(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	college := &models.College{
		ID:      1,
		Name:    "Updated College",
		Address: "456 New St",
		City:    "New City",
		State:   "New State",
		Country: "New Country",
	}

	sqlRegex := `UPDATE college SET name = \$1, address = \$2, city = \$3, state = \$4, country = \$5, updated_at = \$6 WHERE id = \$7`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(college.Name, college.Address, college.City, college.State, college.Country, pgxmock.AnyArg(), college.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateCollege(ctx, college)

	assert.NoError(t, err)
	assert.False(t, college.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCollege_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	college := &models.College{
		ID:      999,
		Name:    "Non-existent",
		Address: "No Address",
		City:    "No City",
		State:   "No State",
		Country: "No Country",
	}

	sqlRegex := `UPDATE college SET name = \$1, address = \$2, city = \$3, state = \$4, country = \$5, updated_at = \$6 WHERE id = \$7`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(college.Name, college.Address, college.City, college.State, college.Country, pgxmock.AnyArg(), college.ID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.UpdateCollege(ctx, college)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no college found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCollege(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 1

	sqlRegex := `DELETE FROM college WHERE id = \$1`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(collegeID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.DeleteCollege(ctx, collegeID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteCollege_NotFound(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 999

	sqlRegex := `DELETE FROM college WHERE id = \$1`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(collegeID).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.DeleteCollege(ctx, collegeID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no college found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListColleges(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	limit := uint64(10)
	offset := uint64(0)

	expectedColleges := []*models.College{
		{ID: 1, Name: "College 1", City: "City 1", State: "State 1", Country: "Country 1"},
		{ID: 2, Name: "College 2", City: "City 2", State: "State 2", Country: "Country 2"},
	}

	sqlRegex := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college ORDER BY name ASC LIMIT \$1 OFFSET \$2`
	rows := pgxmock.NewRows([]string{"id", "name", "address", "city", "state", "country", "created_at", "updated_at"})
	for _, c := range expectedColleges {
		rows.AddRow(c.ID, c.Name, c.Address, c.City, c.State, c.Country, time.Now(), time.Now())
	}

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(int32(limit), int32(offset)).
		WillReturnRows(rows)

	colleges, err := repo.ListColleges(ctx, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, colleges, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListColleges_Empty(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	limit := uint64(10)
	offset := uint64(0)

	sqlRegex := `SELECT id, name, address, city, state, country, created_at, updated_at FROM college ORDER BY name ASC LIMIT \$1 OFFSET \$2`
	rows := pgxmock.NewRows([]string{"id", "name", "address", "city", "state", "country", "created_at", "updated_at"})

	mock.ExpectQuery(regexp.QuoteMeta(sqlRegex)).
		WithArgs(int32(limit), int32(offset)).
		WillReturnRows(rows)

	colleges, err := repo.ListColleges(ctx, limit, offset)

	assert.NoError(t, err)
	assert.Len(t, colleges, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCollegePartial(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 1
	name := "New College Name"

	sqlRegex := `UPDATE college SET updated_at = NOW\(\), name = \$1 WHERE id = \$2`

	mock.ExpectExec(regexp.QuoteMeta(sqlRegex)).
		WithArgs(name, collegeID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	req := &models.UpdateCollegeRequest{
		Name: &name,
	}

	err := repo.UpdateCollegePartial(ctx, collegeID, req)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateCollegePartial_NoFields(t *testing.T) {
	mock, _, repo, ctx := setupCollegeTest(t)
	defer mock.Close()

	collegeID := 1
	req := &models.UpdateCollegeRequest{}

	err := repo.UpdateCollegePartial(ctx, collegeID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no fields to update")
}
