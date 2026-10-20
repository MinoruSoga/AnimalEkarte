package trimming

import (
	"context"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func (s *trimmingService) imageLifecycle() (ImageLifecycleRepository, error) {
	repo, ok := s.trimmingDetail.(ImageLifecycleRepository)
	if !ok || repo == nil || s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("trimming image lifecycle is not configured")
	}
	return repo, nil
}

// lockImageReferencesInTx protects every final object reference until commit.
// Writers take booking/appointment locks first, then image locks in key order.
// Deletion only takes image locks and never waits for a booking/appointment lock.
func (s *trimmingService) lockImageReferencesInTx(ctx context.Context, clinicID uint64, stored ...string) error {
	keys := make([]ImageKey, 0, len(stored))
	seen := make(map[string]bool, len(stored))
	for _, value := range stored {
		// Request fields are normalized before the transaction. Unchanged
		// legacy/foreign values remain readable and are never ours to delete.
		key, isObject := trimmingImageObjectKey(value, clinicID)
		if isObject && !seen[key] {
			seen[key] = true
			keys = append(keys, ImageKey{clinicID: clinicID, value: key})
		}
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].value < keys[j].value })
	repo, err := s.imageLifecycle()
	if err != nil {
		return err
	}
	if err := repo.LockImageKeys(ctx, keys); err != nil {
		return apperrors.Wrap(err, "failed to lock trimming image references")
	}
	for _, key := range keys {
		deleted, err := repo.ImageKeyDeleted(ctx, key)
		if err != nil {
			return apperrors.Wrap(err, "failed to check trimming image lifecycle")
		}
		if deleted {
			return apperrors.WrapConflict("trimming image has been deleted; upload it again")
		}
		guard, ok := repo.(interface {
			ProtectCourseImageReference(context.Context, ImageKey) error
		})
		if !ok {
			return apperrors.WrapInternalServerError("course provenance guard unavailable")
		}
		if err := guard.ProtectCourseImageReference(ctx, key); err != nil {
			return err
		}

	}
	return nil
}

// ReserveImageDeletion commits a permanent tombstone before any storage delete.
// A subsequent writer sees the tombstone under the same key lock and rejects the
// reference, even if storage deletion fails or its caller retries. False retains
// every key without positive never-submitted provenance. Saved/retired detail
// references are never excluded; legacy clinical cleanup is retained/deferred.
func (s *trimmingService) ReserveImageDeletion(ctx context.Context, clinicID uint64, stored string, exceptAppointmentID, actorID uint64) (bool, error) {
	key, err := ownedImageKey(clinicID, stored)
	if err != nil {
		return false, err
	}
	if actorID == 0 {
		return false, apperrors.WrapInvalidInput("image deletion requires a staff actor")
	}
	repo, err := s.imageLifecycle()
	if err != nil {
		return false, err
	}
	reserved := false
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := repo.LockImageKeys(txCtx, []ImageKey{key}); err != nil {
			return err
		}
		guard, ok := repo.(interface {
			CourseImageDiscardEligible(context.Context, ImageKey, uint64) (bool, error)
		})
		if !ok {
			return apperrors.WrapInternalServerError("course provenance guard unavailable")
		}
		eligible, err := guard.CourseImageDiscardEligible(txCtx, key, actorID)
		if err != nil {
			return err
		}
		if !eligible {
			return nil
		}
		_, err = repo.ImageKeyDeleted(txCtx, key)
		if err != nil {
			return err
		}
		inUse, err := s.trimmingDetail.ImageKeyInUse(txCtx, clinicID, key.value, 0)
		if err != nil {
			return err
		}
		if inUse {
			return nil
		}
		if err := repo.MarkImageKeyDeleted(txCtx, key, actorID); err != nil {
			return err
		}
		reserved = true
		return nil
	}); err != nil {
		return false, apperrors.Wrap(err, "failed to reserve trimming image deletion")
	}
	return reserved, nil
}
