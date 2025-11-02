package role

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type RoleService interface {
	// Role management
	CreateRole(ctx context.Context, req *models.CreateRoleRequest, createdBy int) (*models.Role, error)
	GetRole(ctx context.Context, roleID int) (*models.Role, error)
	GetRoleWithPermissions(ctx context.Context, roleID int) (*models.RoleWithPermissions, error)
	UpdateRole(ctx context.Context, roleID int, req *models.UpdateRoleRequest) (*models.Role, error)
	DeleteRole(ctx context.Context, roleID int) error
	ListRoles(ctx context.Context, filter models.RoleFilter) ([]*models.Role, int, error)

	// Permission management
	ListPermissions(ctx context.Context, filter models.PermissionFilter) ([]*models.Permission, int, error)
	GetPermission(ctx context.Context, permissionID int) (*models.Permission, error)

	// Role-Permission management
	AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
	RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]*models.Permission, error)

	// User-Role management
	AssignRoleToUser(ctx context.Context, req *models.AssignRoleRequest, assignedBy int) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID int) error
	GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error)
	GetUserPermissions(ctx context.Context, userID int) ([]*models.Permission, error)

	// Permission checking
	UserHasPermission(ctx context.Context, userID int, resource, action string) (bool, error)
	UserHasRole(ctx context.Context, userID int, roleName string) (bool, error)
}

type roleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) RoleService {
	return &roleService{
		roleRepo: roleRepo,
	}
}

func (s *roleService) CreateRole(ctx context.Context, req *models.CreateRoleRequest, createdBy int) (*models.Role, error) {
	role := &models.Role{
		Name:        req.Name,
		Description: req.Description,
		CollegeID:   req.CollegeID,
	}

	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Assign permissions if provided
	if len(req.Permissions) > 0 {
		if err := s.roleRepo.AssignPermissionsToRole(ctx, role.ID, req.Permissions); err != nil {
			return nil, fmt.Errorf("failed to assign permissions to role: %w", err)
		}
	}

	return role, nil
}

func (s *roleService) GetRole(ctx context.Context, roleID int) (*models.Role, error) {
	return s.roleRepo.GetRoleByID(ctx, roleID)
}

func (s *roleService) GetRoleWithPermissions(ctx context.Context, roleID int) (*models.RoleWithPermissions, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	permissionNames := make([]string, len(permissions))
	for i, p := range permissions {
		permissionNames[i] = p.Name
	}

	return &models.RoleWithPermissions{
		Role:            *role,
		PermissionNames: permissionNames,
	}, nil
}

func (s *roleService) UpdateRole(ctx context.Context, roleID int, req *models.UpdateRoleRequest) (*models.Role, error) {
	role, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if role.IsSystemRole {
		return nil, fmt.Errorf("cannot update system role")
	}

	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	if err := s.roleRepo.UpdateRole(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Update permissions if provided
	if req.Permissions != nil {
		// Get current permissions
		currentPerms, err := s.roleRepo.GetRolePermissions(ctx, roleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current permissions: %w", err)
		}

		// Build map of current permission IDs
		currentPermMap := make(map[int]bool)
		for _, p := range currentPerms {
			currentPermMap[p.ID] = true
		}

		// Find permissions to add and remove
		newPermMap := make(map[int]bool)
		for _, id := range *req.Permissions {
			newPermMap[id] = true
		}

		var toAdd, toRemove []int
		for id := range newPermMap {
			if !currentPermMap[id] {
				toAdd = append(toAdd, id)
			}
		}
		for id := range currentPermMap {
			if !newPermMap[id] {
				toRemove = append(toRemove, id)
			}
		}

		if len(toAdd) > 0 {
			if err := s.roleRepo.AssignPermissionsToRole(ctx, roleID, toAdd); err != nil {
				return nil, fmt.Errorf("failed to assign new permissions: %w", err)
			}
		}
		if len(toRemove) > 0 {
			if err := s.roleRepo.RemovePermissionsFromRole(ctx, roleID, toRemove); err != nil {
				return nil, fmt.Errorf("failed to remove permissions: %w", err)
			}
		}
	}

	return role, nil
}

func (s *roleService) DeleteRole(ctx context.Context, roleID int) error {
	return s.roleRepo.DeleteRole(ctx, roleID)
}

func (s *roleService) ListRoles(ctx context.Context, filter models.RoleFilter) ([]*models.Role, int, error) {
	roles, err := s.roleRepo.ListRoles(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.roleRepo.CountRoles(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return roles, count, nil
}

func (s *roleService) ListPermissions(ctx context.Context, filter models.PermissionFilter) ([]*models.Permission, int, error) {
	permissions, err := s.roleRepo.ListPermissions(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.roleRepo.CountPermissions(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return permissions, count, nil
}

func (s *roleService) GetPermission(ctx context.Context, permissionID int) (*models.Permission, error) {
	return s.roleRepo.GetPermissionByID(ctx, permissionID)
}

func (s *roleService) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	// Verify role exists
	_, err := s.roleRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	return s.roleRepo.AssignPermissionsToRole(ctx, roleID, permissionIDs)
}

func (s *roleService) RemovePermissionsFromRole(ctx context.Context, roleID int, permissionIDs []int) error {
	return s.roleRepo.RemovePermissionsFromRole(ctx, roleID, permissionIDs)
}

func (s *roleService) GetRolePermissions(ctx context.Context, roleID int) ([]*models.Permission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

func (s *roleService) AssignRoleToUser(ctx context.Context, req *models.AssignRoleRequest, assignedBy int) error {
	// Verify role exists
	_, err := s.roleRepo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	assignment := &models.UserRoleAssignment{
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		AssignedBy: &assignedBy,
		ExpiresAt:  req.ExpiresAt,
	}

	return s.roleRepo.AssignRoleToUser(ctx, assignment)
}

func (s *roleService) RemoveRoleFromUser(ctx context.Context, userID, roleID int) error {
	return s.roleRepo.RemoveRoleFromUser(ctx, userID, roleID)
}

func (s *roleService) GetUserRoles(ctx context.Context, userID int) ([]*models.Role, error) {
	return s.roleRepo.GetUserRoles(ctx, userID)
}

func (s *roleService) GetUserPermissions(ctx context.Context, userID int) ([]*models.Permission, error) {
	return s.roleRepo.GetUserPermissions(ctx, userID)
}

func (s *roleService) UserHasPermission(ctx context.Context, userID int, resource, action string) (bool, error) {
	return s.roleRepo.UserHasPermission(ctx, userID, resource, action)
}

func (s *roleService) UserHasRole(ctx context.Context, userID int, roleName string) (bool, error) {
	return s.roleRepo.UserHasRole(ctx, userID, roleName)
}
