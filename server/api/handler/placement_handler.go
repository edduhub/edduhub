package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/placement"

	"github.com/labstack/echo/v4"
)

type PlacementHandler struct {
	placementService placement.PlacementService
}

func NewPlacementHandler(placementService placement.PlacementService) *PlacementHandler {
	return &PlacementHandler{
		placementService: placementService,
	}
}

// CreatePlacement creates a new placement record
// POST /api/placements
func (h *PlacementHandler) CreatePlacement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var placement models.Placement
	if err := c.Bind(&placement); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	placement.CollegeID = collegeID

	if err := h.placementService.CreatePlacement(c.Request().Context(), &placement); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, placement, 201)
}

// GetPlacement retrieves a placement by ID
// GET /api/placements/:placementID
func (h *PlacementHandler) GetPlacement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	placementID, err := strconv.Atoi(c.Param("placementID"))
	if err != nil {
		return helpers.Error(c, "invalid placement ID", 400)
	}

	placement, err := h.placementService.GetPlacement(c.Request().Context(), collegeID, placementID)
	if err != nil {
		return helpers.Error(c, "placement not found", 404)
	}

	return helpers.Success(c, placement, 200)
}

// ListPlacements lists all placements with optional filters
// GET /api/placements
func (h *PlacementHandler) ListPlacements(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	limit := uint64(50)
	offset := uint64(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.ParseUint(l, 10, 64); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.ParseUint(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	placements, err := h.placementService.ListPlacementsByCollege(c.Request().Context(), collegeID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, placements, 200)
}

// ListPlacementsByStudent lists placements for a specific student
// GET /api/students/:studentID/placements
func (h *PlacementHandler) ListPlacementsByStudent(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	limit := uint64(50)
	offset := uint64(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.ParseUint(l, 10, 64); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.ParseUint(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	placements, err := h.placementService.ListPlacementsByStudent(c.Request().Context(), collegeID, studentID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, placements, 200)
}

// ListPlacementsByCompany lists placements for a specific company
// GET /api/placements/company/:companyName
func (h *PlacementHandler) ListPlacementsByCompany(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	companyName := c.Param("companyName")
	if companyName == "" {
		return helpers.Error(c, "company name is required", 400)
	}

	limit := uint64(50)
	offset := uint64(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.ParseUint(l, 10, 64); err == nil {
			limit = parsed
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.ParseUint(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	placements, err := h.placementService.ListPlacementsByCompany(c.Request().Context(), collegeID, companyName, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, placements, 200)
}

// UpdatePlacement updates a placement
// PUT /api/placements/:placementID
func (h *PlacementHandler) UpdatePlacement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	placementID, err := strconv.Atoi(c.Param("placementID"))
	if err != nil {
		return helpers.Error(c, "invalid placement ID", 400)
	}

	var placement models.Placement
	if err := c.Bind(&placement); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	placement.ID = placementID
	placement.CollegeID = collegeID

	if err := h.placementService.UpdatePlacement(c.Request().Context(), &placement); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, "placement updated successfully", 200)
}

// DeletePlacement deletes a placement
// DELETE /api/placements/:placementID
func (h *PlacementHandler) DeletePlacement(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	placementID, err := strconv.Atoi(c.Param("placementID"))
	if err != nil {
		return helpers.Error(c, "invalid placement ID", 400)
	}

	if err := h.placementService.DeletePlacement(c.Request().Context(), collegeID, placementID); err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "placement deleted successfully", 200)
}

// GetPlacementStats retrieves placement statistics for the college
// GET /api/placements/stats
func (h *PlacementHandler) GetPlacementStats(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	stats, err := h.placementService.GetPlacementStats(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}

// GetCompanyStats retrieves statistics grouped by company
// GET /api/placements/company-stats
func (h *PlacementHandler) GetCompanyStats(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	stats, err := h.placementService.GetCompanyStats(c.Request().Context(), collegeID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}

// GetStudentPlacementCount retrieves the number of placements for a student
// GET /api/students/:studentID/placement-count
func (h *PlacementHandler) GetStudentPlacementCount(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	studentID, err := strconv.Atoi(c.Param("studentID"))
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	count, err := h.placementService.GetStudentPlacementCount(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, map[string]int{"count": count}, 200)
}
