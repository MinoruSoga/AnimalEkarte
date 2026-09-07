package clinic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type closingSpecialPeriodRepository struct{ db *gorm.DB }

func (r *closingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	var periods []model.ClosingSpecialPeriod
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("start_date ASC").
		Find(&periods).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return periods, nil
}

func (r *closingSpecialPeriodRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id))
	}
	return &p, nil
}

func (r *closingSpecialPeriodRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_date <= ? AND end_date >= ?", date, date).
		First(&p).Error
	if err != nil {
		wrapped := apperrors.FromGORM(err, "closing_special_period", date.Format(time.DateOnly))
		if errors.Is(wrapped, apperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, wrapped
	}
	return &p, nil
}

func (r *closingSpecialPeriodRepository) Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if err := persistence.DBOrTx(ctx, r.db).Create(p).Error; err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return p, nil
}

func closingSpecialPeriodClinicLockKey(clinicID uint64) string {
	return fmt.Sprintf("closing_special_period:clinic:%d", clinicID)
}

// CreateCheckingOverlap serializes CheckOverlap + Create under a clinic-scoped advisory lock (POC-05 / X-05).
func (r *closingSpecialPeriodRepository) CreateCheckingOverlap(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if p == nil {
		return nil, apperrors.WrapInvalidInput("closing special period is required")
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", closingSpecialPeriodClinicLockKey(p.ClinicID)).Error; err != nil {
			return apperrors.Wrap(err, "failed to acquire closing special period clinic lock")
		}
		overlap, err := r.CheckOverlap(txCtx, p.ClinicID, p.StartDate, p.EndDate, nil)
		if err != nil {
			return apperrors.Wrap(err, "failed to check period overlap")
		}
		if overlap {
			return apperrors.WrapConflict("期間が他の特別期間と重複しています")
		}
		if err := tx.WithContext(txCtx).Create(p).Error; err != nil {
			return apperrors.FromGORM(err, "closing_special_period", "")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Update は update+reload を同一 transaction に収める（POC-02 / X-01）。
// reload 失敗時は write をロールバックし、commit 済み成功を 5xx へ反転しない。
func (r *closingSpecialPeriodRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	var loaded *model.ClosingSpecialPeriod
	err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := r.update(txCtx, clinicID, id, specialPeriodUpdateFields(cmd)); err != nil {
			return err
		}
		var p model.ClosingSpecialPeriod
		if err := tx.WithContext(txCtx).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&p).Error; err != nil {
			return apperrors.Wrap(apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id)), "reload closing special period after update")
		}
		loaded = &p
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update and reload closing special period")
	}
	return loaded, nil
}

func (r *closingSpecialPeriodRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	return persistence.UpdateScopedByID(ctx, persistence.DBOrTx(ctx, r.db), &model.ClosingSpecialPeriod{}, "closing_special_period", clinicID, id, fields)
}

// UpdateCheckingOverlap serializes CheckOverlap + update+reload under a clinic advisory lock (POC-05 / X-05).
func (r *closingSpecialPeriodRepository) UpdateCheckingOverlap(ctx context.Context, clinicID, id uint64, startDate, endDate time.Time, cmd UpdateSpecialPeriodInput) (*model.ClosingSpecialPeriod, error) {
	var loaded *model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := persistence.WithTxValue(ctx, tx)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", closingSpecialPeriodClinicLockKey(clinicID)).Error; err != nil {
			return apperrors.Wrap(err, "failed to acquire closing special period clinic lock")
		}
		excludeID := id
		overlap, err := r.CheckOverlap(txCtx, clinicID, startDate, endDate, &excludeID)
		if err != nil {
			return apperrors.Wrap(err, "failed to check period overlap")
		}
		if overlap {
			return apperrors.WrapConflict("期間が他の特別期間と重複しています")
		}
		if err := r.update(txCtx, clinicID, id, specialPeriodUpdateFields(cmd)); err != nil {
			return err
		}
		var p model.ClosingSpecialPeriod
		if err := tx.WithContext(txCtx).
			Scopes(persistence.ClinicScope(clinicID)).
			Where("id = ?", id).
			First(&p).Error; err != nil {
			return apperrors.Wrap(apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id)), "reload closing special period after update")
		}
		loaded = &p
		return nil
	})
	if err != nil {
		return nil, err
	}
	return loaded, nil
}

func (r *closingSpecialPeriodRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.ClosingSpecialPeriod{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "closing_special_period", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("closing_special_period", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *closingSpecialPeriodRepository) CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error) {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.ClosingSpecialPeriod{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("start_date <= ? AND end_date >= ?", endDate, startDate)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return count > 0, nil
}
