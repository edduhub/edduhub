package fee

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type FeeService interface {
	// Fee Structure management
	CreateFeeStructure(ctx context.Context, req *models.CreateFeeStructureRequest, collegeID int) (*models.FeeStructure, error)
	GetFeeStructure(ctx context.Context, feeID int, collegeID int) (*models.FeeStructure, error)
	UpdateFeeStructure(ctx context.Context, feeID int, req *models.UpdateFeeStructureRequest, collegeID int) error
	DeleteFeeStructure(ctx context.Context, feeID int, collegeID int) error
	ListFeeStructures(ctx context.Context, filter models.FeeFilter) ([]*models.FeeStructure, error)

	// Fee Assignment management
	AssignFeeToStudent(ctx context.Context, req *models.AssignFeeRequest) error
	BulkAssignFeeToStudents(ctx context.Context, req *models.BulkAssignFeeRequest) error
	GetStudentFeeAssignments(ctx context.Context, studentID int) ([]*models.FeeAssignment, error)
	GetStudentFeesSummary(ctx context.Context, studentID int) (*models.StudentFeesSummary, error)

	// Fee Payment management
	MakeFeePayment(ctx context.Context, req *models.MakeFeePaymentRequest, studentID int, processedBy *int) (*models.FeePayment, error)
	GetStudentPayments(ctx context.Context, studentID int) ([]*models.FeePayment, error)
	GetPaymentReceipt(ctx context.Context, paymentID int) (*models.FeePayment, error)

	// Online payment flow
	InitiateOnlinePayment(ctx context.Context, req *models.InitiateOnlinePaymentRequest, studentID int) (*models.OnlinePaymentResponse, error)
	ConfirmOnlinePayment(ctx context.Context, req *models.ConfirmOnlinePaymentRequest) error
}

type feeService struct {
	feeRepo repository.FeeRepository
}

func NewFeeService(feeRepo repository.FeeRepository) FeeService {
	return &feeService{
		feeRepo: feeRepo,
	}
}

func (s *feeService) CreateFeeStructure(ctx context.Context, req *models.CreateFeeStructureRequest, collegeID int) (*models.FeeStructure, error) {
	fee := &models.FeeStructure{
		CollegeID:    collegeID,
		Name:         req.Name,
		Description:  req.Description,
		Amount:       req.Amount,
		Currency:     req.Currency,
		FeeType:      req.FeeType,
		Frequency:    req.Frequency,
		AcademicYear: req.AcademicYear,
		Semester:     req.Semester,
		DepartmentID: req.DepartmentID,
		CourseID:     req.CourseID,
		IsMandatory:  req.IsMandatory,
		DueDate:      req.DueDate,
	}

	if fee.Currency == "" {
		fee.Currency = "USD"
	}

	if err := s.feeRepo.CreateFeeStructure(ctx, fee); err != nil {
		return nil, fmt.Errorf("failed to create fee structure: %w", err)
	}

	return fee, nil
}

func (s *feeService) GetFeeStructure(ctx context.Context, feeID int, collegeID int) (*models.FeeStructure, error) {
	return s.feeRepo.GetFeeStructure(ctx, feeID, collegeID)
}

func (s *feeService) UpdateFeeStructure(ctx context.Context, feeID int, req *models.UpdateFeeStructureRequest, collegeID int) error {
	fee, err := s.feeRepo.GetFeeStructure(ctx, feeID, collegeID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		fee.Name = *req.Name
	}
	if req.Description != nil {
		fee.Description = req.Description
	}
	if req.Amount != nil {
		fee.Amount = *req.Amount
	}
	if req.FeeType != nil {
		fee.FeeType = *req.FeeType
	}
	if req.Frequency != nil {
		fee.Frequency = *req.Frequency
	}
	if req.AcademicYear != nil {
		fee.AcademicYear = req.AcademicYear
	}
	if req.Semester != nil {
		fee.Semester = req.Semester
	}
	if req.IsMandatory != nil {
		fee.IsMandatory = *req.IsMandatory
	}
	if req.DueDate != nil {
		fee.DueDate = req.DueDate
	}

	return s.feeRepo.UpdateFeeStructure(ctx, fee)
}

func (s *feeService) DeleteFeeStructure(ctx context.Context, feeID int, collegeID int) error {
	return s.feeRepo.DeleteFeeStructure(ctx, feeID, collegeID)
}

func (s *feeService) ListFeeStructures(ctx context.Context, filter models.FeeFilter) ([]*models.FeeStructure, error) {
	return s.feeRepo.ListFeeStructures(ctx, filter)
}

func (s *feeService) AssignFeeToStudent(ctx context.Context, req *models.AssignFeeRequest) error {
	// Get fee structure to determine default amount
	fee, err := s.feeRepo.GetFeeStructure(ctx, req.FeeStructureID, 0) // College ID will be checked by constraint
	if err != nil {
		return fmt.Errorf("fee structure not found: %w", err)
	}

	amount := fee.Amount
	if req.Amount != nil {
		amount = *req.Amount
	}

	waiverAmount := 0.0
	if req.WaiverAmount != nil {
		waiverAmount = *req.WaiverAmount
	}

	dueDate := fee.DueDate
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	assignment := &models.FeeAssignment{
		StudentID:      req.StudentID,
		FeeStructureID: req.FeeStructureID,
		Amount:         amount,
		WaiverAmount:   waiverAmount,
		WaiverReason:   req.WaiverReason,
		DueDate:        dueDate,
	}

	return s.feeRepo.AssignFeeToStudent(ctx, assignment)
}

