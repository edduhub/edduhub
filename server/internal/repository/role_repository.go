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

type RoleRepository interface {
	// Role CRUD operations
	CreateRole(ctx context.Context, role *models.Role) error
	GetRoleByID(ctx context.Context, roleID int) (*models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, roleID int) error
	ListRoles(ctx context.Context, filter models.RoleFilter) ([]*models.Role, error)
	CountRoles(ctx context.Context, filter models.RoleFilter) (int, error)

	// Permission CRUD operations
	CreatePermission(ctx context.Context, permission *models.Permission) error
	GetPermissionByID(ctx context.Context, permissionID int) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	ListPermissions(ctx context.Context, filter models.PermissionFilter) ([]*models.Permission, error)
	CountPermissions(ctx context.Context, filter models.PermissionFilter) (int, error)

	// Role-Permission relationships
	AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
	RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]*models.Permission, error)
	RoleHasPermission(ctx context.Context, roleID int, resource, action string) (bool, error)

	// User-Role relationships
	AssignRoleToUser(ctx context.Context, assignment *models.UserRoleAssignment) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error)
	GetUserPermissions(ctx context.Context, userID int) ([]*models.Permission, error)
	UserHasPermission(ctx context.Context, userID int, resource, action string) (bool, error)
	UserHasRole(ctx context.Context, userID int, roleName string) (bool, error)

	// Bulk operations
	GetUsersWithRole(ctx context.Context, roleID int) ([]int, error)
}

type roleRepository struct {
	DB *DB
}

func NewRoleRepository(db *DB) RoleRepository {
	return &roleRepository{DB: db}
}

// Role CRUD operations

func (r *roleRepository) CreateRole(ctx context.Context, role *models.Role) error {
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	sql := `INSERT INTO roles (name, description, college_id, is_system_role, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql, role.Name, role.Description, role.CollegeID, role.IsSystemRole, role.CreatedAt, role.UpdatedAt).Scan(&id)
	if err != nil {
		return fmt.Errorf("CreateRole: failed to execute query: %w", err)
	}
	role.ID = id
	return nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, roleID int) (*models.Role, error) {
	sql := `SELECT id, name, description, college_id, is_system_role, created_at, updated_at
			FROM roles WHERE id = $1`

	role := &models.Role{}
	err := pgxscan.Get(ctx, r.DB.Pool, role, sql, roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetRoleByID: role with ID %d not found: %w", roleID, err)
		}
		return nil, fmt.Errorf("GetRoleByID: failed to execute query: %w", err)
	}
	return role, nil
}

func (r *roleRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	sql := `SELECT id, name, description, college_id, is_system_role, created_at, updated_at
			FROM roles WHERE name = $1`

	role := &models.Role{}
	err := pgxscan.Get(ctx, r.DB.Pool, role, sql, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetRoleByName: role with name %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("GetRoleByName: failed to execute query: %w", err)
	}
	return role, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *models.Role) error {
	role.UpdatedAt = time.Now()

	sql := `UPDATE roles SET name = $1, description = $2, college_id = $3, updated_at = $4
			WHERE id = $5`

	commandTag, err := r.DB.Pool.Exec(ctx, sql, role.Name, role.Description, role.CollegeID, role.UpdatedAt, role.ID)
	if err != nil {
		return fmt.Errorf("UpdateRole: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateRole: no role found with ID %d", role.ID)
	}
	return nil
}

func (r *roleRepository) DeleteRole(ctx context.Context, roleID int) error {
	// Check if role is a system role
	sql := `SELECT is_system_role FROM roles WHERE id = $1`
	var isSystemRole bool
	err := r.DB.Pool.QueryRow(ctx, sql, roleID).Scan(&isSystemRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("DeleteRole: role with ID %d not found", roleID)
		}
		return fmt.Errorf("DeleteRole: failed to check system role: %w", err)
	}

	if isSystemRole {
		return fmt.Errorf("DeleteRole: cannot delete system role")
	}

	sql = `DELETE FROM roles WHERE id = $1`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, roleID)
	if err != nil {
		return fmt.Errorf("DeleteRole: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteRole: no role found with ID %d", roleID)
	}
	return nil
}

func (r *roleRepository) ListRoles(ctx context.Context, filter models.RoleFilter) ([]*models.Role, error) {
	sql := `SELECT id, name, description, college_id, is_system_role, created_at, updated_at
			FROM roles WHERE 1=1`
	args := []any{}
	paramCount := 0

	if filter.CollegeID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND college_id = $%d", paramCount)
		args = append(args, *filter.CollegeID)
	}

	if filter.IsSystemRole != nil {
		paramCount++
		sql += fmt.Sprintf(" AND is_system_role = $%d", paramCount)
		args = append(args, *filter.IsSystemRole)
	}

	if filter.Name != nil {
		paramCount++
		sql += fmt.Sprintf(" AND name ILIKE $%d", paramCount)
		args = append(args, "%"+*filter.Name+"%")
	}

	sql += " ORDER BY name ASC"

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

	var roles []*models.Role
	err := pgxscan.Select(ctx, r.DB.Pool, &roles, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Role{}, nil
		}
		return nil, fmt.Errorf("ListRoles: failed to execute query: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) CountRoles(ctx context.Context, filter models.RoleFilter) (int, error) {
	sql := `SELECT COUNT(*) FROM roles WHERE 1=1`
	args := []any{}
	paramCount := 0

	if filter.CollegeID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND college_id = $%d", paramCount)
		args = append(args, *filter.CollegeID)
	}

	if filter.IsSystemRole != nil {
		paramCount++
		sql += fmt.Sprintf(" AND is_system_role = $%d", paramCount)
		args = append(args, *filter.IsSystemRole)
	}

	if filter.Name != nil {
		paramCount++
		sql += fmt.Sprintf(" AND name ILIKE $%d", paramCount)
		args = append(args, "%"+*filter.Name+"%")
	}

	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountRoles: failed to execute query: %w", err)
	}
	return count, nil
}

// Permission CRUD operations

func (r *roleRepository) CreatePermission(ctx context.Context, permission *models.Permission) error {
	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	sql := `INSERT INTO permissions (name, resource, action, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql, permission.Name, permission.Resource, permission.Action, permission.Description, permission.CreatedAt, permission.UpdatedAt).Scan(&id)
	if err != nil {
		return fmt.Errorf("CreatePermission: failed to execute query: %w", err)
	}
	permission.ID = id
	return nil
}

