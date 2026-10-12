package trimming

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
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
	err := s.db.WithContext(ctx).Raw(`
		SELECT object_key FROM trimming_image_deletions
		WHERE clinic_id = ? AND storage_deleted_at IS NULL
		ORDER BY last_attempt_at ASC NULLS FIRST, object_key ASC
		LIMIT ?`, clinicID, limit,
	).Scan(&keys).Error
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
	var attempted bool
	var deleteErr error
	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		attempted = true
		if err := objects.Delete(ctx, key.value); err != nil {
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
	if txErr != nil {
		return false, apperrors.Wrap(txErr, "failed to retry trimming image deletion")
	}
	if deleteErr != nil {
		return true, deleteErr
	}
	return attempted, nil
}
