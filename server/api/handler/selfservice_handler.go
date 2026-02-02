package handler

import (
	"net/http"
	"strconv"
	"time"

	"eduhub/server/internal/helpers"

	"github.com/labstack/echo/v4"
)

// SelfServiceRequest represents a student self-service request
type SelfServiceRequest struct {
	ID             int        `json:"id" db:"id"`
	StudentID      int        `json:"student_id" db:"student_id"`
	CollegeID      int        `json:"college_id" db:"college_id"`
	Type           string     `json:"type" db:"type"` // enrollment, schedule, transcript, document
	Title          string     `json:"title" db:"title"`
	Description    string     `json:"description" db:"description"`
	Status         string     `json:"status" db:"status"` // pending, approved, rejected, processing
	SubmittedAt    time.Time  `json:"submitted_at" db:"submitted_at"`
	RespondedAt    *time.Time `json:"responded_at,omitempty" db:"responded_at"`
	Response       *string    `json:"response,omitempty" db:"response"`
	DocumentType   *string    `json:"document_type,omitempty" db:"document_type"`
	DeliveryMethod *string    `json:"delivery_method,omitempty" db:"delivery_method"`
}

// CreateSelfServiceRequestInput represents input for creating a request
type CreateSelfServiceRequestInput struct {
	Type           string  `json:"type" validate:"required,oneof=enrollment schedule transcript document"`
	Title          string  `json:"title" validate:"required,max=200"`
	Description    string  `json:"description" validate:"required,max=1000"`
	DocumentType   *string `json:"document_type,omitempty"`
	DeliveryMethod *string `json:"delivery_method,omitempty"`
}

// UpdateSelfServiceRequestInput represents input for updating a request
type UpdateSelfServiceRequestInput struct {
	Status   string `json:"status" validate:"required,oneof=approved rejected processing"`
	Response string `json:"response" validate:"required,max=1000"`
}

// SelfServiceHandler handles student self-service requests
type SelfServiceHandler struct {
	// In-memory store for demonstration
	requests []SelfServiceRequest
}

// NewSelfServiceHandler creates a new SelfServiceHandler
func NewSelfServiceHandler() *SelfServiceHandler {
	return &SelfServiceHandler{
		requests: []SelfServiceRequest{},
	}
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

	// Filter requests for this student
	var studentRequests []SelfServiceRequest
	for _, req := range h.requests {
		if req.StudentID == studentID {
			studentRequests = append(studentRequests, req)
		}
	}

	return helpers.Success(c, map[string]interface{}{
		"requests": studentRequests,
		"total":    len(studentRequests),
	}, http.StatusOK)
}

// GetRequest godoc
// @Summary Get a specific self-service request
// @Description Returns details of a specific request
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param requestID path int true "Request ID"
// @Success 200 {object} SelfServiceRequest
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /api/self-service/requests/{requestID} [get]
func (h *SelfServiceHandler) GetRequest(c echo.Context) error {
	studentID, err := helpers.ExtractStudentID(c)
	if err != nil {
		return err
	}

	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "Invalid request ID", http.StatusBadRequest)
	}

	// Find the request
	for _, req := range h.requests {
		if req.ID == requestID {
			// Verify ownership
			if req.StudentID != studentID {
				return helpers.Error(c, "Access denied", http.StatusForbidden)
			}
			return helpers.Success(c, req, http.StatusOK)
		}
	}

	return helpers.NotFound(c, map[string]interface{}{"error": "Request not found"}, http.StatusNotFound)
}

// CreateRequest godoc
// @Summary Create a new self-service request
// @Description Creates a new self-service request (enrollment, schedule change, document, etc.)
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param request body CreateSelfServiceRequestInput true "Request details"
// @Success 201 {object} SelfServiceRequest
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

	var input CreateSelfServiceRequestInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	// Validate input
	if input.Type == "" || input.Title == "" || input.Description == "" {
		return helpers.Error(c, "Type, title, and description are required", http.StatusBadRequest)
	}

	// Create new request
	newRequest := SelfServiceRequest{
		ID:             len(h.requests) + 1,
		StudentID:      studentID,
		CollegeID:      collegeID,
		Type:           input.Type,
		Title:          input.Title,
		Description:    input.Description,
		Status:         "pending",
		SubmittedAt:    time.Now(),
		DocumentType:   input.DocumentType,
		DeliveryMethod: input.DeliveryMethod,
	}

	h.requests = append(h.requests, newRequest)

	return helpers.Success(c, newRequest, http.StatusCreated)
}

// UpdateRequest godoc
// @Summary Update a self-service request (Admin only)
// @Description Updates the status and response of a request
// @Tags Self-Service
// @Accept json
// @Produce json
// @Param requestID path int true "Request ID"
// @Param request body UpdateSelfServiceRequestInput true "Update details"
// @Success 200 {object} SelfServiceRequest
// @Failure 400 {object} helpers.ErrorResponse
// @Failure 401 {object} helpers.ErrorResponse
// @Failure 403 {object} helpers.ErrorResponse
// @Failure 404 {object} helpers.ErrorResponse
// @Router /api/self-service/requests/{requestID} [put]
func (h *SelfServiceHandler) UpdateRequest(c echo.Context) error {
	requestID, err := strconv.Atoi(c.Param("requestID"))
	if err != nil {
		return helpers.Error(c, "Invalid request ID", http.StatusBadRequest)
	}

	var input UpdateSelfServiceRequestInput
	if err := c.Bind(&input); err != nil {
		return helpers.Error(c, "Invalid request body", http.StatusBadRequest)
	}

	if input.Status == "" || input.Response == "" {
		return helpers.Error(c, "Status and response are required", http.StatusBadRequest)
	}

	// Find and update the request
	for i, req := range h.requests {
		if req.ID == requestID {
			now := time.Now()
			h.requests[i].Status = input.Status
			h.requests[i].Response = &input.Response
			h.requests[i].RespondedAt = &now
			return helpers.Success(c, h.requests[i], http.StatusOK)
		}
	}

	return helpers.NotFound(c, map[string]interface{}{"error": "Request not found"}, http.StatusNotFound)
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
	return helpers.Success(c, map[string]interface{}{
		"types": []map[string]interface{}{
			{
				"id":          "enrollment",
				"name":        "Course Enrollment",
				"description": "Request to enroll in a course",
			},
			{
				"id":          "schedule",
				"name":        "Schedule Change",
				"description": "Request to change your class schedule",
			},
			{
				"id":          "transcript",
				"name":        "Official Transcript",
				"description": "Request official academic transcripts",
			},
			{
				"id":          "document",
				"name":        "Document Request",
				"description": "Request other academic documents",
			},
		},
		"documentTypes": []map[string]interface{}{
			{"id": "transcript", "name": "Official Transcript"},
			{"id": "certificate", "name": "Enrollment Certificate"},
			{"id": "id_card", "name": "ID Card Replacement"},
			{"id": "other", "name": "Other Document"},
		},
		"deliveryMethods": []map[string]interface{}{
			{"id": "pickup", "name": "In-Person Pickup"},
			{"id": "email", "name": "Email (Digital Copy)"},
			{"id": "postal", "name": "Postal Mail"},
		},
	}, http.StatusOK)
}
