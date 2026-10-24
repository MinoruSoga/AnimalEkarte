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
	// The fields exist only to detect that a client still sends the key: the key being present with any
	// value (non-empty, empty string, whitespace, null) is rejected with 400 (rejectLegacyLineCredentials)
	// instead of being silently dropped (EMR-549 Q-U5-1). Only a request that omits the keys is accepted;
	// the admin form no longer sends them.
	LineChannelID     jsonKeyPresence `json:"line_channel_id"`
	LineChannelSecret jsonKeyPresence `json:"line_channel_secret"`
	LiffID            jsonKeyPresence `json:"liff_id"`
	LineAccessToken   jsonKeyPresence `json:"line_access_token"`
}

// jsonKeyPresence は JSON キーが body に存在したことだけを記録する。値は読まず保持しない
// （資格情報を struct・ログ・エラーに残さない）。pointer と違い null リテラルでも UnmarshalJSON が
// 呼ばれるため、「省略」と「null」を区別できる。キー名の照合は encoding/json に従い大文字小文字を区別しない。
type jsonKeyPresence bool

// UnmarshalJSON はキーが存在したことを記録する。
func (p *jsonKeyPresence) UnmarshalJSON([]byte) error {
	*p = true
	return nil
}

// rejectLegacyLineCredentials は dropped 済みの LINE 資格情報キーが 1 つでも送られた場合に 400 を返す。
// 値は問わない（空文字・空白のみ・null も拒否）。存在判定は UnmarshalJSON が呼ばれたか（= キーが body にあるか）で行い、
// 値の型には依存しない。値そのものはエラーに含めない。
func (r *upsertLineReservationSettingRequest) rejectLegacyLineCredentials() error {
	var fields []string
	for _, f := range []struct {
		name    string
		present jsonKeyPresence
	}{
		{"line_channel_id", r.LineChannelID},
		{"line_channel_secret", r.LineChannelSecret},
		{"liff_id", r.LiffID},
		{"line_access_token", r.LineAccessToken},
	} {
		if f.present {
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
