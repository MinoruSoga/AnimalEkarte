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
func (s *auditService) LogAuthLogin(ctx context.Context, clinicID *uint64, staffID *uint64, action string, ipAddress string, userAgent string) error {
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
