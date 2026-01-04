package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type FeeRepository interface {
	// Fee Structure operations
	CreateFeeStructure(ctx context.Context, fee *models.FeeStructure) error
	GetFeeStructure(ctx context.Context, feeID int, collegeID int) (*models.FeeStructure, error)
	UpdateFeeStructure(ctx context.Context, fee *models.FeeStructure) error
	DeleteFeeStructure(ctx context.Context, feeID int, collegeID int) error
	ListFeeStructures(ctx context.Context, filter models.FeeFilter) ([]*models.FeeStructure, error)

	// Fee Assignment operations
	AssignFeeToStudent(ctx context.Context, assignment *models.FeeAssignment) error
	GetFeeAssignment(ctx context.Context, assignmentID int) (*models.FeeAssignment, error)
	GetStudentFeeAssignments(ctx context.Context, studentID int) ([]*models.FeeAssignment, error)
	UpdateFeeAssignmentStatus(ctx context.Context, assignmentID int, status string) error

	// Fee Payment operations
	CreateFeePayment(ctx context.Context, payment *models.FeePayment) error
	GetFeePayment(ctx context.Context, paymentID int) (*models.FeePayment, error)
	GetStudentPayments(ctx context.Context, studentID int) ([]*models.FeePayment, error)
	UpdatePaymentStatus(ctx context.Context, paymentID int, status string, transactionID *string) error
	UpdatePaymentStatusByTransactionID(ctx context.Context, orderID string, status string, paymentID *string) error

	// Summary operations
	GetStudentFeesSummary(ctx context.Context, studentID int) (*models.StudentFeesSummary, error)
	GetTotalPaidAmount(ctx context.Context, assignmentID int) (float64, error)
}

type feeRepository struct {
	DB *DB
}

func NewFeeRepository(db *DB) FeeRepository {
	return &feeRepository{DB: db}
}

func (r *feeRepository) CreateFeeStructure(ctx context.Context, fee *models.FeeStructure) error {
	now := time.Now()
	fee.CreatedAt = now
	fee.UpdatedAt = now

	sql := `INSERT INTO fee_structures (college_id, name, description, amount, currency, fee_type,
			frequency, academic_year, semester, department_id, course_id, is_mandatory, due_date,
			created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`

	err := r.DB.Pool.QueryRow(ctx, sql, fee.CollegeID, fee.Name, fee.Description, fee.Amount,
		fee.Currency, fee.FeeType, fee.Frequency, fee.AcademicYear, fee.Semester, fee.DepartmentID,
		fee.CourseID, fee.IsMandatory, fee.DueDate, fee.CreatedAt, fee.UpdatedAt).Scan(&fee.ID)

	if err != nil {
		return fmt.Errorf("CreateFeeStructure: %w", err)
	}
	return nil
}

func (r *feeRepository) GetFeeStructure(ctx context.Context, feeID int, collegeID int) (*models.FeeStructure, error) {
	sql := `SELECT * FROM fee_structures WHERE id = $1 AND college_id = $2`

	fee := &models.FeeStructure{}
	err := pgxscan.Get(ctx, r.DB.Pool, fee, sql, feeID, collegeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("fee structure not found")
		}
		return nil, fmt.Errorf("GetFeeStructure: %w", err)
	}
	return fee, nil
}

func (r *feeRepository) UpdateFeeStructure(ctx context.Context, fee *models.FeeStructure) error {
	fee.UpdatedAt = time.Now()

	sql := `UPDATE fee_structures SET name = $1, description = $2, amount = $3, fee_type = $4,
			frequency = $5, academic_year = $6, semester = $7, is_mandatory = $8, due_date = $9,
			updated_at = $10 WHERE id = $11 AND college_id = $12`

	result, err := r.DB.Pool.Exec(ctx, sql, fee.Name, fee.Description, fee.Amount, fee.FeeType,
		fee.Frequency, fee.AcademicYear, fee.Semester, fee.IsMandatory, fee.DueDate,
		fee.UpdatedAt, fee.ID, fee.CollegeID)

	if err != nil {
		return fmt.Errorf("UpdateFeeStructure: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("fee structure not found")
	}
	return nil
}

func (r *feeRepository) DeleteFeeStructure(ctx context.Context, feeID int, collegeID int) error {
	sql := `DELETE FROM fee_structures WHERE id = $1 AND college_id = $2`
	result, err := r.DB.Pool.Exec(ctx, sql, feeID, collegeID)
	if err != nil {
		return fmt.Errorf("DeleteFeeStructure: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("fee structure not found")
	}
	return nil
}

func (r *feeRepository) ListFeeStructures(ctx context.Context, filter models.FeeFilter) ([]*models.FeeStructure, error) {
	sql := `SELECT * FROM fee_structures WHERE college_id = $1`
	args := []interface{}{filter.CollegeID}
	paramCount := 1

	if filter.DepartmentID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND department_id = $%d", paramCount)
		args = append(args, *filter.DepartmentID)
	}
	if filter.CourseID != nil {
		paramCount++
		sql += fmt.Sprintf(" AND course_id = $%d", paramCount)
		args = append(args, *filter.CourseID)
	}
	if filter.FeeType != nil {
		paramCount++
		sql += fmt.Sprintf(" AND fee_type = $%d", paramCount)
		args = append(args, *filter.FeeType)
	}

	sql += " ORDER BY due_date ASC, name ASC"

	if filter.Limit > 0 {
		paramCount++
		sql += fmt.Sprintf(" LIMIT $%d", paramCount)
		args = append(args, filter.Limit)
	}

	var fees []*models.FeeStructure
	err := pgxscan.Select(ctx, r.DB.Pool, &fees, sql, args...)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("ListFeeStructures: %w", err)
	}
	return fees, nil
}

