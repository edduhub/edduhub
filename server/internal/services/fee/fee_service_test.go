package fee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"eduhub/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFeeRepository struct {
	mock.Mock
}

func (m *mockFeeRepository) CreateFeeStructure(ctx context.Context, fee *models.FeeStructure) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

func (m *mockFeeRepository) GetFeeStructure(ctx context.Context, feeID int, collegeID int) (*models.FeeStructure, error) {
	args := m.Called(ctx, feeID, collegeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FeeStructure), args.Error(1)
}

func (m *mockFeeRepository) UpdateFeeStructure(ctx context.Context, fee *models.FeeStructure) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

func (m *mockFeeRepository) DeleteFeeStructure(ctx context.Context, feeID int, collegeID int) error {
	args := m.Called(ctx, feeID, collegeID)
	return args.Error(0)
}

func (m *mockFeeRepository) ListFeeStructures(ctx context.Context, filter models.FeeFilter) ([]*models.FeeStructure, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FeeStructure), args.Error(1)
}

func (m *mockFeeRepository) AssignFeeToStudent(ctx context.Context, assignment *models.FeeAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *mockFeeRepository) GetStudentFeeAssignments(ctx context.Context, studentID int) ([]*models.FeeAssignment, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FeeAssignment), args.Error(1)
}

func (m *mockFeeRepository) GetStudentFeesSummary(ctx context.Context, studentID int) (*models.StudentFeesSummary, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StudentFeesSummary), args.Error(1)
}

func (m *mockFeeRepository) CreateFeePayment(ctx context.Context, payment *models.FeePayment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *mockFeeRepository) GetFeePayment(ctx context.Context, paymentID int) (*models.FeePayment, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FeePayment), args.Error(1)
}

func (m *mockFeeRepository) UpdatePaymentStatus(ctx context.Context, paymentID int, status string, transactionID *string) error {
	args := m.Called(ctx, paymentID, status, transactionID)
	return args.Error(0)
}

func (m *mockFeeRepository) UpdatePaymentStatusByTransactionID(ctx context.Context, orderID string, status string, paymentID *string) error {
	args := m.Called(ctx, orderID, status, paymentID)
	return args.Error(0)
}

func (m *mockFeeRepository) GetStudentPayments(ctx context.Context, studentID int) ([]*models.FeePayment, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.FeePayment), args.Error(1)
}

func (m *mockFeeRepository) GetFeeAssignment(ctx context.Context, assignmentID int) (*models.FeeAssignment, error) {
	args := m.Called(ctx, assignmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FeeAssignment), args.Error(1)
}

func (m *mockFeeRepository) UpdateFeeAssignmentStatus(ctx context.Context, assignmentID int, status string) error {
	args := m.Called(ctx, assignmentID, status)
	return args.Error(0)
}

func (m *mockFeeRepository) GetTotalPaidAmount(ctx context.Context, assignmentID int) (float64, error) {
	args := m.Called(ctx, assignmentID)
	return args.Get(0).(float64), args.Error(1)
}

func generateValidSignature(orderID, paymentID, secret string) string {
	signatureString := orderID + "|" + paymentID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signatureString))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature_Success(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	body := []byte(`{"event":"payment.captured"}`)

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	result := service.VerifyWebhookSignature(body, validSignature)

	assert.True(t, result)
}

func TestVerifyWebhookSignature_InvalidSignature(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	body := []byte(`{"event":"payment.captured"}`)

	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	result := service.VerifyWebhookSignature(body, "invalid_signature")

	assert.False(t, result)
}

func TestVerifyWebhookSignature_NoSecret(t *testing.T) {
	body := []byte(`{"event":"payment.captured"}`)

	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", "")

	result := service.VerifyWebhookSignature(body, "any_signature")

	assert.False(t, result)
}

