package handler

import (
	"net/http"
	"strconv"

	"eduhub/server/internal/middleware"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/role"

	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	roleService role.RoleService
}

func NewRoleHandler(roleService role.RoleService) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
	}
}

func (h *RoleHandler) CreateRole(c echo.Context) error {
	var req models.CreateRoleRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	userID := c.Get("user_id").(int)
	role, err := h.roleService.CreateRole(c.Request().Context(), &req, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create role: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Role created successfully",
		"data":    role,
	})
}

func (h *RoleHandler) GetRole(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	roleWithPerms, err := h.roleService.GetRoleWithPermissions(c.Request().Context(), roleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Role not found: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": roleWithPerms,
	})
}

func (h *RoleHandler) ListRoles(c echo.Context) error {
	collegeID := c.Get("college_id").(int)

	filter := models.RoleFilter{
		CollegeID: &collegeID,
	}

	roles, _, err := h.roleService.ListRoles(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list roles: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": roles,
	})
}

func (h *RoleHandler) UpdateRole(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	var req models.UpdateRoleRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	role, err := h.roleService.UpdateRole(c.Request().Context(), roleID, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update role: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role updated successfully",
		"data":    role,
	})
}

func (h *RoleHandler) DeleteRole(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	if err := h.roleService.DeleteRole(c.Request().Context(), roleID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete role: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role deleted successfully",
	})
}

func (h *RoleHandler) ListPermissions(c echo.Context) error {
	filter := models.PermissionFilter{}

	permissions, _, err := h.roleService.ListPermissions(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list permissions: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": permissions,
	})
}

func (h *RoleHandler) AssignPermissionsToRole(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("roleID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid role ID")
	}

	var req models.AssignPermissionsRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.roleService.AssignPermissionsToRole(c.Request().Context(), roleID, req.PermissionIDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign permissions: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Permissions assigned successfully",
	})
}

func (h *RoleHandler) AssignRoleToUser(c echo.Context) error {
	var req models.AssignRoleRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	assignedBy := c.Get("user_id").(int)

	if err := h.roleService.AssignRoleToUser(c.Request().Context(), &req, assignedBy); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign role: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Role assigned to user successfully",
	})
}

func (h *RoleHandler) GetUserRoles(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	roles, err := h.roleService.GetUserRoles(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user roles: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": roles,
	})
}
