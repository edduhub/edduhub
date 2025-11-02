package handler

import (
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

	return c.JSON(http.StatusCreated, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Fee assigned to students successfully",
	})
}

func (h *FeeHandler) GetStudentFees(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	assignments, err := h.feeService.GetStudentFeeAssignments(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get student fees: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": assignments,
	})
}

func (h *FeeHandler) GetStudentFeesSummary(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	summary, err := h.feeService.GetStudentFeesSummary(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get fees summary: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
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

	return c.JSON(http.StatusCreated, map[string]interface{}{
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

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

func (h *FeeHandler) GetStudentPayments(c echo.Context) error {
	studentID := c.Get("student_id").(int)

	payments, err := h.feeService.GetStudentPayments(c.Request().Context(), studentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get payments: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": payments,
	})
}
