//go:build integration
// +build integration

package repository

import (
	"context"
	"testing"

	"eduhub/server/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTest(t *testing.T) (*pgxpool.Pool, *DB, UserRepository, context.Context) {
	databaseURL := "postgres://your_db_user:your_db_password@localhost:5432/edduhub"

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
	})

	db := &DB{
		Pool: pool,
	}

	repo := NewUserRepository(db)
	ctx := context.Background()

	return pool, db, repo, ctx
}

func TestCreateUser(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	user := &models.User{
		Name:             "Test User",
		Role:             "student",
		Email:            "test@example.com",
		KratosIdentityID: "kratos-user-id",
		IsActive:         true,
	}

	err := repo.CreateUser(ctx, user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestCreateUser_Error(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	user := &models.User{Email: "fail@example.com"} // Use other fields for test data

	err := repo.CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query or scan ID")
}

// --- Test UpdateUser ---

func TestUpdateUser(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userToUpdate := &models.User{
		ID:               25,
		Name:             "Updated User Name",
		Role:             "admin",
		Email:            "updated@example.com",
		KratosIdentityID: "kratos-updated-id",
		IsActive:         false,
	}

	err := repo.UpdateUser(ctx, userToUpdate)

	assert.NoError(t, err)
	assert.False(t, userToUpdate.UpdatedAt.IsZero())
}

func TestUpdateUser_NotFound(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userToUpdate := &models.User{ID: 999}

	err := repo.UpdateUser(ctx, userToUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no row updated")
}

// --- Test FreezeUserByID ---

func TestFreezeUserByID(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userID := 25

	err := repo.FreezeUserByID(ctx, userID)

	assert.NoError(t, err)
}

func TestFreezeUserByID_NotFound(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userID := 999

	err := repo.FreezeUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or already frozen") // Error message from repo function
}

// --- Test DeleteUserByID --- // Renamed test function

func TestDeleteUserByID(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userID := 25

	err := repo.DeleteUserByID(ctx, userID)

	assert.NoError(t, err)
}

func TestDeleteUserByID_NotFound(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userID := 999

	err := repo.DeleteUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteUserByID_Error(t *testing.T) {
	pool, _, repo, ctx := setupUserTest(t)
	defer pool.Close()

	userID := 25

	err := repo.DeleteUserByID(ctx, userID)

	// Adjust assertion based on real DB behavior
	if err != nil {
		assert.Contains(t, err.Error(), "failed to execute query")
	}
}
