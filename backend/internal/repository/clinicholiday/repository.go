// Package clinicholiday owns clinic_holidays data access (BE8-4 batch11 — leaf domain).
package clinicholiday

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository は個別休診日のデータアクセスインターフェース
type Repository interface {
	FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error)
	FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	Delete(ctx context.Context, clinicID uint64, date time.Time) error
}

type repository struct{ db *gorm.DB }

// New は Repository を初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	var holidays []model.ClinicHoliday
	q := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Order("date ASC")

	if yearMonth != "" {
		q = q.Where("date >= ?::date AND date < (?::date + INTERVAL '1 month')",
			yearMonth+"-01", yearMonth+"-01")
	}

	if err := q.Find(&holidays).Error; err != nil {
		return nil, apperrors.FromGORM(err, "clinic_holiday", yearMonth)
	}
	return holidays, nil
}

func (r *repository) Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	// (clinic_id, date) のユニーク制約を利用してアトミックな UPSERT を実施する。
	// 手動の First→Create/Update パターンはレースコンディションを持つため clause.OnConflict を使用する。
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"reason", "updated_at"}),
		}).
		Create(holiday).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format(time.DateOnly))
	}
	return holiday, nil
}

func (r *repository) Delete(ctx context.Context, clinicID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("date = ?", date.Format(time.DateOnly)).
		Delete(&model.ClinicHoliday{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic_holiday", date.Format(time.DateOnly))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("clinic_holiday", date.Format(time.DateOnly))
	}
	return nil
}

func (r *repository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	var holiday model.ClinicHoliday
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("date = ?", date.Format(time.DateOnly)).
		First(&holiday)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "clinic_holiday", date.Format(time.DateOnly))
	}
	return &holiday, nil
}