func (s *feeService) BulkAssignFeeToStudents(ctx context.Context, req *models.BulkAssignFeeRequest) error {
	// Get fee structure
	fee, err := s.feeRepo.GetFeeStructure(ctx, req.FeeStructureID, 0)
	if err != nil {
		return fmt.Errorf("fee structure not found: %w", err)
	}

	dueDate := fee.DueDate
	if req.DueDate != nil {
		dueDate = req.DueDate
	}

	// Assign fee to each student
	for _, studentID := range req.StudentIDs {
		assignment := &models.FeeAssignment{
			StudentID:      studentID,
			FeeStructureID: req.FeeStructureID,
			Amount:         fee.Amount,
			DueDate:        dueDate,
		}

		if err := s.feeRepo.AssignFeeToStudent(ctx, assignment); err != nil {
			return fmt.Errorf("failed to assign fee to student %d: %w", studentID, err)
		}
	}

	return nil
}

func (s *feeService) GetStudentFeeAssignments(ctx context.Context, studentID int) ([]*models.FeeAssignment, error) {
	return s.feeRepo.GetStudentFeeAssignments(ctx, studentID)
}

func (s *feeService) GetStudentFeesSummary(ctx context.Context, studentID int) (*models.StudentFeesSummary, error) {
	return s.feeRepo.GetStudentFeesSummary(ctx, studentID)
}

func (s *feeService) MakeFeePayment(ctx context.Context, req *models.MakeFeePaymentRequest, studentID int, processedBy *int) (*models.FeePayment, error) {
	// Get assignment to verify it exists
	assignment, err := s.feeRepo.GetFeeAssignment(ctx, req.FeeAssignmentID)
	if err != nil {
		return nil, fmt.Errorf("fee assignment not found: %w", err)
	}

	// Verify student owns this assignment
	if assignment.StudentID != studentID {
		return nil, fmt.Errorf("fee assignment does not belong to this student")
	}

	// Generate receipt number
	receiptNumber := fmt.Sprintf("RCP-%d-%d-%d", studentID, req.FeeAssignmentID, time.Now().Unix())

	payment := &models.FeePayment{
		FeeAssignmentID: req.FeeAssignmentID,
		StudentID:       studentID,
		Amount:          req.Amount,
		Currency:        "USD",
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   "completed",
		ReceiptNumber:   &receiptNumber,
		Notes:           req.Notes,
		ProcessedBy:     processedBy,
	}

	if err := s.feeRepo.CreateFeePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Update assignment status
	paidAmount, _ := s.feeRepo.GetTotalPaidAmount(ctx, req.FeeAssignmentID)
	remainingAmount := assignment.Amount - paidAmount - assignment.WaiverAmount

	if remainingAmount <= 0 {
		s.feeRepo.UpdateFeeAssignmentStatus(ctx, req.FeeAssignmentID, "paid")
	} else {
		s.feeRepo.UpdateFeeAssignmentStatus(ctx, req.FeeAssignmentID, "partial")
	}

	return payment, nil
}

func (s *feeService) GetStudentPayments(ctx context.Context, studentID int) ([]*models.FeePayment, error) {
	return s.feeRepo.GetStudentPayments(ctx, studentID)
}

func (s *feeService) GetPaymentReceipt(ctx context.Context, paymentID int) (*models.FeePayment, error) {
	return s.feeRepo.GetFeePayment(ctx, paymentID)
}

func (s *feeService) InitiateOnlinePayment(ctx context.Context, req *models.InitiateOnlinePaymentRequest, studentID int) (*models.OnlinePaymentResponse, error) {
	// Get assignment
	assignment, err := s.feeRepo.GetFeeAssignment(ctx, req.FeeAssignmentID)
	if err != nil {
		return nil, fmt.Errorf("fee assignment not found: %w", err)
	}

	if assignment.StudentID != studentID {
		return nil, fmt.Errorf("fee assignment does not belong to this student")
	}

	// Create pending payment record
	transactionID := fmt.Sprintf("TXN-%d-%d-%d", studentID, req.FeeAssignmentID, time.Now().Unix())

	payment := &models.FeePayment{
		FeeAssignmentID: req.FeeAssignmentID,
		StudentID:       studentID,
		Amount:          req.Amount,
		Currency:        "USD",
		PaymentMethod:   "online",
		PaymentStatus:   "pending",
		TransactionID:   &transactionID,
		Gateway:         &req.Gateway,
	}

	if err := s.feeRepo.CreateFeePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// In a real implementation, this would integrate with actual payment gateway
	// For now, return a mock response
	checkoutURL := fmt.Sprintf("https://payment-gateway.example.com/checkout/%s", transactionID)

	return &models.OnlinePaymentResponse{
		PaymentID:     payment.ID,
		CheckoutURL:   checkoutURL,
		TransactionID: transactionID,
		Status:        "pending",
	}, nil
}

func (s *feeService) ConfirmOnlinePayment(ctx context.Context, req *models.ConfirmOnlinePaymentRequest) error {
	// In a real implementation, verify payment with gateway
	// For now, just update status

	if err := s.feeRepo.UpdatePaymentStatus(ctx, req.PaymentID, "completed", &req.TransactionID); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Get payment to update assignment status
	payment, err := s.feeRepo.GetFeePayment(ctx, req.PaymentID)
	if err != nil {
		return err
	}

	assignment, err := s.feeRepo.GetFeeAssignment(ctx, payment.FeeAssignmentID)
	if err != nil {
		return err
	}

	paidAmount, _ := s.feeRepo.GetTotalPaidAmount(ctx, payment.FeeAssignmentID)
	remainingAmount := assignment.Amount - paidAmount - assignment.WaiverAmount

	if remainingAmount <= 0 {
		s.feeRepo.UpdateFeeAssignmentStatus(ctx, payment.FeeAssignmentID, "paid")
	} else {
		s.feeRepo.UpdateFeeAssignmentStatus(ctx, payment.FeeAssignmentID, "partial")
	}

	return nil
}