func (r *feeRepository) AssignFeeToStudent(ctx context.Context, assignment *models.FeeAssignment) error {
	now := time.Now()
	assignment.CreatedAt = now
	assignment.UpdatedAt = now
	assignment.Status = "pending"

	sql := `INSERT INTO fee_assignments (student_id, fee_structure_id, amount, waiver_amount,
			waiver_reason, due_date, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (student_id, fee_structure_id) DO UPDATE
			SET amount = $3, waiver_amount = $4, waiver_reason = $5, due_date = $6, updated_at = $9
			RETURNING id`

	err := r.DB.Pool.QueryRow(ctx, sql, assignment.StudentID, assignment.FeeStructureID, assignment.Amount,
		assignment.WaiverAmount, assignment.WaiverReason, assignment.DueDate, assignment.Status,
		assignment.CreatedAt, assignment.UpdatedAt).Scan(&assignment.ID)

	if err != nil {
		return fmt.Errorf("AssignFeeToStudent: %w", err)
	}
	return nil
}

func (r *feeRepository) GetFeeAssignment(ctx context.Context, assignmentID int) (*models.FeeAssignment, error) {
	sql := `SELECT fa.*, fs.name, fs.fee_type, fs.currency
			FROM fee_assignments fa
			JOIN fee_structures fs ON fa.fee_structure_id = fs.id
			WHERE fa.id = $1`

	assignment := &models.FeeAssignment{}
	err := pgxscan.Get(ctx, r.DB.Pool, assignment, sql, assignmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("fee assignment not found")
		}
		return nil, fmt.Errorf("GetFeeAssignment: %w", err)
	}
	return assignment, nil
}

func (r *feeRepository) GetStudentFeeAssignments(ctx context.Context, studentID int) ([]*models.FeeAssignment, error) {
	sql := `SELECT fa.*, fs.name, fs.fee_type, fs.currency, fs.description
			FROM fee_assignments fa
			JOIN fee_structures fs ON fa.fee_structure_id = fs.id
			WHERE fa.student_id = $1
			ORDER BY fa.due_date ASC, fa.status ASC`

	var assignments []*models.FeeAssignment
	err := pgxscan.Select(ctx, r.DB.Pool, &assignments, sql, studentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetStudentFeeAssignments: %w", err)
	}

	// Calculate paid and remaining amounts for each assignment
	for _, assignment := range assignments {
		paidAmount, err := r.GetTotalPaidAmount(ctx, assignment.ID)
		if err != nil {
			return nil, err
		}
		assignment.PaidAmount = paidAmount
		assignment.RemainingAmount = assignment.Amount - paidAmount - assignment.WaiverAmount
	}

	return assignments, nil
}

