// Package closingspecialperiod owns closing_special_periods data access within the clinic domain.
package closingspecialperiod

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository is the data access interface for special closing periods.
type Repository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error)
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
	Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
}

type repository struct{ db *gorm.DB }

// New constructs a Repository.
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	var periods []model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("start_date ASC").
		Find(&periods).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return periods, nil
}

func (r *repository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id))
	}
	return &p, nil
}

func (r *repository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return p, nil
}

func (r *repository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error) {
	if err := repohelpers.UpdateScopedByID(ctx, r.db, &model.ClosingSpecialPeriod{}, "closing_special_period", clinicID, id, fields); err != nil {
		return nil, err
	}
	var p model.ClosingSpecialPeriod
	if err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&p).Error; err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id))
	}
	return &p, nil
}

func (r *repository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
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

func (r *repository) CheckOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error) {
	q := r.db.WithContext(ctx).
		Model(&model.ClosingSpecialPeriod{}).
		Scopes(repohelpers.ClinicScope(clinicID)).
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
