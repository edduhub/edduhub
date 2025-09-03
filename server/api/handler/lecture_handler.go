package handler

import (
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/lecture"

	"github.com/labstack/echo/v4"
)

type LectureHandler struct {
	lectureService lecture.LectureService
}

func NewLectureHandler(lectureService lecture.LectureService) *LectureHandler {
	return &LectureHandler{
		lectureService: lectureService,
	}
}

func (h *LectureHandler) UpdateLecture(c echo.Context) error {
	lectureIDStr := c.Param("lectureID")
	lectureID, err := strconv.Atoi(lectureIDStr)
	if err != nil {
		return helpers.Error(c, "invalid lecture ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var req models.UpdateLectureRequest
	if err := c.Bind(&req); err != nil {
		return helpers.Error(c, "invalid request body", 400)
	}

	err = h.lectureService.UpdateLecturePartial(c.Request().Context(), collegeID, lectureID, &req)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, "Success", 204)
}

func (h *LectureHandler) ListLectures(c echo.Context) error {
	courseIDStr := c.Param("courseID")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return helpers.Error(c, "invalid course ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	// Parse pagination parameters
	limit := uint64(10) // default limit
	offset := uint64(0)  // default offset

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

	lectures, err := h.lectureService.FindLecturesByCourse(c.Request().Context(), collegeID, courseID, limit, offset)
	if err != nil {
		return helpers.Error(c, err.Error(), 500)
	}

	return helpers.Success(c, lectures, 200)
}

func (h *LectureHandler) GetLecture(c echo.Context) error {
	lectureIDStr := c.Param("lectureID")
	lectureID, err := strconv.Atoi(lectureIDStr)
	if err != nil {
		return helpers.Error(c, "invalid lecture ID", 400)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	lecture, err := h.lectureService.GetLectureByID(c.Request().Context(), collegeID, lectureID)
	if err != nil {
		return helpers.Error(c, "lecture not found", 404)
	}

	return helpers.Success(c, lecture, 200)
}