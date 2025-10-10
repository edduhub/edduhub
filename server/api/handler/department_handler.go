package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/department"

	"github.com/labstack/echo/v4"
)

type DepartmentHandler struct {
	departmentService department.DepartmentService
}

func NewDepartmentHandler(departmentService department.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		departmentService: departmentService,
	}
}

func (h *DepartmentHandler) GetDepartments(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limit := uint64(10)
	offset := uint64(0)

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseUint(limitParam, 10, 64); err == nil {
			limit = parsedLimit
		}
	}
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.ParseUint(offsetParam, 10, 64); err == nil {
			offset = parsedOffset
		}
	}

	departments, err := h.departmentService.GetDepartments(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, departments, 200)
}

func (h *DepartmentHandler) CreateDepartment(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var department models.Department
	if err := c.Bind(&department); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	department.CollegeID = collegeID

	err = h.departmentService.CreateDepartment(c.Request().Context(), &department)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, department, 201)
}

func (h *DepartmentHandler) GetDepartment(c echo.Context) error {
	departmentIDStr := c.Param("departmentID")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid department ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	department, err := h.departmentService.GetDepartment(c.Request().Context(), collegeID, departmentID)
	if err != nil {
		return helpers.Error(c, "department not found", 404)
	}

	return helpers.Success(c, department, 200)
}

func (h *DepartmentHandler) UpdateDepartment(c echo.Context) error {
	departmentIDStr := c.Param("departmentID")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid department ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.departmentService.UpdateDepartment(c.Request().Context(), collegeID, departmentID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Department updated successfully", 204)
}

func (h *DepartmentHandler) DeleteDepartment(c echo.Context) error {
	departmentIDStr := c.Param("departmentID")
	departmentID, err := strconv.Atoi(departmentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid department ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	err = h.departmentService.DeleteDepartment(c.Request().Context(), collegeID, departmentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Department deleted successfully", 204)
}
