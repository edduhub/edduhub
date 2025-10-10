package audit

import (
	"context"
	"time"

	"eduhub/server/internal/models"
	"eduhub/server/internal/repository"
)

type AuditStats struct {
	TotalLogs      int            `json:"total_logs"`
	LogsByAction   map[string]int `json:"logs_by_action"`
	LogsByEntity   map[string]int `json:"logs_by_entity"`
	TopUsers       []UserActivity `json:"top_users"`
	RecentActivity int            `json:"recent_activity_24h"`
}

type UserActivity struct {
	UserID      int    `json:"user_id"`
	UserName    string `json:"user_name"`
	ActionCount int    `json:"action_count"`
}

type AuditService interface {
	LogAction(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, collegeID int, userID *int, action, entity string, limit, offset int) ([]*models.AuditLog, error)
	GetUserActivity(ctx context.Context, collegeID, userID, limit int) ([]*models.AuditLog, error)
	GetEntityHistory(ctx context.Context, collegeID int, entityType string, entityID int) ([]*models.AuditLog, error)
	GetAuditStats(ctx context.Context, collegeID int) (*AuditStats, error)
}

type auditService struct {
	auditRepo repository.AuditLogRepository
}

func NewAuditService(auditRepo repository.AuditLogRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
	}
}

func (s *auditService) LogAction(ctx context.Context, log *models.AuditLog) error {
	log.Timestamp = time.Now()
	return s.auditRepo.CreateAuditLog(ctx, log)
}

func (s *auditService) GetAuditLogs(ctx context.Context, collegeID int, userID *int, action, entity string, limit, offset int) ([]*models.AuditLog, error) {
	return s.auditRepo.GetAuditLogs(ctx, collegeID, userID, action, entity, limit, offset)
}

func (s *auditService) GetUserActivity(ctx context.Context, collegeID, userID, limit int) ([]*models.AuditLog, error) {
	return s.auditRepo.GetAuditLogsByUser(ctx, collegeID, userID, limit)
}

func (s *auditService) GetEntityHistory(ctx context.Context, collegeID int, entityType string, entityID int) ([]*models.AuditLog, error) {
	return s.auditRepo.GetAuditLogsByEntity(ctx, collegeID, entityType, entityID)
}

func (s *auditService) GetAuditStats(ctx context.Context, collegeID int) (*AuditStats, error) {
	total, err := s.auditRepo.CountAuditLogs(ctx, collegeID)
	if err != nil {
		return nil, err
	}

	actionCounts, err := s.auditRepo.GetAuditActionCounts(ctx, collegeID)
	if err != nil {
		return nil, err
	}

	entityCounts, err := s.auditRepo.GetAuditEntityCounts(ctx, collegeID)
	if err != nil {
		return nil, err
	}

	userSummaries, err := s.auditRepo.GetTopAuditUsers(ctx, collegeID, 5)
	if err != nil {
		return nil, err
	}
	topUsers := make([]UserActivity, 0, len(userSummaries))
	for _, summary := range userSummaries {
		if summary.UserID == 0 {
			continue
		}
		topUsers = append(topUsers, UserActivity{
			UserID:      summary.UserID,
			UserName:    summary.UserName,
			ActionCount: summary.ActionCount,
		})
	}

	recent, err := s.auditRepo.CountAuditLogsSince(ctx, collegeID, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, err
	}

	return &AuditStats{
		TotalLogs:      total,
		LogsByAction:   actionCounts,
		LogsByEntity:   entityCounts,
		TopUsers:       topUsers,
		RecentActivity: recent,
	}, nil
}
