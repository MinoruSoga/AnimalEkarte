package reservation

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ReservationTypeFinderView は appointment 系（R④）残留 consumer が service 側 alias 経由で
// 参照する exported view（R④移動時に unexported へ戻す）。
type ReservationTypeFinderView = reservationTypeFinder

type reservationTypeFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationType, error)
}

type reservationTypeCapacityCounter interface {
	CountByTypeAndStartTime(ctx context.Context, clinicID, reservationTypeID uint64, startTime time.Time, excludeID *uint64) (int64, error)
}

// reservationTypeCapacityBatchCounter is an optional capability of reservationTypeCapacityCounter
// implementations: batched capacity counting across multiple start_time values in one query.
// BE-refactor.md R2-4 (D8): liff_service.FilterSlotsByCapacity issued one CountByTypeAndStartTime
// query per candidate slot (N+1 per date). Kept as a separate, narrow interface — not merged into
// reservationTypeCapacityCounter — so existing single-count callers (CheckReservationTypeCapacity)
// and their test mocks are unaffected. FilterSlotsByCapacity type-asserts for it and falls back to
// the per-slot path when unavailable (same optional-capability pattern as
// auditSvc.(AuditTxLogger) in service.go).
type reservationTypeCapacityBatchCounter interface {
	// CountByTypeAndStartTimes returns counts keyed by startTime.Unix() (see repository doc comment
	// for why Unix seconds, not time.Time, is used as the map key).
	CountByTypeAndStartTimes(ctx context.Context, clinicID, reservationTypeID uint64, startTimes []time.Time, excludeID *uint64) (map[int64]int64, error)
}

func CheckReservationTypeCapacity(
	ctx context.Context,
	reservationRepo reservationTypeCapacityCounter,
	typeFinder reservationTypeFinder,
	clinicID, reservationTypeID uint64,
	startTime time.Time,
	excludeID *uint64,
) error {
	if reservationRepo == nil || typeFinder == nil {
		return nil
	}
	reservationType, err := typeFinder.FindByID(ctx, clinicID, reservationTypeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find reservation type for capacity")
	}
	if reservationType.MaxConcurrent == nil {
		return nil
	}
	count, err := reservationRepo.CountByTypeAndStartTime(ctx, clinicID, reservationTypeID, startTime, excludeID)
	if err != nil {
		return apperrors.Wrap(err, "failed to count reservations by type and start time")
	}
	if count >= int64(*reservationType.MaxConcurrent) {
		return apperrors.WrapConflict("この時間帯の予約受入枠が満員です")
	}
	return nil
}
