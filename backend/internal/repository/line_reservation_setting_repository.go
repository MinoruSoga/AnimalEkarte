package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LineReservationSettingRepository は予約基本設定のデータアクセスインターフェース
type LineReservationSettingRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	Upsert(ctx context.Context, setting *model.LineReservationSetting) error
}

type reservationSettingRepository struct{ db *gorm.DB }

func NewLineReservationSettingRepository(db *gorm.DB) LineReservationSettingRepository {
	return &reservationSettingRepository{db: db}
}

func (r *reservationSettingRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	var setting model.LineReservationSetting
	err := r.db.WithContext(ctx).Where("clinic_id = ?", clinicID).First(&setting).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_setting", "clinic")
	}
	return &setting, nil
}

func (r *reservationSettingRepository) Upsert(ctx context.Context, setting *model.LineReservationSetting) error {
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns(reservationSettingUpdatableColumns()),
		}).
		Create(setting).Error
	if err != nil {
		return apperrors.FromGORM(err, "reservation_setting", "")
	}
	return nil
}

func reservationSettingUpdatableColumns() []string {
	return []string{
		"status",
		"header_text",
		"reservation_notice",
		"cancel_notice",
		"privacy_policy",
		"closed_weekdays",
		"closed_dates",
		"national_holiday_closed",
		"business_hours",
		"business_hours_by_weekday",
		"break_hours",
		"daily_limit",
		"monthly_limit",
		"booking_window_max_days",
		"booking_window_min_days",
		"calendar_months",
		"phone_number",
		"notification_email",
		"request_example",
		"time_slot_mode",
		"time_slot_interval_minutes",
		"no_staff_mode",
		"show_no_staff_option",
		"additional_fields",
		"line_channel_id",
		"line_channel_secret",
		"liff_id",
		"line_access_token",
		"updated_at",
	}
}
