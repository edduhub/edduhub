package models

import (
	"time"
)

// FeeStructure represents a fee structure defined by the college
type FeeStructure struct {
	ID           int        `db:"id" json:"id"`
	CollegeID    int        `db:"college_id" json:"college_id"`
	Name         string     `db:"name" json:"name"`
	Description  *string    `db:"description" json:"description,omitempty"`
	Amount       float64    `db:"amount" json:"amount"`
	Currency     string     `db:"currency" json:"currency"`
	FeeType      string     `db:"fee_type" json:"fee_type"`
	Frequency    string     `db:"frequency" json:"frequency"`
	AcademicYear *string    `db:"academic_year" json:"academic_year,omitempty"`
	Semester     *string    `db:"semester" json:"semester,omitempty"`
	DepartmentID *int       `db:"department_id" json:"department_id,omitempty"`
	CourseID     *int       `db:"course_id" json:"course_id,omitempty"`
	IsMandatory  bool       `db:"is_mandatory" json:"is_mandatory"`
	DueDate      *time.Time `db:"due_date" json:"due_date,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// FeeAssignment represents a fee assigned to a student
type FeeAssignment struct {
	ID             int        `db:"id" json:"id"`
	StudentID      int        `db:"student_id" json:"student_id"`
	FeeStructureID int        `db:"fee_structure_id" json:"fee_structure_id"`
	Amount         float64    `db:"amount" json:"amount"`
	WaiverAmount   float64    `db:"waiver_amount" json:"waiver_amount"`
	WaiverReason   *string    `db:"waiver_reason" json:"waiver_reason,omitempty"`
	DueDate        *time.Time `db:"due_date" json:"due_date,omitempty"`
	Status         string     `db:"status" json:"status"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`

	// Relations - not stored in DB
	FeeStructure    *FeeStructure `db:"-" json:"fee_structure,omitempty"`
	PaidAmount      float64       `db:"-" json:"paid_amount,omitempty"`
	RemainingAmount float64       `db:"-" json:"remaining_amount,omitempty"`
}

// FeePayment represents a payment made by a student
type FeePayment struct {
	ID              int        `db:"id" json:"id"`
	FeeAssignmentID int        `db:"fee_assignment_id" json:"fee_assignment_id"`
	StudentID       int        `db:"student_id" json:"student_id"`
	Amount          float64    `db:"amount" json:"amount"`
	Currency        string     `db:"currency" json:"currency"`
	PaymentMethod   string     `db:"payment_method" json:"payment_method"`
	PaymentStatus   string     `db:"payment_status" json:"payment_status"`
	TransactionID   *string    `db:"transaction_id" json:"transaction_id,omitempty"`
	Gateway         *string    `db:"gateway" json:"gateway,omitempty"`
	GatewayResponse *string    `db:"gateway_response" json:"gateway_response,omitempty"`
	ReceiptNumber   *string    `db:"receipt_number" json:"receipt_number,omitempty"`
	PaymentDate     *time.Time `db:"payment_date" json:"payment_date,omitempty"`
	Notes           *string    `db:"notes" json:"notes,omitempty"`
	ProcessedBy     *int       `db:"processed_by" json:"processed_by,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// FeePaymentReminder represents a reminder for fee payment
type FeePaymentReminder struct {
	ID              int        `db:"id" json:"id"`
	FeeAssignmentID int        `db:"fee_assignment_id" json:"fee_assignment_id"`
	StudentID       int        `db:"student_id" json:"student_id"`
	ReminderDate    time.Time  `db:"reminder_date" json:"reminder_date"`
	ReminderType    string     `db:"reminder_type" json:"reminder_type"`
	Message         *string    `db:"message" json:"message,omitempty"`
	Sent            bool       `db:"sent" json:"sent"`
	SentAt          *time.Time `db:"sent_at" json:"sent_at,omitempty"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}

// Request/Response types

type CreateFeeStructureRequest struct {
	Name         string     `json:"name" validate:"required,minlen=2,maxlen=200"`
	Description  *string    `json:"description" validate:"omitempty,maxlen=1000"`
	Amount       float64    `json:"amount" validate:"required,gt=0"`
	Currency     string     `json:"currency" validate:"omitempty,len=3"`
	FeeType      string     `json:"fee_type" validate:"required,oneof=tuition hostel exam library misc"`
	Frequency    string     `json:"frequency" validate:"required,oneof=semester annual monthly one-time"`
	AcademicYear *string    `json:"academic_year" validate:"omitempty,maxlen=20"`
	Semester     *string    `json:"semester" validate:"omitempty,maxlen=20"`
	DepartmentID *int       `json:"department_id" validate:"omitempty"`
	CourseID     *int       `json:"course_id" validate:"omitempty"`
	IsMandatory  bool       `json:"is_mandatory"`
	DueDate      *time.Time `json:"due_date" validate:"omitempty"`
}

type UpdateFeeStructureRequest struct {
	Name         *string    `json:"name" validate:"omitempty,minlen=2,maxlen=200"`
	Description  *string    `json:"description" validate:"omitempty,maxlen=1000"`
	Amount       *float64   `json:"amount" validate:"omitempty,gt=0"`
	FeeType      *string    `json:"fee_type" validate:"omitempty,oneof=tuition hostel exam library misc"`
	Frequency    *string    `json:"frequency" validate:"omitempty,oneof=semester annual monthly one-time"`
	AcademicYear *string    `json:"academic_year" validate:"omitempty,maxlen=20"`
	Semester     *string    `json:"semester" validate:"omitempty,maxlen=20"`
	IsMandatory  *bool      `json:"is_mandatory" validate:"omitempty"`
	DueDate      *time.Time `json:"due_date" validate:"omitempty"`
}

type AssignFeeRequest struct {
	StudentID      int        `json:"student_id" validate:"required"`
	FeeStructureID int        `json:"fee_structure_id" validate:"required"`
	Amount         *float64   `json:"amount" validate:"omitempty,gt=0"` // Override amount if needed
	WaiverAmount   *float64   `json:"waiver_amount" validate:"omitempty,gte=0"`
	WaiverReason   *string    `json:"waiver_reason" validate:"omitempty,maxlen=500"`
	DueDate        *time.Time `json:"due_date" validate:"omitempty"`
}

type BulkAssignFeeRequest struct {
	StudentIDs     []int      `json:"student_ids" validate:"required,minlen=1"`
	FeeStructureID int        `json:"fee_structure_id" validate:"required"`
	DueDate        *time.Time `json:"due_date" validate:"omitempty"`
}

type MakeFeePaymentRequest struct {
	FeeAssignmentID int     `json:"fee_assignment_id" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	PaymentMethod   string  `json:"payment_method" validate:"required,oneof=card bank_transfer cash cheque online"`
	Notes           *string `json:"notes" validate:"omitempty,maxlen=1000"`
}

type InitiateOnlinePaymentRequest struct {
	FeeAssignmentID int     `json:"fee_assignment_id" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	Gateway         string  `json:"gateway" validate:"required,oneof=stripe paypal razorpay"`
}

type OnlinePaymentResponse struct {
	PaymentID     int    `json:"payment_id"`
	CheckoutURL   string `json:"checkout_url,omitempty"`
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
}

type ConfirmOnlinePaymentRequest struct {
	PaymentID     int    `json:"payment_id" validate:"required"`
	OrderID       string `json:"order_id" validate:"required"`
	TransactionID string `json:"razorpay_payment_id" validate:"required"`
	Signature     string `json:"razorpay_signature" validate:"required"`
	Gateway       string `json:"gateway" validate:"required"`
}

type FeeFilter struct {
	CollegeID    int
	DepartmentID *int
	CourseID     *int
	FeeType      *string
	Frequency    *string
	AcademicYear *string
	Semester     *string
	Limit        int
	Offset       int
}

type FeeAssignmentFilter struct {
	StudentID      *int
	FeeStructureID *int
	Status         *string
	Overdue        *bool
	Limit          int
	Offset         int
}

type FeePaymentFilter struct {
	StudentID       *int
	FeeAssignmentID *int
	PaymentStatus   *string
	PaymentMethod   *string
	DateFrom        *time.Time
	DateTo          *time.Time
	Limit           int
	Offset          int
}

// StudentFeesSummary represents a summary of a student's fees
type StudentFeesSummary struct {
	TotalFees          float64 `json:"total_fees"`
	PaidAmount         float64 `json:"paid_amount"`
	PendingAmount      float64 `json:"pending_amount"`
	OverdueAmount      float64 `json:"overdue_amount"`
	WaiverAmount       float64 `json:"waiver_amount"`
	TotalAssignments   int     `json:"total_assignments"`
	PaidAssignments    int     `json:"paid_assignments"`
	PendingAssignments int     `json:"pending_assignments"`
	OverdueAssignments int     `json:"overdue_assignments"`
}
