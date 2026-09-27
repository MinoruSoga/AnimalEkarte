package support

import "context"

// AuditLogger is support's consumer-side view of the shared audit kernel —
// same seam as manualarticle.AuditLogger (ADR-006 "aggregator 非経由").
// The composition root adapts service.AuditService to this signature.
type AuditLogger interface {
	LogEntry(ctx context.Context, entry AuditEntry) error
}

// AuditEntry mirrors the subset of the shared audit input support actually sets.
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	IPAddress  string
	UserAgent  string
}