func TestVerifyPayment_MissingFields(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	tests := []struct {
		name    string
		req     *models.ConfirmOnlinePaymentRequest
		wantErr string
	}{
		{
			name: "missing order_id",
			req: &models.ConfirmOnlinePaymentRequest{
				PaymentID:     1,
				TransactionID: "pay_123",
				Signature:     "sig_456",
				Gateway:       "razorpay",
			},
			wantErr: "invalid payment verification request",
		},
		{
			name: "missing transaction_id",
			req: &models.ConfirmOnlinePaymentRequest{
				PaymentID: 1,
				OrderID:   "order_123",
				Signature: "sig_456",
				Gateway:   "razorpay",
			},
			wantErr: "invalid payment verification request",
		},
		{
			name: "missing signature",
			req: &models.ConfirmOnlinePaymentRequest{
				PaymentID:     1,
				OrderID:       "order_123",
				TransactionID: "pay_123",
				Gateway:       "razorpay",
			},
			wantErr: "invalid payment verification request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.VerifyPayment(context.Background(), tt.req)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestVerifyPayment_InvalidSignature(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	req := &models.ConfirmOnlinePaymentRequest{
		PaymentID:     1,
		OrderID:       "order_123",
		TransactionID: "pay_123",
		Signature:     "invalid_signature",
		Gateway:       "razorpay",
	}

	err := service.VerifyPayment(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payment signature")
}

func TestVerifyPayment_Success(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	orderID := "order_123"
	paymentID := "pay_123"
	validSignature := generateValidSignature(orderID, paymentID, webhookSecret)

	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	req := &models.ConfirmOnlinePaymentRequest{
		PaymentID:     1,
		OrderID:       orderID,
		TransactionID: paymentID,
		Signature:     validSignature,
		Gateway:       "razorpay",
	}

	mockFeeAssignment := &models.FeeAssignment{
		Amount:       1000.0,
		WaiverAmount: 0,
	}
	mockPayment := &models.FeePayment{
		ID:              1,
		FeeAssignmentID: 1,
	}

	mockRepo.On("UpdatePaymentStatus", mock.Anything, 1, "completed", &paymentID).Return(nil)
	mockRepo.On("GetFeePayment", mock.Anything, 1).Return(mockPayment, nil)
	mockRepo.On("GetFeeAssignment", mock.Anything, 1).Return(mockFeeAssignment, nil)
	mockRepo.On("GetTotalPaidAmount", mock.Anything, 1).Return(1000.0, nil)
	mockRepo.On("UpdateFeeAssignmentStatus", mock.Anything, 1, "paid").Return(nil)

	err := service.VerifyPayment(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestVerifyPayment_PartialPayment(t *testing.T) {
	webhookSecret := "test_secret_key_123456789"
	orderID := "order_123"
	paymentID := "pay_123"
	validSignature := generateValidSignature(orderID, paymentID, webhookSecret)

	mockRepo := new(mockFeeRepository)
	service := NewFeeService(mockRepo, "test_key", "test_secret", webhookSecret)

	req := &models.ConfirmOnlinePaymentRequest{
		PaymentID:     1,
		OrderID:       orderID,
		TransactionID: paymentID,
		Signature:     validSignature,
		Gateway:       "razorpay",
	}

	mockFeeAssignment := &models.FeeAssignment{
		Amount:       1000.0,
		WaiverAmount: 0,
	}
	mockPayment := &models.FeePayment{
		ID:              1,
		FeeAssignmentID: 1,
	}

	mockRepo.On("UpdatePaymentStatus", mock.Anything, 1, "completed", &paymentID).Return(nil)
	mockRepo.On("GetFeePayment", mock.Anything, 1).Return(mockPayment, nil)
	mockRepo.On("GetFeeAssignment", mock.Anything, 1).Return(mockFeeAssignment, nil)
	mockRepo.On("GetTotalPaidAmount", mock.Anything, 1).Return(500.0, nil)
	mockRepo.On("UpdateFeeAssignmentStatus", mock.Anything, 1, "partial").Return(nil)

	err := service.VerifyPayment(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