func (r *roleRepository) GetPermissionByID(ctx context.Context, permissionID int) (*models.Permission, error) {
	sql := `SELECT id, name, resource, action, description, created_at, updated_at
			FROM permissions WHERE id = $1`

	permission := &models.Permission{}
	err := pgxscan.Get(ctx, r.DB.Pool, permission, sql, permissionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetPermissionByID: permission with ID %d not found: %w", permissionID, err)
		}
		return nil, fmt.Errorf("GetPermissionByID: failed to execute query: %w", err)
	}
	return permission, nil
}

func (r *roleRepository) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	sql := `SELECT id, name, resource, action, description, created_at, updated_at
			FROM permissions WHERE name = $1`

	permission := &models.Permission{}
	err := pgxscan.Get(ctx, r.DB.Pool, permission, sql, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetPermissionByName: permission with name %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("GetPermissionByName: failed to execute query: %w", err)
	}
	return permission, nil
}

func (r *roleRepository) ListPermissions(ctx context.Context, filter models.PermissionFilter) ([]*models.Permission, error) {
	sql := `SELECT id, name, resource, action, description, created_at, updated_at
			FROM permissions WHERE 1=1`
	args := []any{}
	paramCount := 0

	if filter.Resource != nil {
		paramCount++
		sql += fmt.Sprintf(" AND resource = $%d", paramCount)
		args = append(args, *filter.Resource)
	}

	if filter.Action != nil {
		paramCount++
		sql += fmt.Sprintf(" AND action = $%d", paramCount)
		args = append(args, *filter.Action)
	}

	sql += " ORDER BY resource ASC, action ASC"

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

	var permissions []*models.Permission
	err := pgxscan.Select(ctx, r.DB.Pool, &permissions, sql, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Permission{}, nil
		}
		return nil, fmt.Errorf("ListPermissions: failed to execute query: %w", err)
	}
	return permissions, nil
}

func (r *roleRepository) CountPermissions(ctx context.Context, filter models.PermissionFilter) (int, error) {
	sql := `SELECT COUNT(*) FROM permissions WHERE 1=1`
	args := []any{}
	paramCount := 0

	if filter.Resource != nil {
		paramCount++
		sql += fmt.Sprintf(" AND resource = $%d", paramCount)
		args = append(args, *filter.Resource)
	}

	if filter.Action != nil {
		paramCount++
		sql += fmt.Sprintf(" AND action = $%d", paramCount)
		args = append(args, *filter.Action)
	}

	var count int
	err := r.DB.Pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountPermissions: failed to execute query: %w", err)
	}
	return count, nil
}

// Role-Permission relationships

