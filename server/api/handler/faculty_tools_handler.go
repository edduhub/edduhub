package handler

import (
	"errors"
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/services/facultytools"

	"github.com/labstack/echo/v4"
)

type FacultyToolsHandler struct {
	service facultytools.FacultyToolsService
}

func NewFacultyToolsHandler(service facultytools.FacultyToolsService) *FacultyToolsHandler {
	return &FacultyToolsHandler{service: service}
}

func (h *FacultyToolsHandler) ListRubrics(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	role, err := helpers.GetUserRole(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	var facultyID *int
	if role == "faculty" {
		actorID, resolveErr := h.resolveActorUserID(c)
		if resolveErr != nil {
			return resolveErr
		}
		facultyID = &actorID
	}
	if role == "admin" {
		if filterID, parseErr := parseOptionalInt(c.QueryParam("faculty_id")); parseErr != nil {
			return helpers.Error(c, "Invalid faculty_id", http.StatusBadRequest)
		} else if filterID != nil {
			facultyID = filterID
		}
	}

	items, err := h.service.ListRubrics(c.Request().Context(), collegeID, facultyID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch rubrics", http.StatusInternalServerError)
	}
	return helpers.Success(c, map[string]any{"rubrics": items, "total": len(items)}, http.StatusOK)
}

func (h *FacultyToolsHandler) CreateRubric(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var input facultytools.CreateRubricInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	actorID, err := h.resolveActorUserID(c)
	if err != nil {
		return err
	}

	rubric, err := h.service.CreateRubric(c.Request().Context(), collegeID, actorID, &input)
	if err != nil {
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}
	return helpers.Success(c, rubric, http.StatusCreated)
}

func (h *FacultyToolsHandler) GetRubric(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	rubricID, err := strconv.Atoi(c.Param("rubricID"))
	if err != nil {
		return helpers.Error(c, "Invalid rubric ID", http.StatusBadRequest)
	}

	item, err := h.service.GetRubric(c.Request().Context(), collegeID, rubricID)
	if err != nil {
		if errors.Is(err, facultytools.ErrRubricNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Rubric not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, "Failed to fetch rubric", http.StatusInternalServerError)
	}
	return helpers.Success(c, item, http.StatusOK)
}

func (h *FacultyToolsHandler) UpdateRubric(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	rubricID, err := strconv.Atoi(c.Param("rubricID"))
	if err != nil {
		return helpers.Error(c, "Invalid rubric ID", http.StatusBadRequest)
	}

	var input facultytools.UpdateRubricInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	updated, err := h.service.UpdateRubric(c.Request().Context(), collegeID, rubricID, &input)
	if err != nil {
		if errors.Is(err, facultytools.ErrRubricNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Rubric not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}
	return helpers.Success(c, updated, http.StatusOK)
}

func (h *FacultyToolsHandler) DeleteRubric(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	rubricID, err := strconv.Atoi(c.Param("rubricID"))
	if err != nil {
		return helpers.Error(c, "Invalid rubric ID", http.StatusBadRequest)
	}

	if err := h.service.DeleteRubric(c.Request().Context(), collegeID, rubricID); err != nil {
		if errors.Is(err, facultytools.ErrRubricNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Rubric not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, "Failed to delete rubric", http.StatusInternalServerError)
	}
	return helpers.Success(c, map[string]any{"deleted": true}, http.StatusOK)
}

func (h *FacultyToolsHandler) ListOfficeHours(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	role, err := helpers.GetUserRole(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	activeOnly := c.QueryParam("active_only") == "true"
	if role == "student" {
		activeOnly = true
	}

	var facultyID *int
	if role == "faculty" {
		actorID, resolveErr := h.resolveActorUserID(c)
		if resolveErr != nil {
			return resolveErr
		}
		facultyID = &actorID
	}
	if role == "admin" {
		if filterID, parseErr := parseOptionalInt(c.QueryParam("faculty_id")); parseErr != nil {
			return helpers.Error(c, "Invalid faculty_id", http.StatusBadRequest)
		} else if filterID != nil {
			facultyID = filterID
		}
	}

	items, err := h.service.ListOfficeHours(c.Request().Context(), collegeID, facultyID, activeOnly)
	if err != nil {
		return helpers.Error(c, "Failed to fetch office hours", http.StatusInternalServerError)
	}
	return helpers.Success(c, map[string]any{"office_hours": items, "total": len(items)}, http.StatusOK)
}

func (h *FacultyToolsHandler) CreateOfficeHour(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	var input facultytools.CreateOfficeHourInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	actorID, err := h.resolveActorUserID(c)
	if err != nil {
		return err
	}

	slot, err := h.service.CreateOfficeHour(c.Request().Context(), collegeID, actorID, &input)
	if err != nil {
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}
	return helpers.Success(c, slot, http.StatusCreated)
}

func (h *FacultyToolsHandler) GetOfficeHour(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	officeHourID, err := strconv.Atoi(c.Param("officeHourID"))
	if err != nil {
		return helpers.Error(c, "Invalid office hour ID", http.StatusBadRequest)
	}

	item, err := h.service.GetOfficeHour(c.Request().Context(), collegeID, officeHourID)
	if err != nil {
		if errors.Is(err, facultytools.ErrOfficeHourNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Office hour not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, "Failed to fetch office hour", http.StatusInternalServerError)
	}
	return helpers.Success(c, item, http.StatusOK)
}

func (h *FacultyToolsHandler) UpdateOfficeHour(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	officeHourID, err := strconv.Atoi(c.Param("officeHourID"))
	if err != nil {
		return helpers.Error(c, "Invalid office hour ID", http.StatusBadRequest)
	}

	var input facultytools.UpdateOfficeHourInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	updated, err := h.service.UpdateOfficeHour(c.Request().Context(), collegeID, officeHourID, &input)
	if err != nil {
		if errors.Is(err, facultytools.ErrOfficeHourNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Office hour not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}
	return helpers.Success(c, updated, http.StatusOK)
}

func (h *FacultyToolsHandler) DeleteOfficeHour(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	officeHourID, err := strconv.Atoi(c.Param("officeHourID"))
	if err != nil {
		return helpers.Error(c, "Invalid office hour ID", http.StatusBadRequest)
	}

	if err := h.service.DeleteOfficeHour(c.Request().Context(), collegeID, officeHourID); err != nil {
		if errors.Is(err, facultytools.ErrOfficeHourNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Office hour not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, "Failed to delete office hour", http.StatusInternalServerError)
	}
	return helpers.Success(c, map[string]any{"deleted": true}, http.StatusOK)
}

func (h *FacultyToolsHandler) CreateBooking(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}

	var input facultytools.CreateBookingInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	booking, err := h.service.CreateBooking(c.Request().Context(), collegeID, studentID, &input)
	if err != nil {
		if errors.Is(err, facultytools.ErrOfficeHourNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Office hour not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, booking, http.StatusCreated)
}

func (h *FacultyToolsHandler) ListBookings(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	role, err := helpers.GetUserRole(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	officeHourID, err := parseOptionalInt(c.QueryParam("office_hour_id"))
	if err != nil {
		return helpers.Error(c, "Invalid office_hour_id", http.StatusBadRequest)
	}

	var studentID *int
	var facultyID *int

	switch role {
	case "student":
		id, extractErr := helpers.ExtractStudentID(c)
		if extractErr != nil {
			return extractErr
		}
		studentID = &id
	case "faculty":
		actorID, resolveErr := h.resolveActorUserID(c)
		if resolveErr != nil {
			return resolveErr
		}
		facultyID = &actorID
	case "admin":
		if parsed, parseErr := parseOptionalInt(c.QueryParam("student_id")); parseErr != nil {
			return helpers.Error(c, "Invalid student_id", http.StatusBadRequest)
		} else if parsed != nil {
			studentID = parsed
		}
		if parsed, parseErr := parseOptionalInt(c.QueryParam("faculty_id")); parseErr != nil {
			return helpers.Error(c, "Invalid faculty_id", http.StatusBadRequest)
		} else if parsed != nil {
			facultyID = parsed
		}
	}

	items, err := h.service.ListBookings(c.Request().Context(), collegeID, officeHourID, studentID, facultyID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch bookings", http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]any{"bookings": items, "total": len(items)}, http.StatusOK)
}

func (h *FacultyToolsHandler) ListBookingsByOfficeHour(c echo.Context) error {
	officeHourID, err := strconv.Atoi(c.Param("officeHourID"))
	if err != nil {
		return helpers.Error(c, "Invalid office hour ID", http.StatusBadRequest)
	}

	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	role, err := helpers.GetUserRole(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	var studentID *int
	var facultyID *int
	switch role {
	case "student":
		id, extractErr := helpers.ExtractStudentID(c)
		if extractErr != nil {
			return extractErr
		}
		studentID = &id
	case "faculty":
		actorID, resolveErr := h.resolveActorUserID(c)
		if resolveErr != nil {
			return resolveErr
		}
		facultyID = &actorID
	case "admin":
		if parsed, parseErr := parseOptionalInt(c.QueryParam("student_id")); parseErr != nil {
			return helpers.Error(c, "Invalid student_id", http.StatusBadRequest)
		} else if parsed != nil {
			studentID = parsed
		}
	}

	items, err := h.service.ListBookings(c.Request().Context(), collegeID, &officeHourID, studentID, facultyID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch bookings", http.StatusInternalServerError)
	}
	return helpers.Success(c, map[string]any{"bookings": items, "total": len(items)}, http.StatusOK)
}

func (h *FacultyToolsHandler) UpdateBookingStatus(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}
	bookingID, err := strconv.Atoi(c.Param("bookingID"))
	if err != nil {
		return helpers.Error(c, "Invalid booking ID", http.StatusBadRequest)
	}

	role, err := helpers.GetUserRole(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}

	var input facultytools.UpdateBookingStatusInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	var actorStudentID *int
	if role == "student" {
		studentID, extractErr := helpers.ExtractStudentID(c)
		if extractErr != nil {
			return extractErr
		}
		actorStudentID = &studentID
	}

	updated, err := h.service.UpdateBookingStatus(c.Request().Context(), collegeID, bookingID, role, actorStudentID, &input)
	if err != nil {
		switch {
		case errors.Is(err, facultytools.ErrBookingNotFound):
			return helpers.NotFound(c, map[string]any{"error": "Booking not found"}, http.StatusNotFound)
		case errors.Is(err, facultytools.ErrFacultyToolsAccess):
			return helpers.Error(c, "Access denied", http.StatusForbidden)
		case errors.Is(err, facultytools.ErrInvalidStatusTransition):
			return helpers.Error(c, err.Error(), http.StatusBadRequest)
		default:
			return helpers.Error(c, err.Error(), http.StatusBadRequest)
		}
	}
	return helpers.Success(c, updated, http.StatusOK)
}

func (h *FacultyToolsHandler) resolveActorUserID(c echo.Context) (int, error) {
	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return 0, helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}
	userID, err := h.service.ResolveUserIDByKratosID(c.Request().Context(), kratosID)
	if err != nil {
		return 0, helpers.Error(c, "Unable to resolve user", http.StatusUnauthorized)
	}
	return userID, nil
}

func parseOptionalInt(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
