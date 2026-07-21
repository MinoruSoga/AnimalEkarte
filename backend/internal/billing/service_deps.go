package billing

import (
	"context"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// consumer-side narrow views（ADR-006/BE9-2C B①）: billing が必要とする未移行 domain repo の
// 最小メソッド集合。具象は service 集約または cmd/api/main.go が注入する。

// merchandiseItemFinder は物販マスタ（inventory domain）の所有権確認に使う最小view。
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
