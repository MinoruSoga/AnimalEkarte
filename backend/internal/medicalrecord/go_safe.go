package medicalrecord

import (
	"log/slog"
	"runtime/debug"
)

// goSafe is a documented, byte-for-byte duplicate of internal/service.goSafe (go_safe.go),
// kept local so this package does not import internal/service (ADR-006: the repository→
// medicalrecord facade already makes that edge a cycle). It is a pure, stateless helper —
// declaring an interface for it would be speculative generality (see validators.go's rationale).
// checkupService's fire-and-forget follow-up trigger uses it. Follow-up: collapse both copies
// onto a shared package once a second migrated domain needs it.
func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("background goroutine panicked", "name", name, "panic", r, "stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}
