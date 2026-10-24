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
	// FindWebhookRouteByLineBotUserID は webhook destination から clinic を引く（credential は返さない）。
	// 旧 line_bot_user_id は DROP 済みで置換先の routing 保存は未決定（EMR-504 Q7）のため、
	// 決定されるまで apperrors.CodeLineWebhookRoutingPending（422）で fail-closed する。
	FindWebhookRouteByLineBotUserID(ctx context.Context, lineBotUserID string) (clinicID uint64, err error)
	// FindIntegrationValue は clinic_integrations(service='lstep') の許可済みキー（liff_id のみ）の保存値を返す。
	// 行が無ければ ("", nil)。
	FindIntegrationValue(ctx context.Context, clinicID uint64, keyName string) (string, error)
	// FindLineChannelAccessTokenCiphertext は予約の完了・キャンセル通知（LINE Push）専用に、
	// clinic_integrations(service='lstep') の line_channel_access_token の保存値を暗号文のまま返す。
	// 復号は呼び出し側（composition が注入する、lstep 連携設定と同じ cipher）。行が無ければ ("", nil)。
	// FindIntegrationValue の許可キーは広げず、秘密の読み取りはこの専用メソッドだけに閉じる。
	FindLineChannelAccessTokenCiphertext(ctx context.Context, clinicID uint64) (string, error)
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

// FindWebhookRouteByLineBotUserID は routing 保存が決定されるまで常に fail-closed する。
// 旧実装は line_reservation_settings の line_bot_user_id / line_channel_secret 列を SELECT していたが
// いずれも DROP 済みで、置換先のキー・一意 index は 446/466 の DDL 後に決める。
func (r *lineReservationSettingRepository) FindWebhookRouteByLineBotUserID(
	_ context.Context,
	_ string,
) (uint64, error) {
	return 0, apperrors.WrapUnprocessableWithCode(
		apperrors.CodeLineWebhookRoutingPending,
		"LINE webhook の宛先（destination）から医院を引く保存先がまだ決定されていません",
	)
}

// lineIntegrationReadableKeys は FindIntegrationValue が返してよい clinic_integrations のキー
// （LIFF ID のみ・非 secret。汎用の秘密 reader にしない）。
var lineIntegrationReadableKeys = map[string]struct{}{
	model.IntegrationKeyLiffID: {},
}

func (r *lineReservationSettingRepository) FindIntegrationValue(ctx context.Context, clinicID uint64, keyName string) (string, error) {
	if _, ok := lineIntegrationReadableKeys[keyName]; !ok {
		return "", apperrors.WrapInternalServerError("clinic integration key is not readable from reservation: " + keyName)
	}
	return r.findLstepIntegrationValue(ctx, clinicID, keyName)
}

// FindLineChannelAccessTokenCiphertext は通知 Push 専用の読み取り。トークンは暗号文のまま返す。
func (r *lineReservationSettingRepository) FindLineChannelAccessTokenCiphertext(ctx context.Context, clinicID uint64) (string, error) {
	return r.findLstepIntegrationValue(ctx, clinicID, model.IntegrationKeyLineChannelAccessToken)
}

// findLstepIntegrationValue は clinic_integrations(service='lstep') の1キーを clinic で絞って返す。
// 行が無ければ ("", nil)。キーは呼び出し側（許可キーを固定した上記2メソッド）が決める。
func (r *lineReservationSettingRepository) findLstepIntegrationValue(ctx context.Context, clinicID uint64, keyName string) (string, error) {
	var records []model.ClinicIntegration
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("service = ? AND key_name = ?", model.IntegrationServiceLstep, keyName).
		Limit(1).
		Find(&records).Error
	if err != nil {
		return "", apperrors.FromGORM(err, "clinic_integration", keyName)
	}
	if len(records) == 0 {
		return "", nil
	}
	return records[0].KeyValue, nil
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
		// Ensure the clinic row exists without rewriting any column on conflict. DoUpdates column set is intentionally empty here — full intended state
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

		// Force every updatable column, including explicit zeros/false. Does not select id.
		// Same column set as the former OnConflict DoUpdates.
		if err := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Model(&model.LineReservationSetting{}).
			Where("clinic_id = ?", clinicID).
			Select(lineReservationSettingUpdatableColumns()).
			Updates(&intended).Error; err != nil {
			return apperrors.FromGORM(err, "line_reservation_setting", fmt.Sprintf("clinic:%d", clinicID))
		}

		// ON CONFLICT DO NOTHING does not RETURNING on conflict, so Create leaves setting.ID
		// at 0 on the update path. Read identity/timestamps back (not write them) so RSV-03
		// callers can return a non-zero id without a post-commit re-fetch.
		var persisted struct {
			ID        uint64
			CreatedAt time.Time
		}
		if err := tx.
			Scopes(persistence.ClinicScope(clinicID)).
			Model(&model.LineReservationSetting{}).
			Select("id", "created_at").
			Where("clinic_id = ?", clinicID).
			Take(&persisted).Error; err != nil {
			return apperrors.FromGORM(err, "line_reservation_setting", fmt.Sprintf("clinic:%d", clinicID))
		}

		// Reflect persisted intent for callers (Create may have mutated zeros on setting).
		*setting = intended
		setting.ID = persisted.ID
		setting.CreatedAt = persisted.CreatedAt
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
		// EMR-504: line_channel_id / line_channel_secret / liff_id / line_access_token /
		// line_bot_user_id were dropped from line_reservation_settings (001 final). LINE
		// credentials live only in clinic_integrations (lstep-settings API).
		"updated_at",
	}
}
