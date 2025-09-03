package handler

import (
	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/college"

	"github.com/labstack/echo/v4"
)

type CollegeHandler struct {
	collegeService college.CollegeService
}

func NewCollegeHandler(collegeService college.CollegeService) *CollegeHandler {
	return &CollegeHandler{
		collegeService: collegeService,
	}
}

func (h *CollegeHandler) UpdateCollegeDetails(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateCollegeRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.collegeService.UpdateCollegePartial(c.Request().Context(), collegeID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Success", 204)
}

func (h *CollegeHandler) GetCollegeDetails(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	collegeDetails, err := h.collegeService.GetCollegeByID(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 404)
	}

	return helpers.Success(c, collegeDetails, 200)
}

func (h *CollegeHandler) GetCollegeStats(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// For now, return a basic response
	// This would need to be implemented based on the actual requirements
	return helpers.Success(c, map[string]interface{}{
		"college_id": collegeID,
		"stats": "Implementation needed",
	}, 200)
}