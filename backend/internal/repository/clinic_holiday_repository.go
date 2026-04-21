package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicHolidayRepository は個別休診日のデータアクセスインターフェース
type ClinicHolidayRepository interface {
	FindByYearMonth(ctx context.Context, clinicID uint64, yearMonth string) ([]model.ClinicHoliday, error)
	Upsert(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error)
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

func (r *clinicHolidayRepository) Upsert(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
	// (clinic_id, date) のユニーク制約を利用してアトミックな UPSERT を実施する。
	// 手動の First→Create/Update パターンはレースコンディションを持つため clause.OnConflict を使用する。
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"reason", "updated_at"}),
		}).
		Create(holiday).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "clinic_holiday", holiday.Date.Format("2006-01-02"))
	}
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
		return apperrors.WrapNotFound("clinic_holiday", date.Format("2006-01-02"))
	}
	return nil
}
