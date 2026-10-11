package trimming

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Capture values while holding the appointment lock, before any writer can mutate
// its returned model. Detail, end time and doctor do not affect next-visit queries.
type trimmingNextVisitFacts struct {
	ownerID, petID          uint64
	status                  model.ReservationStatus
	startTime, intendedDate time.Time
}

func trimmingNextVisitFactsFor(appointment *model.Reservation) trimmingNextVisitFacts {
	facts := trimmingNextVisitFacts{status: appointment.Status, startTime: appointment.StartTime}
	if appointment.OwnerID != nil {
		facts.ownerID = *appointment.OwnerID
	}
	if appointment.PetID != nil {
		facts.petID = *appointment.PetID
	}
	if appointment.IntendedDate != nil {
		facts.intendedDate = *appointment.IntendedDate
	}
	return facts
}

type trimmingNextVisitTagChange struct {
	previousOwnerID, ownerID uint64
}

func trimmingNextVisitChange(before trimmingNextVisitFacts, result *model.Reservation) *trimmingNextVisitTagChange {
	after := trimmingNextVisitFactsFor(result)
	if before.ownerID == after.ownerID && before.petID == after.petID && before.status == after.status &&
		before.startTime.Equal(after.startTime) && before.intendedDate.Equal(after.intendedDate) {
		return nil
	}
	return &trimmingNextVisitTagChange{previousOwnerID: before.ownerID, ownerID: after.ownerID}
}

func (s *trimmingService) syncNextVisitTagChange(ctx context.Context, clinicID uint64, change *trimmingNextVisitTagChange) {
	if change != nil {
		s.syncNextVisitTagsBestEffort(ctx, clinicID, &change.previousOwnerID, &change.ownerID)
	}
}

// Bound the entire post-commit batch, including every owner and warning log.
const trimmingNextVisitTagSyncTimeout = 30 * time.Second

// Sync only after the outer appointment/detail/audit transaction commits. External
// failures are best-effort and must neither reverse the write nor skip other owners.
func (s *trimmingService) syncNextVisitTagsBestEffort(ctx context.Context, clinicID uint64, ownerIDs ...*uint64) {
	if s.nextVisitTagSync == nil {
		return
	}
	// Keep trace values, but do not reuse request cancellation or a completed tx.
	syncCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), trimmingNextVisitTagSyncTimeout)
	defer cancel()

	seen := make(map[uint64]struct{}, len(ownerIDs))
	for _, id := range ownerIDs {
		if id == nil || *id == 0 {
			continue
		}
		if _, duplicate := seen[*id]; duplicate {
			continue
		}
		seen[*id] = struct{}{}
		if err := s.nextVisitTagSync.SyncNextVisitTag(syncCtx, clinicID, *id); err != nil {
			slog.WarnContext(syncCtx, "next visit tag sync failed (best-effort)", slog.Uint64("clinic_id", clinicID), slog.Uint64("owner_id", *id), slog.Any("error", err))
		}
	}
}
