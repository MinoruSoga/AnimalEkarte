package trimming

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type imageDeletionRetryStore struct {
	db *gorm.DB
}

func NewImageDeletionRetryStore(db *gorm.DB) ImageDeletionRetryStore {
	return &imageDeletionRetryStore{db: db}
}

func (s *imageDeletionRetryStore) ListClinicIDsAfter(ctx context.Context, afterClinicID uint64, limit int) ([]uint64, error) {
	if s.db == nil {
		return nil, errors.New("trimming image deletion retry store requires a database")
	}
	if limit <= 0 || limit > imageDeletionMaxClinicsPerRun {
		limit = imageDeletionMaxClinicsPerRun
	}
	var ids []uint64
	err := s.db.WithContext(ctx).Raw(
		"SELECT id FROM clinics WHERE id > ? ORDER BY id ASC LIMIT ?", afterClinicID, limit,
	).Scan(&ids).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list clinics for trimming image deletion retry")
	}
	return ids, nil
}

func (s *imageDeletionRetryStore) ListPendingImageKeys(ctx context.Context, clinicID uint64, limit int) ([]string, error) {
	if s.db == nil {
		return nil, errors.New("trimming image deletion retry store requires a database")
	}
	if limit <= 0 || limit > imageDeletionMaxKeysPerRun {
		limit = imageDeletionMaxKeysPerRun
	}
	var keys []string
	err := pendingCourseImageDeletionQuery(s.db.WithContext(ctx), clinicID, limit).Scan(&keys).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list pending trimming image deletions")
	}
	return keys, nil
}

func (s *imageDeletionRetryStore) RetryImageDeletion(
	ctx context.Context,
	clinicID uint64,
	stored string,
	objects ImageDeletionObjectStore,
) (bool, error) {
	if s.db == nil {
		return false, errors.New("trimming image deletion retry store requires a database")
	}
	if objects == nil {
		return false, errors.New("trimming image deletion retry store requires an object store")
	}
	if clinicID == 0 {
		return false, errors.New("trimming image deletion retry requires a positive clinic id")
	}
	key, err := ownedImageKey(clinicID, stored)
	if err != nil {
		return false, err
	}
	lockName := fmt.Sprintf("trimming-image:%d:%s", clinicID, key.value)
	providerCtx, providerCancel := context.WithTimeout(ctx, imageDeletionProviderTimeout)
	defer providerCancel()
	txCtx, txCancel := context.WithTimeout(
		persistence.WithTxValue(context.WithoutCancel(ctx), nil),
		imageDeletionProviderTimeout+imageDeletionTxGrace)
	defer txCancel()
	var attempted bool
	var deleteErr error
	txErr := s.db.WithContext(txCtx).Transaction(func(tx *gorm.DB) error {
		var acquired bool
		if err := tx.Raw(
			"SELECT pg_try_advisory_xact_lock(hashtextextended(?, 0))", lockName,
		).Scan(&acquired).Error; err != nil {
			return err
		}
		if !acquired {
			return nil
		}
		var pending bool
		if err := tx.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM trimming_image_deletions
				WHERE clinic_id = ? AND object_key = ? AND storage_deleted_at IS NULL
			)`, clinicID, key.value,
		).Scan(&pending).Error; err != nil {
			return err
		}
		if !pending {
			return nil
		}
		var eligible bool
		if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM trimming_course_image_uploads u JOIN trimming_image_deletions d ON d.clinic_id=u.clinic_id AND d.object_key=u.object_key AND d.actor_staff_id=u.uploader_staff_id WHERE u.clinic_id=? AND u.object_key=? AND u.uploaded_at IS NOT NULL AND u.submitted_at IS NULL)`, clinicID, key.value).Scan(&eligible).Error; err != nil {
			return err
		}
		if !eligible {
			return nil
		}
		detail := &appointmentTrimmingDetailRepository{db: s.db}
		used, err := detail.ImageKeyInUse(persistence.WithTxValue(txCtx, tx), clinicID, key.value, 0)
		if err != nil {
			return err
		}
		if used {
			// Retain the unsafe row, but rotate it behind unattempted eligible work.
			// This is scheduling metadata, never a successful deletion acknowledgement.
			return tx.Exec("UPDATE trimming_image_deletions SET last_attempt_at=NOW() WHERE clinic_id=? AND object_key=? AND storage_deleted_at IS NULL", clinicID, key.value).Error
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		attempted = true
		if err := objects.Delete(providerCtx, key.value); err != nil {
			deleteErr = err
			return tx.Exec(`
				UPDATE trimming_image_deletions SET last_attempt_at = NOW()
				WHERE clinic_id = ? AND object_key = ? AND storage_deleted_at IS NULL`,
				clinicID, key.value,
			).Error
		}
		res := tx.Exec(`
			UPDATE trimming_image_deletions
			SET last_attempt_at = NOW(), storage_deleted_at = NOW()
			WHERE clinic_id = ? AND object_key = ? AND storage_deleted_at IS NULL`,
			clinicID, key.value,
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return fmt.Errorf("trimming image deletion completion updated %d rows", res.RowsAffected)
		}
		return nil
	})
	if txErr != nil && deleteErr != nil {
		return true, errors.Join(deleteErr,
			apperrors.Wrap(txErr, "failed to persist trimming image deletion attempt"))
	}
	if txErr != nil {
		return false, apperrors.Wrap(txErr, "failed to retry trimming image deletion")
	}
	if deleteErr != nil {
		return true, deleteErr
	}
	return attempted, nil
}
