package lstep

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type LstepSyncErrorCounterRepository interface {
	// IncrementFailure atomically increments the failure count and returns the new count.
	IncrementFailure(ctx context.Context, clinicID, ownerID uint64) (int, error)
	// ResetFailure deletes the counter record (row-absent = zero failures).
	ResetFailure(ctx context.Context, clinicID, ownerID uint64) error
	// FindByOwner returns the counter record; returns NotFound error when absent.
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.LstepSyncErrorCounter, error)
}

type lstepSyncErrorCounterRepository struct{ db *gorm.DB }

func NewLstepSyncErrorCounterRepository(db *gorm.DB) LstepSyncErrorCounterRepository {
	return &lstepSyncErrorCounterRepository{db: db}
}

func (r *lstepSyncErrorCounterRepository) IncrementFailure(ctx context.Context, clinicID, ownerID uint64) (int, error) {
	var newCount int
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO lstep_sync_error_counters (clinic_id, owner_id, failure_count, created_at, updated_at)
		VALUES (?, ?, 1, now(), now())
		ON CONFLICT (clinic_id, owner_id) DO UPDATE
		SET failure_count = lstep_sync_error_counters.failure_count + 1,
		    updated_at    = now()
		RETURNING failure_count`, clinicID, ownerID).Scan(&newCount).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "lstep_sync_error_counter", fmt.Sprintf("clinic=%d,owner=%d", clinicID, ownerID))
	}
	return newCount, nil
}

func (r *lstepSyncErrorCounterRepository) ResetFailure(ctx context.Context, clinicID, ownerID uint64) error {
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		Delete(&model.LstepSyncErrorCounter{}).Error
	if err != nil {
		return apperrors.FromGORM(err, "lstep_sync_error_counter", fmt.Sprintf("clinic=%d,owner=%d", clinicID, ownerID))
	}
	return nil
}

func (r *lstepSyncErrorCounterRepository) FindByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.LstepSyncErrorCounter, error) {
	var counter model.LstepSyncErrorCounter
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND owner_id = ?", clinicID, ownerID).
		First(&counter).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lstep_sync_error_counter", fmt.Sprintf("clinic=%d,owner=%d", clinicID, ownerID))
	}
	return &counter, nil
}
