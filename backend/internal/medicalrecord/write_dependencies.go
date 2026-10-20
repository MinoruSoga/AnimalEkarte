package medicalrecord

import "context"

// Transactor は repository.Transactor の consumer-side view（WithTx のみ）。
// checkupFieldResultService などが「削除+挿入+監査」を単一 tx に収めるために使う。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AuditEntry は共有監査入力のうち checkupFieldResultService の
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

// AuditTxLogger は共有監査カーネル（tx 内 LogEntryTx）の
// consumer-side view。composition root が具象 service.AuditService.LogEntryTx(ctx, *service.AuditLogInput)
// を本シグネチャへ adapt する（manualarticle/main.go の adapter 先例）。
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}
