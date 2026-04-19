package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicHolidayRepository は個別休診日のデータアクセスインターフェース
type ClinicHolidayRepository interface {
	FindByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	Upsert(ctx context.Context, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
	Delete(ctx context.Context, clinicID uint64, date time.Time) error
}

type clinicHolidayRepository struct{ db *gorm.DB }

// NewClinicHolidayRepository はClinicHolidayRepositoryを初期化して返す
func NewClinicHolidayRepository(db *gorm.DB) ClinicHolidayRepository {
	return &clinicHolidayRepository{db: db}
}

func (r *clinicHolidayRepository) FindByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error) {
	var holidays []model.ClinicHoliday
	q := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
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

func (r *clinicHolidayRepository) Upsert(ctx context.Context, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	// map を使用してゼロ値（Reason=""）も確実に更新する
	var existing model.ClinicHoliday
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND date = ?", holiday.ClinicID, holiday.Date).
		First(&existing).Error

	if err != nil {
		// レコードなし → 新規作成
		if err := r.db.WithContext(ctx).Create(holiday).Error; err != nil {
			return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
		}
		return holiday, nil
	}

	// 既存レコード → reason を更新
	if err := r.db.WithContext(ctx).
		Model(&model.ClinicHoliday{}).
		Where("id = ?", existing.ID).
		Update("reason", holiday.Reason).Error; err != nil {
		return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
	}
	existing.Reason = holiday.Reason
	*holiday = existing
	return holiday, nil
}

func (r *clinicHolidayRepository) Delete(ctx context.Context, clinicID uint64, date time.Time) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("date = ?", date.Format("2006-01-02")).
		Delete(&model.ClinicHoliday{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "clinic_holiday", date.Format("2006-01-02"))
	}
	if result.RowsAffected == 0 {
		return apperrors.FromGORM(gorm.ErrRecordNotFound, "clinic_holiday", date.Format("2006-01-02"))
	}
	return nil
}
