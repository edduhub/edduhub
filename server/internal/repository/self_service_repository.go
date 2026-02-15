package repository

import (
	"context"
	"fmt"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type SelfServiceRepository interface {
	ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error)
	CreateRequest(ctx context.Context, req *models.SelfServiceRequest) error
	ListByStudent(ctx context.Context, collegeID, studentID int) ([]*models.SelfServiceRequest, error)
	GetByID(ctx context.Context, collegeID, requestID int) (*models.SelfServiceRequest, error)
	UpdateRequest(ctx context.Context, collegeID, requestID int, status, response string, respondedBy int) (*models.SelfServiceRequest, error)
}

type selfServiceRepository struct {
	DB *DB
}

func NewSelfServiceRepository(db *DB) SelfServiceRepository {
	return &selfServiceRepository{DB: db}
}

func (r *selfServiceRepository) ResolveUserIDByKratosID(ctx context.Context, kratosID string) (int, error) {
	var userID int
	err := r.DB.Pool.QueryRow(ctx, `SELECT id FROM users WHERE kratos_identity_id = $1 AND is_active = TRUE`, kratosID).Scan(&userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, err
	}
	return userID, nil
}

func (r *selfServiceRepository) CreateRequest(ctx context.Context, req *models.SelfServiceRequest) error {
	query := `
		INSERT INTO self_service_requests (
			student_id, college_id, type, title, description, status,
			document_type, delivery_method
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, student_id, college_id, type, title, description, status,
			document_type, delivery_method, admin_response, responded_by,
			responded_at, submitted_at, created_at, updated_at`

	return r.DB.Pool.QueryRow(
		ctx,
		query,
		req.StudentID,
		req.CollegeID,
		req.Type,
		req.Title,
		req.Description,
		req.Status,
		req.DocumentType,
		req.DeliveryMethod,
	).Scan(
		&req.ID,
		&req.StudentID,
		&req.CollegeID,
		&req.Type,
		&req.Title,
		&req.Description,
		&req.Status,
		&req.DocumentType,
		&req.DeliveryMethod,
		&req.AdminResponse,
		&req.RespondedBy,
		&req.RespondedAt,
		&req.SubmittedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
}

func (r *selfServiceRepository) ListByStudent(ctx context.Context, collegeID, studentID int) ([]*models.SelfServiceRequest, error) {
	query := `
		SELECT id, student_id, college_id, type, title, description, status,
			document_type, delivery_method, admin_response, responded_by,
			responded_at, submitted_at, created_at, updated_at
		FROM self_service_requests
		WHERE college_id = $1 AND student_id = $2
		ORDER BY submitted_at DESC, id DESC`

	items := make([]*models.SelfServiceRequest, 0)
	if err := pgxscan.Select(ctx, r.DB.Pool, &items, query, collegeID, studentID); err != nil {
		return nil, fmt.Errorf("ListByStudent: %w", err)
	}
	return items, nil
}

func (r *selfServiceRepository) GetByID(ctx context.Context, collegeID, requestID int) (*models.SelfServiceRequest, error) {
	query := `
		SELECT id, student_id, college_id, type, title, description, status,
			document_type, delivery_method, admin_response, responded_by,
			responded_at, submitted_at, created_at, updated_at
		FROM self_service_requests
		WHERE college_id = $1 AND id = $2`

	item := &models.SelfServiceRequest{}
	if err := pgxscan.Get(ctx, r.DB.Pool, item, query, collegeID, requestID); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return item, nil
}

func (r *selfServiceRepository) UpdateRequest(ctx context.Context, collegeID, requestID int, status, response string, respondedBy int) (*models.SelfServiceRequest, error) {
	query := `
		UPDATE self_service_requests
		SET status = $1,
			admin_response = $2,
			responded_by = $3,
			responded_at = NOW(),
			updated_at = NOW()
		WHERE college_id = $4 AND id = $5
		RETURNING id, student_id, college_id, type, title, description, status,
			document_type, delivery_method, admin_response, responded_by,
			responded_at, submitted_at, created_at, updated_at`

	item := &models.SelfServiceRequest{}
	err := r.DB.Pool.QueryRow(ctx, query, status, response, respondedBy, collegeID, requestID).Scan(
		&item.ID,
		&item.StudentID,
		&item.CollegeID,
		&item.Type,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.DocumentType,
		&item.DeliveryMethod,
		&item.AdminResponse,
		&item.RespondedBy,
		&item.RespondedAt,
		&item.SubmittedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("UpdateRequest: %w", err)
	}
	return item, nil
}
