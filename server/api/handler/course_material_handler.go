package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/course_material"

	"github.com/labstack/echo/v4"
)

type CourseMaterialHandler struct {
	courseMaterialService course_material.CourseMaterialService
}

func NewCourseMaterialHandler(courseMaterialService course_material.CourseMaterialService) *CourseMaterialHandler {
	return &CourseMaterialHandler{
		courseMaterialService: courseMaterialService,
	}
}

// --- Module Handlers ---

// CreateModule creates a new course module
// @Summary Create Course Module
// @Description Creates a new module for organizing course materials
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courseID path int true "Course ID"
// @Param module body models.CreateCourseModuleRequest true "Module data"
// @Success 201 {object} models.CourseModule
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/courses/{courseID}/modules [post]
func (h *CourseMaterialHandler) CreateModule(c echo.Context) error {
	ctx := c.Request().Context()

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateCourseModuleRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	module, err := h.courseMaterialService.CreateModule(ctx, courseID, collegeID, userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, module, 201)
}

// GetModule retrieves a specific course module
// @Summary Get Course Module
// @Description Retrieves details of a specific course module
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param moduleID path int true "Module ID"
// @Success 200 {object} models.CourseModule
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/modules/{moduleID} [get]
func (h *CourseMaterialHandler) GetModule(c echo.Context) error {
	ctx := c.Request().Context()

	moduleIDStr := c.Param("moduleID")
	moduleID, err := strconv.Atoi(moduleIDStr)
	if err != nil {
		return helpers.Error(c, "invalid module ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	module, err := h.courseMaterialService.GetModule(ctx, collegeID, moduleID)
	if err != nil {
		return helpers.Error(c, err.Error(), 404)
	}

	return helpers.Success(c, module, 200)
}

// ListModules lists all modules for a course
// @Summary List Course Modules
// @Description Lists all modules for a specific course
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courseID path int true "Course ID"
// @Success 200 {array} models.CourseModule
// @Failure 400 {object} map[string]interface{}
// @Router /api/courses/{courseID}/modules [get]
func (h *CourseMaterialHandler) ListModules(c echo.Context) error {
	ctx := c.Request().Context()

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	modules, err := h.courseMaterialService.ListModules(ctx, collegeID, courseID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, modules, 200)
}

// UpdateModule updates a course module
// @Summary Update Course Module
// @Description Updates an existing course module
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param moduleID path int true "Module ID"
// @Param module body models.UpdateCourseModuleRequest true "Module update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/modules/{moduleID} [put]
func (h *CourseMaterialHandler) UpdateModule(c echo.Context) error {
	ctx := c.Request().Context()

	moduleIDStr := c.Param("moduleID")
	moduleID, err := strconv.Atoi(moduleIDStr)
	if err != nil {
		return helpers.Error(c, "invalid module ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateCourseModuleRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.courseMaterialService.UpdateModule(ctx, collegeID, moduleID, &req); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Module updated successfully",
	}, 200)
}

// DeleteModule deletes a course module
// @Summary Delete Course Module
// @Description Deletes a course module (only if it has no materials)
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param moduleID path int true "Module ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/modules/{moduleID} [delete]
func (h *CourseMaterialHandler) DeleteModule(c echo.Context) error {
	ctx := c.Request().Context()

	moduleIDStr := c.Param("moduleID")
	moduleID, err := strconv.Atoi(moduleIDStr)
	if err != nil {
		return helpers.Error(c, "invalid module ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	if err := h.courseMaterialService.DeleteModule(ctx, collegeID, moduleID); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Module deleted successfully",
	}, 200)
}

// --- Material Handlers ---

// CreateMaterial creates a new course material
// @Summary Create Course Material
// @Description Creates a new material (document, video, link, etc.) for a course
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courseID path int true "Course ID"
// @Param material body models.CreateCourseMaterialRequest true "Material data"
// @Success 201 {object} models.CourseMaterial
// @Failure 400 {object} map[string]interface{}
// @Router /api/courses/{courseID}/materials [post]
func (h *CourseMaterialHandler) CreateMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	var req models.CreateCourseMaterialRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	material, err := h.courseMaterialService.CreateMaterial(ctx, courseID, collegeID, userID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, material, 201)
}

// GetMaterial retrieves a specific course material with details
// @Summary Get Course Material
// @Description Retrieves details of a specific course material including file information
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Success 200 {object} models.CourseMaterialWithDetails
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/materials/{materialID} [get]
func (h *CourseMaterialHandler) GetMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	material, err := h.courseMaterialService.GetMaterial(ctx, collegeID, materialID)
	if err != nil {
		return helpers.Error(c, err.Error(), 404)
	}

	return helpers.Success(c, material, 200)
}