func (r *roleRepository) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	// Use transaction to ensure atomicity
	tx, err := r.DB.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AssignPermissionsToRole: failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert role_permissions entries (ignore duplicates)
	for _, permID := range permissionIDs {
		sql := `INSERT INTO role_permissions (role_id, permission_id, created_at)
				VALUES ($1, $2, $3)
				ON CONFLICT (role_id, permission_id) DO NOTHING`
		_, err := tx.Exec(ctx, sql, roleID, permID, time.Now())
		if err != nil {
			return fmt.Errorf("AssignPermissionsToRole: failed to assign permission %d: %w", permID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("AssignPermissionsToRole: failed to commit transaction: %w", err)
	}

	return nil
}

func (r *roleRepository) RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	sql := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = ANY($2)`
	_, err := r.DB.Pool.Exec(ctx, sql, roleID, permissionIDs)
	if err != nil {
		return fmt.Errorf("RemovePermissionsFromRole: failed to execute query: %w", err)
	}
	return nil
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID int) ([]*models.Permission, error) {
	sql := `SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			WHERE rp.role_id = $1
			ORDER BY p.resource ASC, p.action ASC`

	var permissions []*models.Permission
	err := pgxscan.Select(ctx, r.DB.Pool, &permissions, sql, roleID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Permission{}, nil
		}
		return nil, fmt.Errorf("GetRolePermissions: failed to execute query: %w", err)
	}
	return permissions, nil
}

func (r *roleRepository) RoleHasPermission(ctx context.Context, roleID int, resource, action string) (bool, error) {
	sql := `SELECT EXISTS(
				SELECT 1 FROM role_permissions rp
				INNER JOIN permissions p ON rp.permission_id = p.id
				WHERE rp.role_id = $1 AND p.resource = $2 AND p.action = $3
			)`

	var exists bool
	err := r.DB.Pool.QueryRow(ctx, sql, roleID, resource, action).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("RoleHasPermission: failed to execute query: %w", err)
	}
	return exists, nil
}

// User-Role relationships

func (r *roleRepository) AssignRoleToUser(ctx context.Context, assignment *models.UserRoleAssignment) error {
	assignment.AssignedAt = time.Now()

	sql := `INSERT INTO user_role_assignments (user_id, role_id, assigned_by, assigned_at, expires_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (user_id, role_id) DO UPDATE
			SET assigned_by = $3, assigned_at = $4, expires_at = $5
			RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql, assignment.UserID, assignment.RoleID, assignment.AssignedBy, assignment.AssignedAt, assignment.ExpiresAt).Scan(&id)
	if err != nil {
		return fmt.Errorf("AssignRoleToUser: failed to execute query: %w", err)
	}
	assignment.ID = id
	return nil
}

func (r *roleRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	sql := `DELETE FROM user_role_assignments WHERE user_id = $1 AND role_id = $2`
	commandTag, err := r.DB.Pool.Exec(ctx, sql, userID, roleID)
	if err != nil {
		return fmt.Errorf("RemoveRoleFromUser: failed to execute query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("RemoveRoleFromUser: no assignment found for user %d and role %d", userID, roleID)
	}
	return nil
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error) {
	sql := `SELECT r.id, r.name, r.description, r.college_id, r.is_system_role, r.created_at, r.updated_at
			FROM roles r
			INNER JOIN user_role_assignments ura ON r.id = ura.role_id
			WHERE ura.user_id = $1 AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
			ORDER BY r.name ASC`

	var roles []*models.Role
	err := pgxscan.Select(ctx, r.DB.Pool, &roles, sql, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Role{}, nil
		}
		return nil, fmt.Errorf("GetUserRoles: failed to execute query: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) GetUserPermissions(ctx context.Context, userID int) ([]*models.Permission, error) {
	sql := `SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN user_role_assignments ura ON rp.role_id = ura.role_id
			WHERE ura.user_id = $1 AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
			ORDER BY p.resource ASC, p.action ASC`

	var permissions []*models.Permission
	err := pgxscan.Select(ctx, r.DB.Pool, &permissions, sql, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []*models.Permission{}, nil
		}
		return nil, fmt.Errorf("GetUserPermissions: failed to execute query: %w", err)
	}
	return permissions, nil
}

func (r *roleRepository) UserHasPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	sql := `SELECT EXISTS(
				SELECT 1 FROM user_role_assignments ura
				INNER JOIN role_permissions rp ON ura.role_id = rp.role_id
				INNER JOIN permissions p ON rp.permission_id = p.id
				WHERE ura.user_id = $1
				AND p.resource = $2
				AND p.action = $3
				AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
			)`

	var exists bool
	err := r.DB.Pool.QueryRow(ctx, sql, userID, resource, action).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("UserHasPermission: failed to execute query: %w", err)
	}
	return exists, nil
}

func (r *roleRepository) UserHasRole(ctx context.Context, userID int, roleName string) (bool, error) {
	sql := `SELECT EXISTS(
				SELECT 1 FROM user_role_assignments ura
				INNER JOIN roles r ON ura.role_id = r.id
				WHERE ura.user_id = $1
				AND r.name = $2
				AND (ura.expires_at IS NULL OR ura.expires_at > NOW())
			)`

	var exists bool
	err := r.DB.Pool.QueryRow(ctx, sql, userID, roleName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("UserHasRole: failed to execute query: %w", err)
	}
	return exists, nil
}

func (r *roleRepository) GetUsersWithRole(ctx context.Context, roleID int) ([]int, error) {
	sql := `SELECT DISTINCT user_id
			FROM user_role_assignments
			WHERE role_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	var userIDs []int
	rows, err := r.DB.Pool.Query(ctx, sql, roleID)
	if err != nil {
		return nil, fmt.Errorf("GetUsersWithRole: failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("GetUsersWithRole: failed to scan row: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetUsersWithRole: error iterating rows: %w", err)
	}

	return userIDs, nil
}