func (r *feeRepository) UpdateFeeAssignmentStatus(ctx context.Context, assignmentID int, status string) error {
	sql := `UPDATE fee_assignments SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.DB.Pool.Exec(ctx, sql, status, time.Now(), assignmentID)
	if err != nil {
		return fmt.Errorf("UpdateFeeAssignmentStatus: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("fee assignment not found")
	}
	return nil
}

func (r *feeRepository) CreateFeePayment(ctx context.Context, payment *models.FeePayment) error {
	now := time.Now()
	payment.CreatedAt = now
	payment.UpdatedAt = now

	if payment.PaymentDate == nil {
		payment.PaymentDate = &now
	}

	sql := `INSERT INTO fee_payments (fee_assignment_id, student_id, amount, currency, payment_method,
			payment_status, transaction_id, gateway, gateway_response, receipt_number, payment_date,
			notes, processed_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING id`

	err := r.DB.Pool.QueryRow(ctx, sql, payment.FeeAssignmentID, payment.StudentID, payment.Amount,
		payment.Currency, payment.PaymentMethod, payment.PaymentStatus, payment.TransactionID,
		payment.Gateway, payment.GatewayResponse, payment.ReceiptNumber, payment.PaymentDate,
		payment.Notes, payment.ProcessedBy, payment.CreatedAt, payment.UpdatedAt).Scan(&payment.ID)

	if err != nil {
		return fmt.Errorf("CreateFeePayment: %w", err)
	}
	return nil
}

func (r *feeRepository) GetFeePayment(ctx context.Context, paymentID int) (*models.FeePayment, error) {
	sql := `SELECT * FROM fee_payments WHERE id = $1`

	payment := &models.FeePayment{}
	err := pgxscan.Get(ctx, r.DB.Pool, payment, sql, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("payment not found")
		}
		return nil, fmt.Errorf("GetFeePayment: %w", err)
	}
	return payment, nil
}

func (r *feeRepository) GetStudentPayments(ctx context.Context, studentID int) ([]*models.FeePayment, error) {
	sql := `SELECT * FROM fee_payments WHERE student_id = $1 ORDER BY payment_date DESC`

	var payments []*models.FeePayment
	err := pgxscan.Select(ctx, r.DB.Pool, &payments, sql, studentID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetStudentPayments: %w", err)
	}
	return payments, nil
}

func (r *feeRepository) UpdatePaymentStatus(ctx context.Context, paymentID int, status string, transactionID *string) error {
	sql := `UPDATE fee_payments SET payment_status = $1, transaction_id = $2, updated_at = $3 WHERE id = $4`
	result, err := r.DB.Pool.Exec(ctx, sql, status, transactionID, time.Now(), paymentID)
	if err != nil {
		return fmt.Errorf("UpdatePaymentStatus: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment not found")
	}
	return nil
}

func (r *feeRepository) GetStudentFeesSummary(ctx context.Context, studentID int) (*models.StudentFeesSummary, error) {
	sql := `SELECT
				COUNT(fa.id) as total_assignments,
				COALESCE(SUM(fa.amount), 0) as total_fees,
				COALESCE(SUM(fa.waiver_amount), 0) as waiver_amount,
				COALESCE(SUM(CASE WHEN fa.status = 'paid' THEN 1 ELSE 0 END), 0) as paid_assignments,
				COALESCE(SUM(CASE WHEN fa.status = 'pending' OR fa.status = 'partial' THEN 1 ELSE 0 END), 0) as pending_assignments,
				COALESCE(SUM(CASE WHEN fa.status = 'overdue' THEN 1 ELSE 0 END), 0) as overdue_assignments,
				COALESCE(SUM(CASE WHEN fa.due_date < NOW() AND fa.status != 'paid' THEN fa.amount ELSE 0 END), 0) as overdue_amount
			FROM fee_assignments fa
			WHERE fa.student_id = $1`

	summary := &models.StudentFeesSummary{}
	err := r.DB.Pool.QueryRow(ctx, sql, studentID).Scan(
		&summary.TotalAssignments,
		&summary.TotalFees,
		&summary.WaiverAmount,
		&summary.PaidAssignments,
		&summary.PendingAssignments,
		&summary.OverdueAssignments,
		&summary.OverdueAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetStudentFeesSummary: %w", err)
	}

	// Get total paid amount
	paidSQL := `SELECT COALESCE(SUM(fp.amount), 0)
				FROM fee_payments fp
				WHERE fp.student_id = $1 AND fp.payment_status = 'completed'`

	err = r.DB.Pool.QueryRow(ctx, paidSQL, studentID).Scan(&summary.PaidAmount)
	if err != nil {
		return nil, fmt.Errorf("GetStudentFeesSummary (paid amount): %w", err)
	}

	summary.PendingAmount = summary.TotalFees - summary.PaidAmount - summary.WaiverAmount

	return summary, nil
}

func (r *feeRepository) GetTotalPaidAmount(ctx context.Context, assignmentID int) (float64, error) {
	sql := `SELECT COALESCE(SUM(amount), 0) FROM fee_payments
			WHERE fee_assignment_id = $1 AND payment_status = 'completed'`

	var total float64
	err := r.DB.Pool.QueryRow(ctx, sql, assignmentID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("GetTotalPaidAmount: %w", err)
	}
	return total, nil
}

// UpdatePaymentStatusByTransactionID updates payment status by the Razorpay order ID (transaction_id)
// This is used for webhook processing where we receive order_id instead of internal payment_id
func (r *feeRepository) UpdatePaymentStatusByTransactionID(ctx context.Context, orderID string, status string, paymentID *string) error {
	var sql string
	var result interface{ RowsAffected() int64 }
	var err error

	if paymentID != nil {
		// Update both status and the actual Razorpay payment ID
		sql = `UPDATE fee_payments SET payment_status = $1, transaction_id = $2, updated_at = $3 
		       WHERE transaction_id = $4 OR (transaction_id IS NULL AND id IN (
		           SELECT id FROM fee_payments WHERE payment_status = 'pending' LIMIT 1
		       ))`
		result, err = r.DB.Pool.Exec(ctx, sql, status, *paymentID, time.Now(), orderID)
	} else {
		sql = `UPDATE fee_payments SET payment_status = $1, updated_at = $2 WHERE transaction_id = $3`
		result, err = r.DB.Pool.Exec(ctx, sql, status, time.Now(), orderID)
	}

	if err != nil {
		return fmt.Errorf("UpdatePaymentStatusByTransactionID: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("payment with order_id %s not found", orderID)
	}
	return nil
}
