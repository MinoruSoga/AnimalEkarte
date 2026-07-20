package service

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// normalizeIPAddress は空文字列を nil に正規化する（X-3）。audit_logs.ip_address は実DDLで
// inet NULL のため、空文字列をそのまま INSERT すると 22P02 (invalid_text_representation) になる。
// IPAddress を一度もセットしない呼び出し元（LogLstepOperationWithMetadata 等）のゼロ値も含め、
// model.AuditLog を組み立てる箇所はすべて本関数を経由させる。
func normalizeIPAddress(ip string) *string {
	if ip == "" {
		return nil
	}
	return &ip
}

// buildAuditLog は AuditLogInput を model.AuditLog に変換する（LogEntry / LogEntryTx 共通の buildFunc）。
func buildAuditLog(input *AuditLogInput) *model.AuditLog {
	return &model.AuditLog{
		ClinicID:   input.ClinicID,
		ActorID:    input.ActorID,
		ActorType:  input.ActorType,
		Action:     input.Action,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		OldValue:   repository.MarshalAuditJSON(input.OldValue),
		NewValue:   repository.MarshalAuditJSON(input.NewValue),
		Metadata:   repository.MarshalAuditJSON(input.Metadata),
		IPAddress:  normalizeIPAddress(input.IPAddress),
		UserAgent:  input.UserAgent,
	}
}

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

// AuditTxLogger は caller の ambient transaction に参加して監査ログを記録する経路を提供する（#211）。
// AuditService とは独立の小さなインターフェースとして定義し、tx 内監査（fail-closed）を要する
// service のみが依存する。AuditService 本体を広げないことで、既存利用者（refund 等）や
// 各 service の AuditService モックへの波及を避ける（後方互換）。具象 *auditService が本 IF も実装する。
type AuditTxLogger interface {
	// LogEntryTx は ctx の ambient transaction（Transactor.WithTx）に参加して監査ログを書き込む。
	// 監査書込が失敗したらエラーを返し、caller が tx を rollback することで監査も操作も巻き戻る。
	LogEntryTx(ctx context.Context, input *AuditLogInput) error
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

// 具象 *auditService が tx 内監査の AuditTxLogger を実装することをコンパイル時に保証する（#211）。
// service.go の auditSvc.(AuditTxLogger) アサーションが、将来 concrete 型が変わった際に
// ランタイム panic でなくビルドエラーで検出されるようにする。
var _ AuditTxLogger = (*auditService)(nil)

func validateAuditLog(log *model.AuditLog) error {
	if log == nil {
		return apperrors.WrapInvalidInput("audit log is required")
	}
	if log.ClinicID == nil || *log.ClinicID == 0 {
		return apperrors.WrapInvalidInput("audit log clinic_id is required")
	}
	if strings.TrimSpace(log.ActorType) == "" {
		return apperrors.WrapInvalidInput("audit log actor_type is required")
	}
	if strings.TrimSpace(log.Action) == "" {
		return apperrors.WrapInvalidInput("audit log action is required")
	}
	if strings.TrimSpace(log.Resource) == "" {
		return apperrors.WrapInvalidInput("audit log resource is required")
	}
	switch log.ActorType {
	case model.AuditActorTypeStaff:
		if log.ActorID == nil || *log.ActorID == 0 {
			return apperrors.WrapInvalidInput("audit log actor_id is required for staff actor")
		}
	case model.AuditActorTypeSystem:
		if log.ActorID != nil {
			return apperrors.WrapInvalidInput("audit log actor_id must be empty for system actor")
		}
	default:
		return apperrors.WrapInvalidInput("audit log actor_type is invalid")
	}
	return nil
}

// auditActorTypeFor は actorID の有無から監査ログの actor 種別を導出する。
func auditActorTypeFor(actorID *uint64) string {
	return sharedkernel.AuditActorTypeFor(actorID)
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

// Log は監査ログを記録する。
// repo.Create 失敗時は統一キー "audit_write_failed" で ErrorContext に記録する（PERF-AUDIT-TX P1）。
// 分散していた呼び出し側の既存 Warn ログはそのまま残す（挙動保存）。ここは best-effort 書込の
// 全経路が集約する唯一の箇所のため、STG 監視クエリ（STG-CONTINUOUS-OPERATIONS.md）はこのキーのみを見る。
// 戻り値の Wrap 挙動・呼び出し側のエラーハンドリングは不変。
func (s *auditService) Log(ctx context.Context, log *model.AuditLog) error {
	if err := validateAuditLog(log); err != nil {
		return err
	}
	if err := s.repo.Create(ctx, log); err != nil {
		// validateAuditLog を通過済みのため log.ClinicID は non-nil。
		slog.ErrorContext(ctx, "audit_write_failed",
			"action", log.Action,
			"resource", log.Resource,
			"clinic_id", *log.ClinicID,
			"error", err,
		)
		return apperrors.Wrap(err, "failed to create audit log")
	}
	return nil
}

func (s *auditService) LogEntry(ctx context.Context, input *AuditLogInput) error {
	return s.Log(ctx, buildAuditLog(input))
}

// LogEntryTx は caller の ambient transaction に参加して監査ログを記録する（#211 tx 内監査）。
// repo.CreateTx 経由で dbOrTx を使うため、Transactor.WithTx 内から呼ぶと監査書込が同一 tx に入り、
// caller がエラーを返して tx を rollback すれば監査書込も巻き戻る（fail-closed の原子監査）。
// 既存 LogEntry / Log（非 tx・base db 書込）は不変。
func (s *auditService) LogEntryTx(ctx context.Context, input *AuditLogInput) error {
	log := buildAuditLog(input)
	if err := validateAuditLog(log); err != nil {
		return err
	}
	if err := s.repo.CreateTx(ctx, log); err != nil {
		return apperrors.Wrap(err, "failed to create audit log")
	}
	return nil
}

// LogAuthLogin は認証イベントログを記録する
func (s *auditService) LogAuthLogin(ctx context.Context, clinicID, staffID *uint64, action, ipAddress, userAgent string) error {
	log := &model.AuditLog{
		ClinicID:  clinicID,
		ActorID:   staffID,
		ActorType: model.AuditActorTypeStaff,
		Action:    action,
		Resource:  "auth",
		IPAddress: normalizeIPAddress(ipAddress),
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
	actorType := auditActorTypeFor(actorID)
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
	actorType := auditActorTypeFor(actorID)
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
	actorType := auditActorTypeFor(actorID)
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
	actorType := auditActorTypeFor(actorID)
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
		IPAddress: normalizeIPAddress(ipAddress),
		UserAgent: userAgent,
	}
	return s.Log(ctx, log)
}

// LogAddendumCreate は医療カルテ追記の Create 操作を監査ログに記録する（AUDIT-H1）。
func (s *auditService) LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error {
	actorType := auditActorTypeFor(actorID)
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
