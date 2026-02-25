package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"eduhub/server/internal/middleware"
	"eduhub/server/internal/models"
	"eduhub/server/internal/services/fee"

	"github.com/labstack/echo/v4"
)

type FeeHandler struct {
	feeService fee.FeeService
}

func NewFeeHandler(feeService fee.FeeService) *FeeHandler {
	return &FeeHandler{
		feeService: feeService,
	}
}

// Fee Structure Management

func (h *FeeHandler) CreateFeeStructure(c echo.Context) error {
	var req models.CreateFeeStructureRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	collegeID := c.Get("college_id").(int)
	feeStructure, err := h.feeService.CreateFeeStructure(c.Request().Context(), &req, collegeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create fee structure: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Fee structure created successfully",
		"data":    feeStructure,
	})
}

func (h *FeeHandler) ListFeeStructures(c echo.Context) error {
	collegeID := c.Get("college_id").(int)

	filter := models.FeeFilter{
		CollegeID: collegeID,
	}

	fees, err := h.feeService.ListFeeStructures(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list fee structures: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": fees,
	})
}

func (h *FeeHandler) UpdateFeeStructure(c echo.Context) error {
	feeID, err := strconv.Atoi(c.Param("feeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid fee ID")
	}

	var req models.UpdateFeeStructureRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	collegeID := c.Get("college_id").(int)

	if err := h.feeService.UpdateFeeStructure(c.Request().Context(), feeID, &req, collegeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update fee structure: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Fee structure updated successfully",
	})
}

func (h *FeeHandler) DeleteFeeStructure(c echo.Context) error {
	feeID, err := strconv.Atoi(c.Param("feeID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid fee ID")
	}

	collegeID := c.Get("college_id").(int)

	if err := h.feeService.DeleteFeeStructure(c.Request().Context(), feeID, collegeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete fee structure: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Fee structure deleted successfully",
	})
}

// Fee Assignment Management

func (h *FeeHandler) AssignFeeToStudent(c echo.Context) error {
	var req models.AssignFeeRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.feeService.AssignFeeToStudent(c.Request().Context(), &req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign fee: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Fee assigned to student successfully",
	})
}

func (h *FeeHandler) BulkAssignFee(c echo.Context) error {
	var req models.BulkAssignFeeRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.feeService.BulkAssignFeeToStudents(c.Request().Context(), &req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to bulk assign fee: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Fee assigned to students successfully",
	})
}

func (h *FeeHandler) GetStudentFees(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	assignments, err := h.feeService.GetStudentFeeAssignments(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get student fees: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": assignments,
	})
}

func (h *FeeHandler) GetStudentFeesSummary(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	summary, err := h.feeService.GetStudentFeesSummary(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get fees summary: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": summary,
	})
}

// Fee Payment Management

func (h *FeeHandler) MakeFeePayment(c echo.Context) error {
	var req models.MakeFeePaymentRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	studentID := c.Get("student_id").(int)
	userID := c.Get("user_id").(int)

	payment, err := h.feeService.MakeFeePayment(c.Request().Context(), &req, studentID, &userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to make payment: "+err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Payment successful",
		"data":    payment,
	})
}

func (h *FeeHandler) InitiateOnlinePayment(c echo.Context) error {
	var req models.InitiateOnlinePaymentRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	studentID := c.Get("student_id").(int)

	response, err := h.feeService.InitiateOnlinePayment(c.Request().Context(), &req, studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate payment: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": response,
	})
}

func (h *FeeHandler) VerifyPayment(c echo.Context) error {
	var req models.ConfirmOnlinePaymentRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return err
	}

	if err := h.feeService.VerifyPayment(c.Request().Context(), &req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify payment: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Payment verified and completed successfully",
	})
}

// HandleWebhook processes Razorpay webhook events with proper signature verification.
// Security: Implements HMAC-SHA256 signature verification to prevent fraudulent webhook calls.
func (h *FeeHandler) HandleWebhook(c echo.Context) error {
	// Read raw body for signature verification (must be done before any parsing)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	// Get the signature from header
	signature := c.Request().Header.Get("X-Razorpay-Signature")
	if signature == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing X-Razorpay-Signature header",
		})
	}

	// Verify the signature
	if !h.feeService.VerifyWebhookSignature(body, signature) {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid webhook signature",
		})
	}

	// Parse the webhook payload
	var payload struct {
		Event   string         `json:"event"`
		Payload map[string]any `json:"payload"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
		})
	}

	// Process the webhook event
	if err := h.feeService.ProcessWebhookEvent(c.Request().Context(), payload.Event, payload.Payload); err != nil {
		// Log the error but return 200 to prevent Razorpay from retrying indefinitely
		// In production, you'd want to log this to a monitoring service
		return c.JSON(http.StatusOK, map[string]string{
			"status": "acknowledged",
			"note":   "Event processing had issues but acknowledged",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "processed",
	})
}

func (h *FeeHandler) GetStudentPayments(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	payments, err := h.feeService.GetStudentPayments(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get payments: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": payments,
	})
}
