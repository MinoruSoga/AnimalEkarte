package billing

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// consumer-side narrow views（ADR-006/BE9-2C B①）: billing が必要とする未移行 domain repo の
// 最小メソッド集合。具象は service 集約または cmd/api/main.go が注入する。

// merchandiseItemFinder は物販マスタ（inventory domain）の所有権確認に使う最小view。
// ambient transaction 配下では FindByID が FOR SHARE を取り、キャンペーン対象付け替えと
// 物販 soft-delete を直列化する（BE-ACT-CAMPAIGN-TARGET-SERIALIZATION）。
type merchandiseItemFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MerchandiseItem, error)
}

// billingMedicalRecordLocker は medicalrecord domain（topoでbillingより下流・import不可）の
// カルテ確定ロック/取得の最小view（sharedkernel.MedicalRecordLocker + FindByID）。
type billingMedicalRecordLocker interface {
	sharedkernel.MedicalRecordLocker
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// AuditEntry は監査行の domain 内表現（medicalrecord.AuditEntry と同形・service.AuditLogInput へ
// composition root の adapter が写像する）。
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

// billingAuditLogger は監査サービス（未移行）の最小view（best-effort LogEntry・main.go adapter注入）。
type billingAuditLogger interface {
	LogEntry(ctx context.Context, entry *AuditEntry) error
}

// Transactor は repository.Transactor の consumer-side view（WithTx のみ・他domain同型）。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// accountingBillingView は accounting repo（B④未移行）の請求参照/ロックの最小view。
type accountingBillingView interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
	LockAndFindByID(ctx context.Context, clinicID, id uint64) (*model.Billing, error)
}

// treatmentBillingReader は treatment（medicalrecord=topo下流・import不可）の請求参照最小view。
type treatmentBillingReader interface {
	FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
	CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
}

// billingAuditTxLogger は fail-closed 監査（#211・ambient tx 参加 LogEntryTx）の最小view
// （medicalrecord AuditTxLogger と同型・composition root が adapter 注入）。
type billingAuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}

// trimmingCourseFinder / trimmingOptionFinder は trimming マスタ（未移行 domain）の
// 所有権検証（X-4）+一覧の最小view。
type trimmingCourseFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingCourse, error)
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingCourse, error)
}

type trimmingOptionFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingOption, error)
	FindAll(ctx context.Context, clinicID uint64) ([]model.TrimmingOption, error)
}

// billingOwnerReader は飼主割引取得（owner domain）の最小view（#81 段階2b）。
type billingOwnerReader interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

// billingHospitalizationFinder は入院（medicalrecord=topo下流・import不可）のFK検証最小view（AUD-002）。
type billingHospitalizationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Hospitalization, error)
}

// cpmTagSyncer は CPM ステージタグ同期（lstep=topo下流・import不可）の最小view（best-effort）。
type cpmTagSyncer interface {
	SyncCPMStageTag(ctx context.Context, clinicID, ownerID uint64) error
}

// billingClinicReader / billingHolidayReader は clinic domain（未移行）のレポート用最小view。
type billingClinicReader interface {
	FindByID(ctx context.Context, id uint64) (*model.Clinic, error)
}

type billingHolidayReader interface {
	FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
}

// closingScheduleResolver は締め時間スケジュール解決（clinic domain 未移行 service）の最小view。
type closingScheduleResolver interface {
	ResolveSchedule(ctx context.Context, clinicID uint64, date time.Time) (*sharedkernel.DaySchedule, error)
}
