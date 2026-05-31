package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

func (s *lstepTagSyncService) SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{apperrors.Wrap(err, "failed to check lstep sync enabled")}
	} else if skip {
		return 0, nil
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	var errs []error
	count := 0
	for i := range owners {
		ownerID := owners[i].ID
		syncFns := []struct {
			name string
			fn   func() error
		}{
			{"SyncHealthcheckTags", func() error { return s.SyncHealthcheckTags(ctx, clinicID, ownerID) }},
			{"SyncAnnual4CheckupTag", func() error { return s.SyncAnnual4CheckupTag(ctx, clinicID, ownerID) }},
			{"SyncVaccineDeadlineTag", func() error { return s.SyncVaccineDeadlineTag(ctx, clinicID, ownerID) }},
			{"SyncFilariaTag", func() error { return s.SyncFilariaTag(ctx, clinicID, ownerID) }},
			{"SyncFleaTickTag", func() error { return s.SyncFleaTickTag(ctx, clinicID, ownerID) }},
			{"SyncFoodPurchaseTag", func() error { return s.SyncFoodPurchaseTag(ctx, clinicID, ownerID) }},
		}
		ownerFailed := false
		for _, sf := range syncFns {
			if syncErr := sf.fn(); syncErr != nil {
				slog.ErrorContext(ctx, "health-prevention batch: sync failed",
					"clinic_id", clinicID, "owner_id", ownerID, "method", sf.name, "error", syncErr)
				errs = append(errs, apperrors.Wrap(syncErr, sf.name))
				ownerFailed = true
			}
		}
		if !ownerFailed {
			count++
		}
	}
	return count, errs
}
