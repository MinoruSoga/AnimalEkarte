package manualarticle

import "context"

// AuditLogger is manualarticle's consumer-side view of the shared audit kernel
// (internal/service.AuditService — BE9-2A classification: "keep", a permanent shared-kernel
// package, not a BE9-2 migration target). Declaring a minimal local interface + local
// AuditEntry struct — instead of importing internal/service directly — keeps
// internal/manualarticle free of any dependency on internal/service's aggregate Services
// struct (ADR-006 "aggregator 非経由"). The composition root (cmd/api/main.go) adapts the
// concrete service.AuditService.LogEntry(ctx, *service.AuditLogInput) to this signature.
type AuditLogger interface {
	LogEntry(ctx context.Context, entry AuditEntry) error
}

// AuditEntry mirrors the subset of service.AuditLogInput's fields manualarticle actually
// sets (no Metadata — UpsertManualArticle/DeleteManualArticle never use it).
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
