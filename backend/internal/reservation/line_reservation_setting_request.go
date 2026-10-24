package reservation

import (
	"encoding/json"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// upsertLineReservationSettingRequest is the HTTP body for LINE reservation settings upsert.
// Binding tags enforce guidelines:151 (enum, range, length, email) so invalid values
// never reach persistence or available-dates make(cap) (RSV-04 / U-X02-RESERVATION-SETTINGS).
type upsertLineReservationSettingRequest struct {
	Status                  string         `json:"status" binding:"required,oneof=running stopped"`
	HeaderText              string         `json:"header_text" binding:"max=2000"`
	ReservationNotice       string         `json:"reservation_notice" binding:"max=10000"`
	CancelNotice            string         `json:"cancel_notice" binding:"max=10000"`
	PrivacyPolicy           string         `json:"privacy_policy" binding:"max=100000"`
	ClosedWeekdays          jsonRawOrEmpty `json:"closed_weekdays"`
	ClosedDates             jsonRawOrEmpty `json:"closed_dates"`
	NationalHolidayClosed   bool           `json:"national_holiday_closed"`
	BusinessHours           jsonRawOrEmpty `json:"business_hours"`
	BusinessHoursByWeekday  jsonRawOrEmpty `json:"business_hours_by_weekday"`
	BreakHours              jsonRawOrEmpty `json:"break_hours"`
	DailyLimit              *int           `json:"daily_limit" binding:"omitempty,min=0,max=100000"`
	MonthlyLimit            *int           `json:"monthly_limit" binding:"omitempty,min=0,max=100000"`
	BookingWindowMaxDays    int            `json:"booking_window_max_days" binding:"min=0,max=366"`
	BookingWindowMinDays    int            `json:"booking_window_min_days" binding:"min=0,max=366"`
	CalendarMonths          int            `json:"calendar_months" binding:"min=0,max=12"`
	PhoneNumber             string         `json:"phone_number" binding:"max=32"`
	NotificationEmail       string         `json:"notification_email" binding:"omitempty,email,max=254"`
	RequestExample          string         `json:"request_example" binding:"max=2000"`
	TimeSlotMode            string         `json:"time_slot_mode" binding:"required,oneof=minimize_gaps allow_gaps"`
	TimeSlotIntervalMinutes int            `json:"time_slot_interval_minutes" binding:"min=1,max=1440"`
	NoStaffMode             string         `json:"no_staff_mode" binding:"required,oneof=first_available top_priority"`
	// ShowNoStaffOption is *bool so JSON binding can distinguish omitted / false / true.
	// Omitted (nil) resolves to true in toServiceInput.
	ShowNoStaffOption *bool          `json:"show_no_staff_option"`
	AdditionalFields  jsonRawOrEmpty `json:"additional_fields"`
	// EMR-504: LINE credentials (channel ID / channel secret / LIFF ID / access token) are not part of
	// the reservation setting; their only store is clinic_integrations (lstep-settings API).
	// The pointers exist only to detect that a client still sends a value: a non-empty value (after
	// trimming whitespace) is rejected with 400 (rejectLegacyLineCredentials) instead of being silently
	// dropped. Absent, null and empty string are accepted and nothing is stored, because the admin form
	// always sends line_channel_id and liff_id.
	LineChannelID     *string `json:"line_channel_id"`
	LineChannelSecret *string `json:"line_channel_secret"`
	LiffID            *string `json:"liff_id"`
	LineAccessToken   *string `json:"line_access_token"`
}

// rejectLegacyLineCredentials は dropped 済みの LINE 資格情報フィールドに空でない値（前後の空白を除く）が
// 送られた場合に 400 を返す。省略・null・空文字は受理し、何も保存しない（管理画面フォームが
// line_channel_id / liff_id を常に送るため）。値そのものはエラーに含めない。
func (r *upsertLineReservationSettingRequest) rejectLegacyLineCredentials() error {
	var fields []string
	for _, f := range []struct {
		name  string
		value *string
	}{
		{"line_channel_id", r.LineChannelID},
		{"line_channel_secret", r.LineChannelSecret},
		{"liff_id", r.LiffID},
		{"line_access_token", r.LineAccessToken},
	} {
		if f.value != nil && strings.TrimSpace(*f.value) != "" {
			fields = append(fields, f.name)
		}
	}
	if len(fields) == 0 {
		return nil
	}
	return apperrors.WrapInvalidInput(strings.Join(fields, ", ") +
		" はLINE予約設定では受け付けません。LINE連携設定（/lstep-settings）で設定してください")
}

func (r *upsertLineReservationSettingRequest) toServiceInput() *UpsertLineReservationSettingInput {
	return &UpsertLineReservationSettingInput{
		Status:                  r.Status,
		HeaderText:              r.HeaderText,
		ReservationNotice:       r.ReservationNotice,
		CancelNotice:            r.CancelNotice,
		PrivacyPolicy:           r.PrivacyPolicy,
		ClosedWeekdays:          r.ClosedWeekdays,
		ClosedDates:             r.ClosedDates,
		NationalHolidayClosed:   r.NationalHolidayClosed,
		BusinessHours:           r.BusinessHours,
		BusinessHoursByWeekday:  r.BusinessHoursByWeekday,
		BreakHours:              r.BreakHours,
		DailyLimit:              r.DailyLimit,
		MonthlyLimit:            r.MonthlyLimit,
		BookingWindowMaxDays:    r.BookingWindowMaxDays,
		BookingWindowMinDays:    r.BookingWindowMinDays,
		CalendarMonths:          r.CalendarMonths,
		PhoneNumber:             r.PhoneNumber,
		NotificationEmail:       r.NotificationEmail,
		RequestExample:          r.RequestExample,
		TimeSlotMode:            r.TimeSlotMode,
		TimeSlotIntervalMinutes: r.TimeSlotIntervalMinutes,
		NoStaffMode:             r.NoStaffMode,
		ShowNoStaffOption:       resolveBoolDefaultTrue(r.ShowNoStaffOption),
		AdditionalFields:        r.AdditionalFields,
	}
}

// resolveBoolDefaultTrue maps optional bool presence: nil → true, otherwise the pointed value.
func resolveBoolDefaultTrue(v *bool) bool {
	if v == nil {
		return true
	}
	return *v
}

// jsonRawOrEmpty は JSON フィールドを生の JSON として保持するためのエイリアス。
// 標準の []byte ではなく json.RawMessage を使うのは、encoding/json が []byte フィールドを
// base64 エンコード文字列として扱うため（BUG-LINE-007）。json.RawMessage は MarshalJSON /
// UnmarshalJSON が実装されており任意の JSON オブジェクト・配列をそのまま保持できる。
// 実体は []byte なので service 層の []byte 引数にはそのまま渡せる。
type jsonRawOrEmpty = json.RawMessage
