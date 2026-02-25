package handler

import (
	"errors"
	"net/http"
	"strconv"

	"eduhub/server/internal/helpers"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/selfservice"

	"github.com/labstack/echo/v4"
)

// SelfServiceHandler handles student self-service requests.
type SelfServiceHandler struct {
	service selfservice.SelfServiceService
}

// NewSelfServiceHandler creates a new SelfServiceHandler.
func NewSelfServiceHandler(service selfservice.SelfServiceService) *SelfServiceHandler {
	return &SelfServiceHandler{service: service}
}

// GetMyRequests godoc
// @Summary Get student's self-service requests
// @Description Returns a list of self-service requests for the authenticated student
// @Tags Self-Service
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/self-service/requests [get]
func (h *SelfServiceHandler) GetMyRequests(c echo.Context) error {
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	requests, err := h.service.ListStudentRequests(c.Request().Context(), collegeID, studentID)
	if err != nil {
		return helpers.Error(c, "Failed to fetch requests", http.StatusInternalServerError)
	}

	return helpers.Success(c, map[string]any{
		"requests": requests,
		"total":    len(requests),
	}, http.StatusOK)
}

// GetRequest godoc
// @Summary Get a specific self-service request
// @Description Returns details of a specific request
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param requestID path int true "Request ID"
// @Success 200 {object} models.SelfServiceRequest
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /api/self-service/requests/{requestID} [get]
func (h *SelfServiceHandler) GetRequest(c echo.Context) error {
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "Invalid request ID", http.StatusBadRequest)
	}

	request, err := h.service.GetStudentRequest(c.Request().Context(), collegeID, studentID, requestID)
	if err != nil {
		if errors.Is(err, selfservice.ErrRequestNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Request not found"}, http.StatusNotFound)
		}
		if errors.Is(err, selfservice.ErrAccessDenied) {
			return helpers.Error(c, "Access denied", http.StatusForbidden)
		}
		return helpers.Error(c, "Failed to fetch request", http.StatusInternalServerError)
	}

	return helpers.Success(c, request, http.StatusOK)
}

// CreateRequest godoc
// @Summary Create a new self-service request
// @Description Creates a new self-service request (enrollment, schedule change, document, etc.)
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param request body models.CreateSelfServiceRequestInput true "Request details"
// @Success 201 {object} models.SelfServiceRequest
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 500 {object} helpers.ErrorResponse
// @Router /api/self-service/requests [post]
func (h *SelfServiceHandler) CreateRequest(c echo.Context) error {
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	var input models.CreateSelfServiceRequestInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	request := &models.SelfServiceRequest{
		StudentID:      studentID,
		CollegeID:      collegeID,
		Type:           input.Type,
		Title:          input.Title,
		Description:    input.Description,
		DocumentType:   input.DocumentType,
		DeliveryMethod: input.DeliveryMethod,
	}

	if err := h.service.CreateStudentRequest(c.Request().Context(), request); err != nil {
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, request, http.StatusCreated)
}

// UpdateRequest godoc
// @Summary Update a self-service request (Admin only)
// @Description Updates the status and response of a request
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param requestID path int true "Request ID"
// @Param request body models.UpdateSelfServiceRequestInput true "Update details"
// @Success 200 {object} models.SelfServiceRequest
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /api/self-service/requests/{requestID} [put]
func (h *SelfServiceHandler) UpdateRequest(c echo.Context) error {
	collegeID, err := helpers.ExtractCollegeID(c)
	if err != nil {
		return err
	}

	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "Invalid request ID", http.StatusBadRequest)
	}

	var input models.UpdateSelfServiceRequestInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	kratosID, err := helpers.GetKratosID(c)
	if err != nil {
		return helpers.Error(c, "Unauthorized", http.StatusUnauthorized)
	}
	responderUserID, err := h.service.ResolveUserIDByKratosID(c.Request().Context(), kratosID)
	if err != nil {
		return helpers.Error(c, "Unable to resolve admin user", http.StatusUnauthorized)
	}

	updated, err := h.service.UpdateRequest(c.Request().Context(), collegeID, requestID, responderUserID, input.Status, input.Response)
	if err != nil {
		if errors.Is(err, selfservice.ErrRequestNotFound) {
			return helpers.NotFound(c, map[string]any{"error": "Request not found"}, http.StatusNotFound)
		}
		return helpers.Error(c, err.Error(), http.StatusBadRequest)
	}

	return helpers.Success(c, updated, http.StatusOK)
}

// GetRequestTypes godoc
// @Summary Get available request types
// @Description Returns a list of available self-service request types
// @Tags Self-Service
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/self-service/types [get]
func (h *SelfServiceHandler) GetRequestTypes(c echo.Context) error {
	return helpers.Success(c, map[string]any{
		"types":           h.service.RequestTypes(),
		"documentTypes":   h.service.DocumentTypes(),
		"deliveryMethods": h.service.DeliveryMethods(),
	}, http.StatusOK)
}
