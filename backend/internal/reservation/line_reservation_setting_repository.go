package reservation

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LineReservationSettingRepository は予約基本設定のデータアクセスインターフェース
type LineReservationSettingRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	// FindAll は全クリニックの設定を返す（Webhook署名検証用）。
	FindAll(ctx context.Context) ([]model.LineReservationSetting, error)
	Save(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error
}

type lineReservationSettingRepository struct{ db *gorm.DB }

func NewLineReservationSettingRepository(db *gorm.DB) LineReservationSettingRepository {
	return &lineReservationSettingRepository{db: db}
}

func (r *lineReservationSettingRepository) FindAll(ctx context.Context) ([]model.LineReservationSetting, error) {
	var settings []model.LineReservationSetting
	if err := r.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, apperrors.FromGORM(err, "line_reservation_setting", "")
	}
	return settings, nil
}

func (r *lineReservationSettingRepository) FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error) {
	var setting model.LineReservationSetting
	err := r.db.WithContext(ctx).Scopes(persistence.ClinicScope(clinicID)).First(&setting).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "line_reservation_setting", "clinic")
	}
	return &setting, nil
}

func (r *lineReservationSettingRepository) Save(ctx context.Context, clinicID uint64, setting *model.LineReservationSetting) error {
	db := r.db.WithContext(ctx)
	// Capture intent before Create: gorm default:true omits zero bools from INSERT.
	wantShowNoStaff := setting.ShowNoStaffOption
	err := db.
		Scopes(persistence.ClinicScope(clinicID)).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "clinic_id"}},
			DoUpdates: clause.AssignmentColumns(lineReservationSettingUpdatableColumns()),
		}).
		Create(setting).Error
	if err != nil {
		return apperrors.FromGORM(err, "line_reservation_setting", "")
	}
	if !wantShowNoStaff {
		if err := db.Scopes(persistence.ClinicScope(clinicID)).
			Model(&model.LineReservationSetting{}).
			Where("clinic_id = ?", clinicID).
			Update("show_no_staff_option", false).Error; err != nil {
			return apperrors.FromGORM(err, "line_reservation_setting", fmt.Sprintf("clinic:%d", clinicID))
		}
		setting.ShowNoStaffOption = false
	}
	return nil
}

func lineReservationSettingUpdatableColumns() []string {
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
