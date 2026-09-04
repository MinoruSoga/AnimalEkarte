package clinic

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

type clinicHolidayRepository struct{ db *gorm.DB }

func (r *clinicHolidayRepository) FindAllByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	var holidays []model.ClinicHoliday
	q := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
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

func (r *clinicHolidayRepository) Save(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	// 同一日の追加は上書きせず unique 違反にする（BUG-015）。理由変更は削除してから再追加する。
	// INSERT に Scopes(WHERE) は効かないため、書き込み先は arg の clinicID を正とする（POC-09）。
	holiday.ClinicID = clinicID
	err := r.db.WithContext(ctx).Create(holiday).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format(time.DateOnly))
	}
	return holiday, nil
}

func (r *clinicHolidayRepository) Delete(ctx context.Context, clinicID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
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

func (r *clinicHolidayRepository) FindByDate(ctx context.Context, clinicID uint64, date time.Time) (*model.ClinicHoliday, error) {
	var holiday model.ClinicHoliday
	result := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("date = ?", date.Format(time.DateOnly)).
		First(&holiday)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "clinic_holiday", date.Format(time.DateOnly))
	}
	return &holiday, nil
}
