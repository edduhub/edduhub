package fee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"

	"github.com/razorpay/razorpay-go"
	"github.com/rs/zerolog/log"
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
	VerifyPayment(ctx context.Context, req *models.ConfirmOnlinePaymentRequest) error

	// Webhook processing
	VerifyWebhookSignature(body []byte, signature string) bool
	ProcessWebhookEvent(ctx context.Context, eventType string, payload map[string]interface{}) error
}

type feeService struct {
	feeRepo       repository.FeeRepository
	rzp           *razorpay.Client
	webhookSecret string
}

func NewFeeService(feeRepo repository.FeeRepository, rzpKey, rzpSecret, webhookSecret string) FeeService {
	client := razorpay.NewClient(rzpKey, rzpSecret)

	if webhookSecret == "" {
		log.Warn().Msg("RAZORPAY_WEBHOOK_SECRET is not set - payment and webhook signature verification will fail")
	}

	return &feeService{
		feeRepo:       feeRepo,
		rzp:           client,
		webhookSecret: webhookSecret,
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
		// Check for existing assignment to prevent duplicates
		existing, _ := s.feeRepo.GetStudentFeeAssignments(ctx, studentID)
		isDuplicate := false
		for _, e := range existing {
			if e.FeeStructureID == req.FeeStructureID {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue // Skip if already assigned
		}

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

	// For Razorpay, we create an order
	amountInPaise := int(req.Amount * 100)
	orderData := map[string]interface{}{
		"amount":          amountInPaise,
		"currency":        "INR", // Razorpay usually expects INR
		"receipt":         fmt.Sprintf("rcpt_%d_%d", studentID, time.Now().Unix()),
		"payment_capture": 1,
	}

	body, err := s.rzp.Order.Create(orderData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
	}

	razorpayOrderID := body["id"].(string)

	// Create pending payment record
	payment := &models.FeePayment{
		FeeAssignmentID: req.FeeAssignmentID,
		StudentID:       studentID,
		Amount:          req.Amount,
		Currency:        "INR",
		PaymentMethod:   "online",
		PaymentStatus:   "pending",
		TransactionID:   &razorpayOrderID, // Store order ID as transaction ID initially
		Gateway:         &req.Gateway,
	}

	if err := s.feeRepo.CreateFeePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return &models.OnlinePaymentResponse{
		PaymentID:     payment.ID,
		TransactionID: razorpayOrderID,
		Status:        "pending",
	}, nil
}

func (s *feeService) VerifyPayment(ctx context.Context, req *models.ConfirmOnlinePaymentRequest) error {
	if req.OrderID == "" || req.TransactionID == "" || req.Signature == "" {
		return fmt.Errorf("invalid payment verification request: missing order_id, payment_id, or signature")
	}

	signatureString := req.OrderID + "|" + req.TransactionID

	signature := hmac.New(sha256.New, []byte(s.webhookSecret))
	signature.Write([]byte(signatureString))
	digest := hex.EncodeToString(signature.Sum(nil))

	if !hmac.Equal([]byte(digest), []byte(req.Signature)) {
		log.Error().
			Str("order_id", req.OrderID).
			Str("payment_id", req.TransactionID).
			Msg("Payment signature verification failed")
		return fmt.Errorf("invalid payment signature: verification failed")
	}

	log.Info().
		Str("order_id", req.OrderID).
		Str("payment_id", req.TransactionID).
		Msg("Payment signature verified successfully")

	if err := s.feeRepo.UpdatePaymentStatus(ctx, req.PaymentID, "completed", &req.TransactionID); err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

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

func (s *feeService) VerifyWebhookSignature(body []byte, signature string) bool {
	if s.webhookSecret == "" {
		log.Error().Msg("Webhook signature verification failed: RAZORPAY_WEBHOOK_SECRET is not configured")
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	isValid := hmac.Equal([]byte(expectedSignature), []byte(signature))
	if !isValid {
		log.Error().Msg("Webhook signature verification failed: invalid signature")
	} else {
		log.Debug().Msg("Webhook signature verified successfully")
	}

	return isValid
}

// ProcessWebhookEvent processes Razorpay webhook events (payment.captured, payment.failed, etc.)
func (s *feeService) ProcessWebhookEvent(ctx context.Context, eventType string, payload map[string]interface{}) error {
	switch eventType {
	case "payment.captured":
		return s.handlePaymentCaptured(ctx, payload)
	case "payment.failed":
		return s.handlePaymentFailed(ctx, payload)
	case "order.paid":
		return s.handleOrderPaid(ctx, payload)
	default:
		// Unhandled event type - log and ignore
		return nil
	}
}

// handlePaymentCaptured processes successful payment webhooks
func (s *feeService) handlePaymentCaptured(ctx context.Context, payload map[string]interface{}) error {
	paymentData, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment payload structure")
	}

	entity, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment entity structure")
	}

	orderID, _ := entity["order_id"].(string)
	paymentID, _ := entity["id"].(string)

	if orderID == "" || paymentID == "" {
		return fmt.Errorf("missing order_id or payment_id in webhook payload")
	}

	// Update payment status in database using order_id (which is our transaction_id)
	return s.feeRepo.UpdatePaymentStatusByTransactionID(ctx, orderID, "completed", &paymentID)
}

// handlePaymentFailed processes failed payment webhooks
func (s *feeService) handlePaymentFailed(ctx context.Context, payload map[string]interface{}) error {
	paymentData, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment payload structure")
	}

	entity, ok := paymentData["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment entity structure")
	}

	orderID, _ := entity["order_id"].(string)
	if orderID == "" {
		return fmt.Errorf("missing order_id in webhook payload")
	}

	// Update payment status to failed
	return s.feeRepo.UpdatePaymentStatusByTransactionID(ctx, orderID, "failed", nil)
}

// handleOrderPaid processes order.paid webhooks (alternative to payment.captured)
func (s *feeService) handleOrderPaid(ctx context.Context, payload map[string]interface{}) error {
	orderData, ok := payload["order"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid order payload structure")
	}

	entity, ok := orderData["entity"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid order entity structure")
	}

	orderID, _ := entity["id"].(string)
	if orderID == "" {
		return fmt.Errorf("missing order_id in webhook payload")
	}

	// Update payment status to completed
	return s.feeRepo.UpdatePaymentStatusByTransactionID(ctx, orderID, "completed", nil)
}
