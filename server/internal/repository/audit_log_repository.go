package repository

import (
	"context"
	"fmt"
	"time"

	"eduhub/server/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type AuditLogRepository interface {
	CreateAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, collegeID int, userID *int, action, entity string, limit, offset int) ([]*models.AuditLog, error)
	GetAuditLogsByUser(ctx context.Context, collegeID, userID, limit int) ([]*models.AuditLog, error)
	GetAuditLogsByEntity(ctx context.Context, collegeID int, entityType string, entityID int) ([]*models.AuditLog, error)
	CountAuditLogs(ctx context.Context, collegeID int) (int, error)
	CountAuditLogsSince(ctx context.Context, collegeID int, since time.Time) (int, error)
	GetAuditActionCounts(ctx context.Context, collegeID int) (map[string]int, error)
	GetAuditEntityCounts(ctx context.Context, collegeID int) (map[string]int, error)
	GetTopAuditUsers(ctx context.Context, collegeID, limit int) ([]AuditUserSummary, error)
}

type AuditUserSummary struct {
	UserID      int    `db:"user_id"`
	UserName    string `db:"user_name"`
	ActionCount int    `db:"action_count"`
}

type auditLogRepository struct {
	DB *DB
}

func NewAuditLogRepository(db *DB) AuditLogRepository {
	return &auditLogRepository{DB: db}
}

func (r *auditLogRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	log.Timestamp = time.Now()

	sql := `INSERT INTO audit_logs (college_id, user_id, action, entity_type, entity_id, changes, ip_address, user_agent, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	var id int
	err := r.DB.Pool.QueryRow(ctx, sql,
		log.CollegeID,
		log.UserID,
		log.Action,
		log.EntityType,
		log.EntityID,
		log.Changes,
		log.IPAddress,
		log.UserAgent,
		log.Timestamp,
	).Scan(&id)

	if err != nil {
		return err
	}

	log.ID = id
	return nil
}

func (r *auditLogRepository) GetAuditLogs(ctx context.Context, collegeID int, userID *int, action, entity string, limit, offset int) ([]*models.AuditLog, error) {
	sql := `SELECT * FROM audit_logs WHERE college_id = $1`
	args := []any{collegeID}
	idx := 2

	if userID != nil {
		sql += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *userID)
		idx++
	}

	if action != "" {
		sql += fmt.Sprintf(" AND action = $%d", idx)
		args = append(args, action)
		idx++
	}

	if entity != "" {
		sql += fmt.Sprintf(" AND entity_type = $%d", idx)
		args = append(args, entity)
		idx++
	}

	sql += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	var logs []*models.AuditLog
	err := pgxscan.Select(ctx, r.DB.Pool, &logs, sql, args...)
	return logs, err
}

func (r *auditLogRepository) GetAuditLogsByUser(ctx context.Context, collegeID, userID, limit int) ([]*models.AuditLog, error) {
	sql := `SELECT * FROM audit_logs WHERE college_id = $1 AND user_id = $2
			ORDER BY timestamp DESC LIMIT $3`

	var logs []*models.AuditLog
	err := pgxscan.Select(ctx, r.DB.Pool, &logs, sql, collegeID, userID, limit)
	return logs, err
}

func (r *auditLogRepository) GetAuditLogsByEntity(ctx context.Context, collegeID int, entityType string, entityID int) ([]*models.AuditLog, error) {
	sql := `SELECT * FROM audit_logs WHERE college_id = $1 AND entity_type = $2 AND entity_id = $3
			ORDER BY timestamp DESC`

	var logs []*models.AuditLog
	err := pgxscan.Select(ctx, r.DB.Pool, &logs, sql, collegeID, entityType, entityID)
	return logs, err
}

func (r *auditLogRepository) CountAuditLogs(ctx context.Context, collegeID int) (int, error) {
	var total int
	err := r.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE college_id = $1`, collegeID).Scan(&total)
	return total, err
}

func (r *auditLogRepository) CountAuditLogsSince(ctx context.Context, collegeID int, since time.Time) (int, error) {
	var total int
	err := r.DB.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs WHERE college_id = $1 AND timestamp >= $2`, collegeID, since).Scan(&total)
	return total, err
}

func (r *auditLogRepository) GetAuditActionCounts(ctx context.Context, collegeID int) (map[string]int, error) {
	rows, err := r.DB.Pool.Query(ctx, `SELECT action, COUNT(*) FROM audit_logs WHERE college_id = $1 GROUP BY action`, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		result[action] = count
	}

	return result, nil
}

func (r *auditLogRepository) GetAuditEntityCounts(ctx context.Context, collegeID int) (map[string]int, error) {
	rows, err := r.DB.Pool.Query(ctx, `SELECT entity_type, COUNT(*) FROM audit_logs WHERE college_id = $1 GROUP BY entity_type`, collegeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var entity string
		var count int
		if err := rows.Scan(&entity, &count); err != nil {
			return nil, err
		}
		result[entity] = count
	}

	return result, nil
}

func (r *auditLogRepository) GetTopAuditUsers(ctx context.Context, collegeID, limit int) ([]AuditUserSummary, error) {
	rows, err := r.DB.Pool.Query(ctx, `SELECT l.user_id, COALESCE(u.name, '') AS user_name, COUNT(*) AS action_count
		FROM audit_logs l
		LEFT JOIN users u ON u.id = l.user_id
		WHERE l.college_id = $1
		GROUP BY l.user_id, u.name
		ORDER BY action_count DESC
		LIMIT $2`, collegeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]AuditUserSummary, 0)
	for rows.Next() {
		var summary AuditUserSummary
		if err := rows.Scan(&summary.UserID, &summary.UserName, &summary.ActionCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
