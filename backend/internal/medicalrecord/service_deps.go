package medicalrecord

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
)

// This file declares medicalrecord's consumer-side views of dependencies that live outside
// this package (ADR-006 "aggregator 非経由"): the medical-record repository (kept in
// internal/repository), the shared transaction boundary (repository.Transactor), the shared
// audit kernel (internal/service.AuditService), and the LSTEP tag-sync / delivery-trigger
// services (kept in internal/service pending a later lstep batch). Following Go's "define
// interfaces where they are consumed" guidance and the internal/manualarticle precedent
// (audit.go's AuditLogger), each is the minimal method set this package's services actually
// call — never the full source interface. Concrete implementations are passed in from the
// composition root (cmd/api/main.go, via NewServices in the BE9-2D middle state) by structural
// typing; a nil dependency keeps the original nil-guard semantics of the moved services.

// medicalRecordFinder は checkupService が親カルテ存在/確定状態を読むための最小 view
// （repository.MedicalRecordRepository.FindByID 相当）。
type medicalRecordFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// medicalRecordLocker は LockByIDForUpdate 1メソッドの narrow interface（BE-refactor.md E-5）。
// prescriptionService / checkupFieldResultService が確定と子書込の競合防止（X-11）に使う。
type medicalRecordLocker interface {
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// Transactor は repository.Transactor の consumer-side view（WithTx のみ）。
// prescriptionService / checkupFieldResultService が「削除+挿入+監査」を単一 tx に収めるために使う。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AuditEntry は internal/service.AuditLogInput のうち checkupFieldResultService の
// tx 内削除監査（#211 fail-closed）が実際に設定する部分集合を写したもの。Metadata は
// AuditLogInput.Metadata と field-for-field に一致させるため any を維持する。IPAddress/UserAgent は
// この経路では未使用のため持たない（manualarticle.AuditEntry が Metadata を持たないのと同方針）。
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
}

// AuditTxLogger は共有監査カーネル（internal/service.AuditService の tx 内 LogEntryTx）の
// consumer-side view。composition root が具象 service.AuditService.LogEntryTx(ctx, *service.AuditLogInput)
// を本シグネチャへ adapt する（manualarticle/main.go の adapter 先例）。
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}

// checkupTagSyncer は checkupService が使う LstepTagSyncService の最小 view。nil 許容（未設定なら同期しない）。
type checkupTagSyncer interface {
	SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error
}

// vaccinationTagSyncer は vaccinationService が使う LstepTagSyncService の最小 view。nil 許容。
type vaccinationTagSyncer interface {
	SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error
}

// prescriptionTagSyncer は prescriptionService が使う LstepTagSyncService の最小 view。nil 許容。
type prescriptionTagSyncer interface {
	SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error
}

// checkupFollowUpTrigger は checkupService の健診フォローアップ配信トリガーの最小 view
// （LstepDeliveryTriggerService.TriggerCheckupFollowUp 相当）。nil 許容（未設定ならトリガーしない）。
type checkupFollowUpTrigger interface {
	TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error
}

// ── lab import/report saga consumer-side views (BE9-2D sub-batch③) ──
// The lab services (labImportExaminationService / labReportQueryService / labAuditLogger) moved
// here from internal/service as a leaf domain. Following the same "define interfaces where they
// are consumed" rule as above, each below is the minimal method set the lab services actually
// call over repository.ExaminationRepository / repository.PetRepository / the shared audit kernel.
// The composition root (cmd/api/main.go in the final state; NewServices in the BE9-2D Batch B
// middle state) passes the concrete repository.* implementations in by structural typing.

// examinationImportRepo is labImportExaminationService's write-side view of the examination
// repository: Create + ReplaceItemsByExamID persist the exam and its results, and Delete performs
// the P2-7 orphan-exam compensation (dropping it silently would make a failed row un-retriable).
type examinationImportRepo interface {
	Create(ctx context.Context, exam *model.Examination) error
	ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// examinationReportRepo is labReportQueryService's read-side view of the examination repository
// (clinic-scoped report queries only).
type examinationReportRepo interface {
	FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
}

// petFinder is labImportExaminationService's minimal view of the pet repository, used to verify
// a request-derived pet_id belongs to the import clinic before persisting the exam (P1-2, #186).
type petFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// AuditLogger is the non-tx consumer-side view of the shared audit kernel
// (internal/service.AuditService's LogEntry) that labAuditLogger.logBestEffort writes through.
// Distinct from AuditTxLogger above (tx-internal LogEntryTx): lab audit is best-effort and does
// not join the import flow's transaction. The composition root adapts the concrete
// service.AuditService.LogEntry(ctx, *service.AuditLogInput) to this signature (see the lab audit
// adapter in cmd/api/main.go in the final state / internal/service/lab_middle_state.go in Batch B).
type AuditLogger interface {
	LogEntry(ctx context.Context, entry *AuditEntry) error
}