// ListMaterials lists all materials for a course
// @Summary List Course Materials
// @Description Lists all materials for a specific course, optionally filtered by module
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courseID path int true "Course ID"
// @Param module_id query int false "Module ID filter"
// @Param only_published query bool false "Show only published materials"
// @Success 200 {array} models.CourseMaterialWithDetails
// @Failure 400 {object} map[string]interface{}
// @Router /api/courses/{courseID}/materials [get]
func (h *CourseMaterialHandler) ListMaterials(c echo.Context) error {
	ctx := c.Request().Context()

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var moduleID *int
	moduleIDStr := c.QueryParam("module_id")
	if moduleIDStr != "" {
		moduleIDVal, err := strconv.Atoi(moduleIDStr)
		if err == nil {
			moduleID = &moduleIDVal
		}
	}

	onlyPublished := c.QueryParam("only_published") == "true"

	materials, err := h.courseMaterialService.ListMaterials(ctx, collegeID, courseID, moduleID, onlyPublished)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, materials, 200)
}

// UpdateMaterial updates a course material
// @Summary Update Course Material
// @Description Updates an existing course material
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Param material body models.UpdateCourseMaterialRequest true "Material update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID} [put]
func (h *CourseMaterialHandler) UpdateMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateCourseMaterialRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	if err := h.courseMaterialService.UpdateMaterial(ctx, collegeID, materialID, &req); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Material updated successfully",
	}, 200)
}

// DeleteMaterial deletes a course material
// @Summary Delete Course Material
// @Description Deletes a course material
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID} [delete]
func (h *CourseMaterialHandler) DeleteMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	if err := h.courseMaterialService.DeleteMaterial(ctx, collegeID, materialID); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Material deleted successfully",
	}, 200)
}

// PublishMaterial publishes a course material
// @Summary Publish Course Material
// @Description Publishes a material making it visible to students
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID}/publish [post]
func (h *CourseMaterialHandler) PublishMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	if err := h.courseMaterialService.PublishMaterial(ctx, collegeID, materialID); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Material published successfully",
	}, 200)
}

// UnpublishMaterial unpublishes a course material
// @Summary Unpublish Course Material
// @Description Unpublishes a material hiding it from students
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID}/unpublish [post]
func (h *CourseMaterialHandler) UnpublishMaterial(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	if err := h.courseMaterialService.UnpublishMaterial(ctx, collegeID, materialID); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Material unpublished successfully",
	}, 200)
}

// --- Access Tracking Handlers ---

// LogMaterialAccess logs student access to a material
// @Summary Log Material Access
// @Description Logs when a student accesses a material and tracks completion
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Param access body models.MaterialAccessLog true "Access log data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID}/access [post]
func (h *CourseMaterialHandler) LogMaterialAccess(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	userID, err := helpers.ExtractUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		DurationSeconds int  `json:"duration_seconds"`
		Completed       bool `json:"completed"`
	}
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	// Note: In production, you'd need to get the student ID from the user ID
	// For now, we'll use userID as studentID (this should be properly mapped)
	if err := h.courseMaterialService.LogMaterialAccess(ctx, materialID, userID, req.DurationSeconds, req.Completed); err != nil {
		return helpers.Error(c, err.Error(), 400)
	}

	return helpers.Success(c, map[string]any{
		"message": "Access logged successfully",
	}, 200)
}

// GetMaterialAccessStats retrieves access statistics for a material
// @Summary Get Material Access Statistics
// @Description Retrieves statistics about student access to a material
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param materialID path int true "Material ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/materials/{materialID}/stats [get]
func (h *CourseMaterialHandler) GetMaterialAccessStats(c echo.Context) error {
	ctx := c.Request().Context()

	materialIDStr := c.Param("materialID")
	materialID, err := strconv.Atoi(materialIDStr)
	if err != nil {
		return helpers.Error(c, "invalid material ID", 400)
	}

	stats, err := h.courseMaterialService.GetMaterialAccessStats(ctx, materialID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, stats, 200)
}

// GetStudentProgress retrieves student's progress in a course
// @Summary Get Student Course Progress
// @Description Retrieves a student's progress through course materials
// @Tags Course Materials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param courseID path int true "Course ID"
// @Param studentID path int true "Student ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/courses/{courseID}/students/{studentID}/progress [get]
func (h *CourseMaterialHandler) GetStudentProgress(c echo.Context) error {
	ctx := c.Request().Context()

	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	studentIDStr := c.Param("studentID")
	studentID, err := strconv.Atoi(studentIDStr)
	if err != nil {
		return helpers.Error(c, "invalid student ID", 400)
	}

	progress, err := h.courseMaterialService.GetStudentProgress(ctx, courseID, studentID)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, progress, 200)
}
