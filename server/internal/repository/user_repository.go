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

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error

	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserPartial(ctx context.Context, userID int, req *models.UpdateUserRequest) error
	FreezeUserByID(ctx context.Context, userID int) error // Changed to operate on ID
	DeleteUserByID(ctx context.Context, userID int) error // Changed to operate on ID
	GetUserByID(ctx context.Context, userID int) (*models.User, error)
	GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error)
	UnFreezeUserByID(ctx context.Context, userID int) error

	FindAllUsers(ctx context.Context, limit, offset uint64) ([]*models.User, error)
	CountUsers(ctx context.Context) (int, error)
}

// userRepository now holds a direct reference to *DB
type userRepository struct {
	DB *DB
}

// NewUserRepository receives the *DB directly
func NewUserRepository(db *DB) UserRepository {
	return &userRepository{
		DB: db,
	}
}

const userTable = "users"

// CreateUser inserts a new user record into the database.
func (u *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Set timestamps if they are zero-valued
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	// Build the INSERT query directly
	sql := `INSERT INTO users (name, role, email, kratos_identity_id, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, name, role, email, kratos_identity_id, is_active, created_at, updated_at`
	args := []any{user.Name, user.Role, user.Email, user.KratosIdentityID, user.IsActive, user.CreatedAt, user.UpdatedAt}

	var result models.User
	err := pgxscan.Get(ctx, u.DB.Pool, &result, sql, args...)
	if err != nil {
		return fmt.Errorf("CreateUser: failed to execute query or scan: %w", err)
	}
	*user = result

	return nil // Success
}

// GetUserByID retrieves a user by their primary ID.
func (u *userRepository) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	sql := `SELECT id, name, role, email, kratos_identity_id, is_active, created_at, updated_at FROM users WHERE id = $1`
	args := []any{userID}

	user := &models.User{}
	err := pgxscan.Get(ctx, u.DB.Pool, user, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetUserByID: user with ID %d not found", userID)
		}
		return nil, fmt.Errorf("GetUserByID: failed to execute query or scan: %w", err)
	}

	return user, nil
}

// GetUserByKratosID retrieves a user by their Kratos identity ID.
func (u *userRepository) GetUserByKratosID(ctx context.Context, kratosID string) (*models.User, error) {
	sql := `SELECT id, name, role, email, kratos_identity_id, is_active, created_at, updated_at FROM users WHERE kratos_identity_id = $1`
	args := []any{kratosID}

	user := &models.User{}
	err := pgxscan.Get(ctx, u.DB.Pool, user, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetUserByKratosID: user with Kratos ID %s not found", kratosID)
		}
		return nil, fmt.Errorf("GetUserByKratosID: failed to execute query or scan: %w", err)
	}

	return user, nil
}

// UnFreezeUserByID sets the IsActive status of a user to true based on their ID.
func (u *userRepository) UnFreezeUserByID(ctx context.Context, userID int) error {
	now := time.Now()
	sql := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3`
	args := []any{true, now, userID}

	commandTag, err := u.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UnFreezeUserByID: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UnFreezeUserByID: user with ID %d not found or already active", userID)
	}

	return nil
}

// FindAllUsers retrieves a paginated list of all users.
func (u *userRepository) FindAllUsers(ctx context.Context, limit, offset uint64) ([]*models.User, error) {
	sql := `SELECT id, name, role, email, kratos_identity_id, is_active, created_at, updated_at FROM users ORDER BY name ASC LIMIT $1 OFFSET $2`
	args := []any{limit, offset}

	users := []*models.User{}
	err := pgxscan.Select(ctx, u.DB.Pool, &users, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("FindAllUsers: failed to execute query or scan: %w", err)
	}

	return users, nil
}

// CountUsers counts the total number of users.
func (u *userRepository) CountUsers(ctx context.Context) (int, error) {
	sql := `SELECT COUNT(*) as count FROM users`

	var result struct {
		Count int `db:"count"`
	}
	err := pgxscan.Get(ctx, u.DB.Pool, &result, sql)
	if err != nil {
		return 0, fmt.Errorf("CountUsers: failed to execute query or scan: %w", err)
	}
	return result.Count, nil
}

// UpdateUser updates an existing user record.
func (u *userRepository) UpdateUser(ctx context.Context, model *models.User) error {
	// Update the UpdatedAt timestamp
	model.UpdatedAt = time.Now()

	sql := `UPDATE users SET name = $1, role = $2, email = $3, kratos_identity_id = $4, is_active = $5, updated_at = $6 WHERE id = $7`
	args := []any{model.Name, model.Role, model.Email, model.KratosIdentityID, model.IsActive, model.UpdatedAt, model.ID}

	commandTag, err := u.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateUser: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateUser: no row updated for ID %d", model.ID)
	}

	return nil // Success
}

// FreezeUser sets the IsActive status of a user to false based on their roll number.
// This implementation updates directly by ID without fetching first.
func (u *userRepository) FreezeUserByID(ctx context.Context, userID int) error {
	now := time.Now()
	sql := `UPDATE users SET is_active = $1, updated_at = $2 WHERE id = $3`
	args := []any{false, now, userID}

	commandTag, err := u.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("FreezeUser: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("FreezeUserByID: user with ID %d not found or already frozen", userID)
	}

	return nil // Success
}

// DeleteUser deletes a user record based on their roll number.
// This implementation deletes directly by ID without fetching first.
func (u *userRepository) DeleteUserByID(ctx context.Context, userID int) error {
	sql := `DELETE FROM users WHERE id = $1`
	args := []any{userID}

	commandTag, err := u.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("DeleteUser: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteUserByID: user with ID %d not found", userID)
	}

	return nil // Success
}

func (u *userRepository) UpdateUserPartial(ctx context.Context, userID int, req *models.UpdateUserRequest) error {
	// Build dynamic query based on non-nil fields
	sql := `UPDATE users SET updated_at = NOW()`
	args := []any{}
	argIndex := 1

	if req.Name != nil {
		sql += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, *req.Name)
		argIndex++
	}
	if req.Role != nil {
		sql += fmt.Sprintf(`, role = $%d`, argIndex)
		args = append(args, *req.Role)
		argIndex++
	}
	if req.Email != nil {
		sql += fmt.Sprintf(`, email = $%d`, argIndex)
		args = append(args, *req.Email)
		argIndex++
	}
	if req.KratosIdentityID != nil {
		sql += fmt.Sprintf(`, kratos_identity_id = $%d`, argIndex)
		args = append(args, *req.KratosIdentityID)
		argIndex++
	}
	if req.IsActive != nil {
		sql += fmt.Sprintf(`, is_active = $%d`, argIndex)
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(args) == 0 {
		return fmt.Errorf("no fields to update")
	}

	sql += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, userID)

	commandTag, err := u.DB.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UpdateUserPartial: failed to execute query: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateUserPartial: user with ID %d not found", userID)
	}

	return nil
}
