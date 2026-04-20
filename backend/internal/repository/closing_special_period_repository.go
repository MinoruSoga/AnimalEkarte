package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClosingSpecialPeriodRepository は特別締め期間のデータアクセスインターフェース
type ClosingSpecialPeriodRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error)
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error)
	Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error)
	UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	HasOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error)
}

type closingSpecialPeriodRepository struct{ db *gorm.DB }

// NewClosingSpecialPeriodRepository は ClosingSpecialPeriodRepository を初期化して返す
func NewClosingSpecialPeriodRepository(db *gorm.DB) ClosingSpecialPeriodRepository {
	return &closingSpecialPeriodRepository{db: db}
}

func (r *closingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
	var periods []model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("start_date ASC").
		Find(&periods).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return periods, nil
}

func (r *closingSpecialPeriodRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		First(&p).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id))
	}
	return &p, nil
}

func (r *closingSpecialPeriodRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClosingSpecialPeriod, error) {
	var p model.ClosingSpecialPeriod
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("start_date <= ? AND end_date >= ?", date, date).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find closing special period by date")
	}
	return &p, nil
}

func (r *closingSpecialPeriodRepository) Create(ctx context.Context, p *model.ClosingSpecialPeriod) (*model.ClosingSpecialPeriod, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", "")
	}
	return p, nil
}

func (r *closingSpecialPeriodRepository) UpdateFields(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.ClosingSpecialPeriod, error) {
	result := r.db.WithContext(ctx).
		Model(&model.ClosingSpecialPeriod{}).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "closing_special_period", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("closing_special_period", fmt.Sprintf("%d", id))
	}
	var p model.ClosingSpecialPeriod
	if err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		First(&p).Error; err != nil {
		return nil, apperrors.FromGORM(err, "closing_special_period", fmt.Sprintf("%d", id))
	}
	return &p, nil
}

func (r *closingSpecialPeriodRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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

func (r *closingSpecialPeriodRepository) HasOverlap(ctx context.Context, clinicID uint64, startDate, endDate time.Time, excludeID *uint64) (bool, error) {
	q := r.db.WithContext(ctx).
		Model(&model.ClosingSpecialPeriod{}).
		Scopes(clinicScope(clinicID)).
		Where("start_date <= ? AND end_date >= ?", endDate, startDate)
	if excludeID != nil {
		q = q.Where("id != ?", *excludeID)
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false, apperrors.Wrap(err, "failed to check overlap")
	}
	return count > 0, nil
}
