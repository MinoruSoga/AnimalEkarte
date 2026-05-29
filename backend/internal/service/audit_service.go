package service

import (
	"context"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type AuditService interface {
	Log(ctx context.Context, log *model.AuditLog) error
	LogEntry(ctx context.Context, input *AuditLogInput) error
	LogAuthLogin(ctx context.Context, clinicID *uint64, staffID *uint64, action string, ipAddress string, userAgent string) error
	// LogLstepOperation はLステップ / LINE連携操作を監査ログに記録する。
	// actorID: 操作スタッフID（nil = システム自動実行）, resource: 対象リソース種別, resourceID: 対象ID
	LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
	// LogLstepOperationWithMetadata は LogLstepOperation に metadata を加えて永続化する（ISSUE-010）。
	// metadata は any 型で受け取り、内部で JSON シリアライズして audit_logs.metadata jsonb に保存する。
	// nil の場合は metadata を NULL として保存する（=既存の LogLstepOperation と等価）。
	// 既存呼び出しの後方互換のため、LogLstepOperation を破壊せずに新規メソッドとして追加した。
	LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
	// LogMedicalRecordChange は医療カルテの Create/Update/Delete/Finalize 操作を監査ログに記録する（AUDIT-H1）。
	// action: "create" | "update" | "delete" | "finalize"
	// actorID=nil はシステム実行として扱う。
	LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error
	// LogVitalChange はバイタル記録の Create/Update/Delete 操作を監査ログに記録する（AUDIT-H1）。
	// metadata に medical_record_id を含める。
	LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error
	// LogAddendumCreate は医療カルテ追記の Create 操作を監査ログに記録する（AUDIT-H1）。
	// new_value に addendum の内容を含める。metadata に medical_record_id を含める。
	LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error
	// LogClinicSwitch はクリニック切替操作を監査ログに記録する（FEAT-374 Phase 2）。
	// old_value={"clinic_id": fromClinicID}, new_value={"clinic_id": toClinicID}。
	// middleware からベストエフォートで呼ばれるため、呼び出し元はエラーを無視してよい。
	LogClinicSwitch(ctx context.Context, actorID *uint64, fromClinicID, toClinicID uint64, ipAddress, userAgent string) error
}

type AuditLogInput struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
	IPAddress  string
	UserAgent  string
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

func (s *auditService) LogEntry(ctx context.Context, input *AuditLogInput) error {
	log := &model.AuditLog{
		ClinicID:   input.ClinicID,
		ActorID:    input.ActorID,
		ActorType:  input.ActorType,
		Action:     input.Action,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		OldValue:   repository.MarshalAuditJSON(input.OldValue),
		NewValue:   repository.MarshalAuditJSON(input.NewValue),
		Metadata:   repository.MarshalAuditJSON(input.Metadata),
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
	}
	return s.Log(ctx, log)
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

// LogMedicalRecordChange は医療カルテの Create/Update/Delete/Finalize 操作を監査ログに記録する（AUDIT-H1）。
func (s *auditService) LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error {
	actorType := "system"
	if actorID != nil {
		actorType = "staff"
	}
	var oldJSON, newJSON []byte
	if oldValue != nil {
		oldJSON = repository.MarshalAuditJSON(oldValue)
	}
	if newValue != nil {
		newJSON = repository.MarshalAuditJSON(newValue)
	}
	log := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   "medical_record",
		ResourceID: &recordID,
		OldValue:   oldJSON,
		NewValue:   newJSON,
	}
	return s.Log(ctx, log)
}

// LogVitalChange はバイタル記録の Create/Update/Delete 操作を監査ログに記録する（AUDIT-H1）。
func (s *auditService) LogVitalChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, vitalID, medicalRecordID uint64, oldValue, newValue map[string]any) error {
	actorType := "system"
	if actorID != nil {
		actorType = "staff"
	}
	metadata := map[string]any{"medical_record_id": medicalRecordID}
	var oldJSON, newJSON []byte
	if oldValue != nil {
		oldJSON = repository.MarshalAuditJSON(oldValue)
	}
	if newValue != nil {
		newJSON = repository.MarshalAuditJSON(newValue)
	}
	log := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     action,
		Resource:   "vital",
		ResourceID: &vitalID,
		OldValue:   oldJSON,
		NewValue:   newJSON,
		Metadata:   repository.MarshalAuditJSON(metadata),
	}
	return s.Log(ctx, log)
}

// LogClinicSwitch はクリニック切替操作を監査ログに記録する（FEAT-374 Phase 2）。
func (s *auditService) LogClinicSwitch(ctx context.Context, actorID *uint64, fromClinicID, toClinicID uint64, ipAddress, userAgent string) error {
	actorType := "system"
	if actorID != nil {
		actorType = "staff"
	}
	oldValue := map[string]any{"clinic_id": fromClinicID}
	newValue := map[string]any{"clinic_id": toClinicID}
	log := &model.AuditLog{
		ClinicID:  &toClinicID,
		ActorID:   actorID,
		ActorType: actorType,
		Action:    "switch_clinic",
		Resource:  "auth",
		OldValue:  repository.MarshalAuditJSON(oldValue),
		NewValue:  repository.MarshalAuditJSON(newValue),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
	return s.Log(ctx, log)
}

// LogAddendumCreate は医療カルテ追記の Create 操作を監査ログに記録する（AUDIT-H1）。
func (s *auditService) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	actorType := "system"
	if actorID != nil {
		actorType = "staff"
	}
	newValue := map[string]any{
		"before_text":    addendum.BeforeText,
		"after_text":     addendum.AfterText,
		"reason":         addendum.Reason,
		"author_user_id": addendum.AuthorUserID,
	}
	metadata := map[string]any{"medical_record_id": medicalRecordID}
	log := &model.AuditLog{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  actorType,
		Action:     "create",
		Resource:   "medical_record_addendum",
		ResourceID: &addendumID,
		NewValue:   repository.MarshalAuditJSON(newValue),
		Metadata:   repository.MarshalAuditJSON(metadata),
	}
	return s.Log(ctx, log)
}
