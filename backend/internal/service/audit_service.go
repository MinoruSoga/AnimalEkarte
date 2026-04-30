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
	// LogLstepOperationWithMetadata は LogLstepOperation に metadata を加えて永続化する（ISSUE-010）。
	// metadata は any 型で受け取り、内部で JSON シリアライズして audit_logs.metadata jsonb に保存する。
	// nil の場合は metadata を NULL として保存する（=既存の LogLstepOperation と等価）。
	// 既存呼び出しの後方互換のため、LogLstepOperation を破壊せずに新規メソッドとして追加した。
	LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
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
// 既存呼び出しの後方互換のため、metadata=nil で LogLstepOperationWithMetadata に委譲する。
func (s *auditService) LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error {
	return s.LogLstepOperationWithMetadata(ctx, clinicID, actorID, action, resource, resourceID, nil)
}

// LogLstepOperationWithMetadata は metadata 付きで LSTEP 監査ログを記録する（ISSUE-010）。
// metadata が nil の場合は audit_logs.metadata に NULL を保存する。
// metadata の JSON シリアライズに失敗した場合でも監査ログ本体（action / resource / resource_id）は保存される
// （MarshalAuditJSON はベストエフォート: シリアライズ失敗時は nil を返す）。
func (s *auditService) LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error {
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
		Metadata:   repository.MarshalAuditJSON(metadata),
	}
	return s.Log(ctx, log)
}
