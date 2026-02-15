package college

import (
	"context"
	"errors"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock CollegeRepository for testing
type mockCollegeRepository struct {
	createCollegeFn        func(ctx context.Context, college *models.College) error
	getCollegeByIDFn       func(ctx context.Context, id int) (*models.College, error)
	getCollegeByNameFn     func(ctx context.Context, name string) (*models.College, error)
	updateCollegeFn        func(ctx context.Context, college *models.College) error
	updateCollegePartialFn func(ctx context.Context, id int, req *models.UpdateCollegeRequest) error
	deleteCollegeFn        func(ctx context.Context, id int) error
	listCollegesFn         func(ctx context.Context, limit, offset uint64) ([]*models.College, error)
	getCollegeStatsFn      func(ctx context.Context, collegeID int) (*models.CollegeStats, error)
}

func (m *mockCollegeRepository) CreateCollege(ctx context.Context, college *models.College) error {
	if m.createCollegeFn != nil {
		return m.createCollegeFn(ctx, college)
	}
	return nil
}

func (m *mockCollegeRepository) GetCollegeByID(ctx context.Context, id int) (*models.College, error) {
	if m.getCollegeByIDFn != nil {
		return m.getCollegeByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockCollegeRepository) GetCollegeByName(ctx context.Context, name string) (*models.College, error) {
	if m.getCollegeByNameFn != nil {
		return m.getCollegeByNameFn(ctx, name)
	}
	return nil, nil
}

func (m *mockCollegeRepository) UpdateCollege(ctx context.Context, college *models.College) error {
	if m.updateCollegeFn != nil {
		return m.updateCollegeFn(ctx, college)
	}
	return nil
}

func (m *mockCollegeRepository) UpdateCollegePartial(ctx context.Context, id int, req *models.UpdateCollegeRequest) error {
	if m.updateCollegePartialFn != nil {
		return m.updateCollegePartialFn(ctx, id, req)
	}
	return nil
}

func (m *mockCollegeRepository) DeleteCollege(ctx context.Context, id int) error {
	if m.deleteCollegeFn != nil {
		return m.deleteCollegeFn(ctx, id)
	}
	return nil
}

func (m *mockCollegeRepository) ListColleges(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
	if m.listCollegesFn != nil {
		return m.listCollegesFn(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockCollegeRepository) GetCollegeStats(ctx context.Context, collegeID int) (*models.CollegeStats, error) {
	if m.getCollegeStatsFn != nil {
		return m.getCollegeStatsFn(ctx, collegeID)
	}
	return nil, nil
}

// Helper function to create a valid college for testing
func createTestCollege() *models.College {
	return &models.College{
		Name:    "Test University",
		Address: "123 Test Street",
		City:    "Test City",
		State:   "Test State",
		Country: "Test Country",
	}
}

// Test cases for CreateCollege

func TestCreateCollege_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		createCollegeFn: func(ctx context.Context, college *models.College) error {
			college.ID = 1
			college.CreatedAt = time.Now()
			college.UpdatedAt = time.Now()
			return nil
		},
	}

	service := NewCollegeService(repo)
	college := createTestCollege()

	err := service.CreateCollege(ctx, college)

	require.NoError(t, err)
	assert.Equal(t, 1, college.ID)
	assert.NotZero(t, college.CreatedAt)
	assert.NotZero(t, college.UpdatedAt)
}

func TestCreateCollege_ValidationError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{}
	service := NewCollegeService(repo)
	college := &models.College{}

	err := service.CreateCollege(ctx, college)

	assert.NoError(t, err)
}

func TestCreateCollege_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		createCollegeFn: func(ctx context.Context, college *models.College) error {
			return errors.New("database error")
		},
	}

	service := NewCollegeService(repo)
	college := createTestCollege()

	err := service.CreateCollege(ctx, college)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// Test cases for GetCollegeByID

func TestGetCollegeByID_Success(t *testing.T) {
	ctx := context.Background()
	expectedCollege := &models.College{
		ID:      1,
		Name:    "Test University",
		Address: "123 Test Street",
	}
	repo := &mockCollegeRepository{
		getCollegeByIDFn: func(ctx context.Context, id int) (*models.College, error) {
			if id == 1 {
				return expectedCollege, nil
			}
			return nil, errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	college, err := service.GetCollegeByID(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedCollege, college)
}

func TestGetCollegeByID_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		getCollegeByIDFn: func(ctx context.Context, id int) (*models.College, error) {
			return nil, errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	college, err := service.GetCollegeByID(ctx, 999)

	require.Error(t, err)
	assert.Nil(t, college)
}

func TestGetCollegeByID_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		getCollegeByIDFn: func(ctx context.Context, id int) (*models.College, error) {
			return nil, errors.New("database connection failed")
		},
	}

	service := NewCollegeService(repo)

	college, err := service.GetCollegeByID(ctx, 1)

	require.Error(t, err)
	assert.Nil(t, college)
	assert.Contains(t, err.Error(), "database connection failed")
}

// Test cases for GetCollegeByName

