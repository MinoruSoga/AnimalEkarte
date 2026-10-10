package trimming

import (
	"context"
	"log/slog"
)

// Sync only after the outer appointment/detail/audit transaction commits. External
// failures are best-effort and must neither reverse the write nor skip other owners.
func (s *trimmingService) syncNextVisitTagsBestEffort(ctx context.Context, clinicID uint64, ownerIDs ...*uint64) {
	if s.nextVisitTagSync == nil {
		return
	}
	seen := make(map[uint64]struct{}, len(ownerIDs))
	for _, id := range ownerIDs {
		if id == nil || *id == 0 {
			continue
		}
		if _, duplicate := seen[*id]; duplicate {
			continue
		}
		seen[*id] = struct{}{}
		if err := s.nextVisitTagSync.SyncNextVisitTag(ctx, clinicID, *id); err != nil {
			slog.WarnContext(ctx, "next visit tag sync failed (best-effort)", slog.Uint64("clinic_id", clinicID), slog.Uint64("owner_id", *id), slog.Any("error", err))
		}
	}
}
