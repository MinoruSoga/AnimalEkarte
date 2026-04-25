package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, log *model.AuditLog) error
	LogAuthLogin(ctx context.Context, clinicID *uint64, staffID *uint64, action string, ipAddress string, userAgent string) error
	// LogLstepOperation はLステップ / LINE連携操作を監査ログに記録する。
	// actorID: 操作スタッフID（nil = システム自動実行）, resource: 対象リソース種別, resourceID: 対象ID
	LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

// Log は監査ログを記録する
func (s *auditService) Log(ctx context.Context, log *model.AuditLog) error {
	if err := s.repo.Create(ctx, log); err != nil {
		return apperrors.Wrap(err, "failed to create audit log")
	}
	return nil
}

// LogAuthLogin は認証イベントログを記録する
func (s *auditService) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ipAddress, userAgent string) error {
	log := &model.AuditLog{
		ClinicID:  clinicID,
		ActorID:   staffID,
		ActorType: "staff",
		Action:    action,
		Resource:  "auth",
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	return s.Log(ctx, log)
}

// LogLstepOperation はLステップ / LINE連携操作を監査ログに記録する。
func (s *auditService) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
	actorType := "system"
	if actorID != nil {
		actorType = "staff"
	}
	log := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
	}
	return s.Log(ctx, log)
}