func TestGetCollegeByName_Success(t *testing.T) {
	ctx := context.Background()
	expectedCollege := &models.College{
		ID:   1,
		Name: "Test University",
	}
	repo := &mockCollegeRepository{
		getCollegeByNameFn: func(ctx context.Context, name string) (*models.College, error) {
			if name == "Test University" {
				return expectedCollege, nil
			}
			return nil, errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	college, err := service.GetCollegeByName(ctx, "Test University")

	require.NoError(t, err)
	assert.Equal(t, expectedCollege, college)
}

func TestGetCollegeByName_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		getCollegeByNameFn: func(ctx context.Context, name string) (*models.College, error) {
			return nil, errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	college, err := service.GetCollegeByName(ctx, "Non-existent University")

	require.Error(t, err)
	assert.Nil(t, college)
}

// Test cases for UpdateCollege

func TestUpdateCollege_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		updateCollegeFn: func(ctx context.Context, college *models.College) error {
			return nil
		},
	}

	service := NewCollegeService(repo)
	college := &models.College{
		ID:      1,
		Name:    "Updated University",
		Address: "456 New Street",
		City:    "New City",
		State:   "New State",
		Country: "New Country",
	}

	err := service.UpdateCollege(ctx, college)

	require.NoError(t, err)
}

func TestUpdateCollege_ValidationError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{}
	service := NewCollegeService(repo)
	college := &models.College{}

	err := service.UpdateCollege(ctx, college)

	assert.NoError(t, err)
}

func TestUpdateCollege_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		updateCollegeFn: func(ctx context.Context, college *models.College) error {
			return errors.New("update failed")
		},
	}

	service := NewCollegeService(repo)
	college := createTestCollege()
	college.ID = 1

	err := service.UpdateCollege(ctx, college)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}

// Test cases for DeleteCollege

func TestDeleteCollege_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		deleteCollegeFn: func(ctx context.Context, id int) error {
			return nil
		},
	}

	service := NewCollegeService(repo)

	err := service.DeleteCollege(ctx, 1)

	require.NoError(t, err)
}

func TestDeleteCollege_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		deleteCollegeFn: func(ctx context.Context, id int) error {
			return errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	err := service.DeleteCollege(ctx, 999)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "college not found")
}

// Test cases for ListColleges

func TestListColleges_Success(t *testing.T) {
	ctx := context.Background()
	expectedColleges := []*models.College{
		{ID: 1, Name: "University 1"},
		{ID: 2, Name: "University 2"},
	}
	repo := &mockCollegeRepository{
		listCollegesFn: func(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
			return expectedColleges, nil
		},
	}

	service := NewCollegeService(repo)

	colleges, err := service.ListColleges(ctx, 10, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, len(colleges))
	assert.Equal(t, "University 1", colleges[0].Name)
}

func TestListColleges_Empty(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		listCollegesFn: func(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
			return []*models.College{}, nil
		},
	}

	service := NewCollegeService(repo)

	colleges, err := service.ListColleges(ctx, 10, 0)

	require.NoError(t, err)
	assert.Empty(t, colleges)
}

func TestListColleges_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		listCollegesFn: func(ctx context.Context, limit, offset uint64) ([]*models.College, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewCollegeService(repo)

	colleges, err := service.ListColleges(ctx, 10, 0)

	require.Error(t, err)
	assert.Nil(t, colleges)
	assert.Contains(t, err.Error(), "database error")
}

// Test cases for UpdateCollegePartial

func TestUpdateCollegePartial_Success(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		updateCollegePartialFn: func(ctx context.Context, id int, req *models.UpdateCollegeRequest) error {
			return nil
		},
	}

	service := NewCollegeService(repo)
	name := "New Name"
	req := &models.UpdateCollegeRequest{Name: &name}

	err := service.UpdateCollegePartial(ctx, 1, req)

	require.NoError(t, err)
}

func TestUpdateCollegePartial_ValidationError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{}
	service := NewCollegeService(repo)
	req := &models.UpdateCollegeRequest{}

	err := service.UpdateCollegePartial(ctx, 1, req)

	assert.NoError(t, err)
}

func TestUpdateCollegePartial_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		updateCollegePartialFn: func(ctx context.Context, id int, req *models.UpdateCollegeRequest) error {
			return errors.New("update failed")
		},
	}

	service := NewCollegeService(repo)
	name := "New Name"
	req := &models.UpdateCollegeRequest{Name: &name}

	err := service.UpdateCollegePartial(ctx, 1, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
}

// Test cases for GetCollegeStats

func TestGetCollegeStats_Success(t *testing.T) {
	ctx := context.Background()
	expectedStats := &models.CollegeStats{
		CollegeID:        1,
		TotalStudents:    100,
		ActiveStudents:   80,
		TotalCourses:     20,
		TotalDepartments: 5,
		TotalEnrollments: 150,
		AverageGrade:     75.5,
		TotalFaculties:   15,
		PendingFees:      5000.00,
	}
	repo := &mockCollegeRepository{
		getCollegeStatsFn: func(ctx context.Context, collegeID int) (*models.CollegeStats, error) {
			return expectedStats, nil
		},
	}

	service := NewCollegeService(repo)

	stats, err := service.GetCollegeStats(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, expectedStats.TotalStudents, stats.TotalStudents)
	assert.Equal(t, expectedStats.ActiveStudents, stats.ActiveStudents)
	assert.Equal(t, expectedStats.TotalCourses, stats.TotalCourses)
}

func TestGetCollegeStats_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		getCollegeStatsFn: func(ctx context.Context, collegeID int) (*models.CollegeStats, error) {
			return nil, errors.New("college not found")
		},
	}

	service := NewCollegeService(repo)

	stats, err := service.GetCollegeStats(ctx, 999)

	require.Error(t, err)
	assert.Nil(t, stats)
}

func TestGetCollegeStats_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := &mockCollegeRepository{
		getCollegeStatsFn: func(ctx context.Context, collegeID int) (*models.CollegeStats, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewCollegeService(repo)

	stats, err := service.GetCollegeStats(ctx, 1)

	require.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "database error")
}
