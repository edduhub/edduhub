//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eduhub/server/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserTest(t *testing.T) (*pgxpool.Pool, UserRepository, context.Context) {
	t.Helper()

	ctx, db, pool := setupIntegrationDB(t, "users")
	repo := NewUserRepository(db)

	return pool, repo, ctx
}

func createFixtureUser(t *testing.T, repo UserRepository, ctx context.Context) *models.User {
	t.Helper()

	user := &models.User{
		Name:             fmt.Sprintf("Test User %d", time.Now().UnixNano()),
		Role:             "student",
		Email:            fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		KratosIdentityID: fmt.Sprintf("kratos-user-%d", time.Now().UnixNano()),
		IsActive:         true,
	}

	require.NoError(t, repo.CreateUser(ctx, user))
	t.Cleanup(func() {
		_ = repo.DeleteUserByID(ctx, user.ID)
	})

	return user
}

func TestCreateUser(t *testing.T) {
	_, repo, ctx := setupUserTest(t)

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
	_, repo, ctx := setupUserTest(t)

	user := &models.User{Email: "fail@example.com"} // Use other fields for test data

	err := repo.CreateUser(ctx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query or scan ID")
}

// --- Test UpdateUser ---

func TestUpdateUser(t *testing.T) {
	_, repo, ctx := setupUserTest(t)
	user := createFixtureUser(t, repo, ctx)

	userToUpdate := &models.User{
		ID:               user.ID,
		Name:             "Updated User Name",
		Role:             "admin",
		Email:            "updated-" + user.Email,
		KratosIdentityID: user.KratosIdentityID + "-updated",
		IsActive:         false,
	}

	err := repo.UpdateUser(ctx, userToUpdate)

	assert.NoError(t, err)
	assert.False(t, userToUpdate.UpdatedAt.IsZero())
}

func TestUpdateUser_NotFound(t *testing.T) {
	_, repo, ctx := setupUserTest(t)

	userToUpdate := &models.User{ID: 999}

	err := repo.UpdateUser(ctx, userToUpdate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no row updated")
}

// --- Test FreezeUserByID ---

func TestFreezeUserByID(t *testing.T) {
	_, repo, ctx := setupUserTest(t)
	user := createFixtureUser(t, repo, ctx)

	userID := user.ID

	err := repo.FreezeUserByID(ctx, userID)

	assert.NoError(t, err)
	frozenUser, err := repo.GetUserByID(ctx, userID)
	require.NoError(t, err)
	assert.False(t, frozenUser.IsActive)
}

func TestFreezeUserByID_NotFound(t *testing.T) {
	_, repo, ctx := setupUserTest(t)

	userID := 999

	err := repo.FreezeUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found or already frozen") // Error message from repo function
}

// --- Test DeleteUserByID --- // Renamed test function

func TestDeleteUserByID(t *testing.T) {
	_, repo, ctx := setupUserTest(t)
	user := createFixtureUser(t, repo, ctx)

	userID := user.ID

	err := repo.DeleteUserByID(ctx, userID)

	assert.NoError(t, err)
	_, err = repo.GetUserByID(ctx, userID)
	assert.Error(t, err)
}

func TestDeleteUserByID_NotFound(t *testing.T) {
	_, repo, ctx := setupUserTest(t)

	userID := 999

	err := repo.DeleteUserByID(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteUserByID_Error(t *testing.T) {
	_, repo, ctx := setupUserTest(t)
	user := createFixtureUser(t, repo, ctx)

	userID := user.ID

	err := repo.DeleteUserByID(ctx, userID)
	require.NoError(t, err)

	err = repo.DeleteUserByID(ctx, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteUserByID")
}
