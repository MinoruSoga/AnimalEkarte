package reservation

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LineReservationSettingRepository は予約基本設定のデータアクセスインターフェース
type LineReservationSettingRepository interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	// FindWebhookRouteByLineBotUserID は webhook routing metadata のみを返し、credential
	// payload は返さない。line_bot_user_id が空の行はマッチさせない。
	FindWebhookRouteByLineBotUserID(ctx context.Context, lineBotUserID string) (clinicID uint64, legacyCredentialPresent bool, err error)
	// FindAll は全クリニックの設定を返す（管理用途）。Webhook 署名検証では使わない。
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

// FindWebhookRouteByLineBotUserID は credential 値をSELECTせず、routing identity と
// legacy credential presence metadata のみを返す。
func (r *lineReservationSettingRepository) FindWebhookRouteByLineBotUserID(
	ctx context.Context,
	lineBotUserID string,
) (uint64, bool, error) {
	if lineBotUserID == "" {
		return 0, false, apperrors.FromGORM(gorm.ErrRecordNotFound, "line_reservation_setting", "line_bot_user_id")
	}
	var route struct {
		ClinicID                uint64
		LegacyCredentialPresent bool
	}
	err := r.db.WithContext(ctx).
		Model(&model.LineReservationSetting{}).
		Select("clinic_id, (line_channel_secret <> '') AS legacy_credential_present").
		Where("line_bot_user_id = ? AND line_bot_user_id <> ''", lineBotUserID).
		Take(&route).Error
	if err != nil {
		return 0, false, apperrors.FromGORM(err, "line_reservation_setting", "line_bot_user_id")
	}
	return route.ClinicID, route.LegacyCredentialPresent, nil
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
	// Tenant authority is the clinicID argument — never trust a mismatched setting.ClinicID.
	setting.ClinicID = clinicID

	// Snapshot before Create: GORM 1.31 ConvertToCreateValues replaces zero-valued
	// default-tagged fields with DefaultValueInterface in-place on the struct AND in
	// the INSERT row. Select() alone does not prevent that, so excluded.* / the struct
	// cannot carry explicit 0/false through Create (BUG-030).
	intended := *setting
	intended.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Ensure the clinic row exists without rewriting credentials/routing identity on
		// conflict. DoUpdates column set is intentionally empty here — full intended state
		// is written next via Select+Updates over lineReservationSettingUpdatableColumns only.
		if err := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "clinic_id"}},
				DoNothing: true,
			}).
			Create(setting).Error; err != nil {
			return apperrors.FromGORM(err, "line_reservation_setting", "")
		}

		// Force every updatable column, including explicit zeros/false. Does not select
		// line_channel_secret, line_bot_user_id, or id — provisioned routing/credentials
		// survive UI Save (SEC-CS-F05-R1). Same column set as the former OnConflict DoUpdates.
		if err := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Model(&model.LineReservationSetting{}).
			Where("clinic_id = ?", clinicID).
			Select(lineReservationSettingUpdatableColumns()).
			Updates(&intended).Error; err != nil {
			return apperrors.FromGORM(err, "line_reservation_setting", fmt.Sprintf("clinic:%d", clinicID))
		}

		// Reflect persisted intent for callers (Create may have mutated zeros on setting).
		id := setting.ID
		*setting = intended
		if id != 0 {
			setting.ID = id
		}
		return nil
	})
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
		// R-05 Phase B: line_channel_secret is excluded from Save Updates.
		// Canonical channel secret SoT is clinic_integrations. Column remains for
		// legacy_credential_present presence SELECT until inventory-zero DROP (HOLD).
		// line_bot_user_id is ops/migration-provisioned for O(1) webhook routing
		// (SEC-CS-F05-R1). Exclude from Save Updates so UI/API Save cannot wipe a
		// provisioned bot user ID with the zero value.
		"liff_id",
		"line_access_token",
		"updated_at",
	}
}
