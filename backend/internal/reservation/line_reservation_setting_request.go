package reservation

import (
	"encoding/json"
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
	LineChannelID     string         `json:"line_channel_id" binding:"max=255"`
	// R-05 Phase B: line_channel_secret is intentionally not accepted here.
	// Canonical write owner is clinic_integrations via L-step settings API.
	// Unknown JSON key is ignored by encoding/json; presence SELECT on the
	// legacy column remains until inventory-zero DROP (HOLD).
	LiffID          string `json:"liff_id" binding:"max=255"`
	LineAccessToken string `json:"line_access_token" binding:"max=2000"`
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
		LineChannelID:           r.LineChannelID,
		LiffID:                  r.LiffID,
		LineAccessToken:         r.LineAccessToken,
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
