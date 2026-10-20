package trimming

import (
	"context"
	"sort"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// ImageKey is a canonical, clinic-owned storage reference. Only this package
// constructs it; lifecycle operations cannot accept a foreign clinic or URL.
type ImageKey struct {
	clinicID uint64
	value    string
}

// ImageLifecycleRepository is the narrow capability shared by reference writes
// and deletion reservations. Locks must last until the ambient transaction ends.
type ImageLifecycleRepository interface {
	LockImageKeys(context.Context, []ImageKey) error
	ImageKeyDeleted(context.Context, ImageKey) (bool, error)
	MarkImageKeyDeleted(context.Context, ImageKey, uint64) error
	// ImageKeyInUse reports whether any trimming detail row or course master
	// still references the key by exact object key match (EMR-457 course
	// images included). exceptAppointmentID excludes one appointment's own
	// detail rows; course writes pass 0 because the course row is updated
	// before the check and can no longer reference the superseded key.
	ImageKeyInUse(context.Context, uint64, string, uint64) (bool, error)
}

func ownedImageKey(clinicID uint64, stored string) (ImageKey, error) {
	key, ok := trimmingImageObjectKey(stored, clinicID)
	if !ok {
		return ImageKey{}, apperrors.WrapInvalidInput("invalid trimming image key")
	}
	return ImageKey{clinicID: clinicID, value: key}, nil
}

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
	}
	return nil
}

// ReserveImageDeletion commits a permanent tombstone before any storage delete.
// A subsequent writer sees the tombstone under the same key lock and rejects the
// reference, even if storage deletion fails or its caller retries. False retains
// a shared image. exceptAppointmentID is only used after that record was deleted;
// PATCH cleanup passes zero and includes the current record in the latest check.
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
		deleted, err := repo.ImageKeyDeleted(txCtx, key)
		if err != nil {
			return err
		}
		if deleted {
			// The first reservation proved no live references and closed the
			// key to writers. Retry must not be blocked by the retired detail
			// row of a soft-deleted appointment that initially was excluded.
			reserved = true
			return nil
		}
		inUse, err := s.trimmingDetail.ImageKeyInUse(txCtx, clinicID, key.value, exceptAppointmentID)
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
