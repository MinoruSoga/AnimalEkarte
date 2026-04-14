package handler

import "encoding/json"

type upsertLineReservationSettingRequest struct {
	Status                  string         `json:"status"`
	HeaderText              string         `json:"header_text"`
	ReservationNotice       string         `json:"reservation_notice"`
	CancelNotice            string         `json:"cancel_notice"`
	PrivacyPolicy           string         `json:"privacy_policy"`
	ClosedWeekdays          jsonRawOrEmpty `json:"closed_weekdays"`
	ClosedDates             jsonRawOrEmpty `json:"closed_dates"`
	NationalHolidayClosed   bool           `json:"national_holiday_closed"`
	BusinessHours           jsonRawOrEmpty `json:"business_hours"`
	BusinessHoursByWeekday  jsonRawOrEmpty `json:"business_hours_by_weekday"`
	BreakHours              jsonRawOrEmpty `json:"break_hours"`
	DailyLimit              *int           `json:"daily_limit"`
	MonthlyLimit            *int           `json:"monthly_limit"`
	BookingWindowMaxDays    int            `json:"booking_window_max_days"`
	BookingWindowMinDays    int            `json:"booking_window_min_days"`
	CalendarMonths          int            `json:"calendar_months"`
	PhoneNumber             string         `json:"phone_number"`
	NotificationEmail       string         `json:"notification_email"`
	RequestExample          string         `json:"request_example"`
	TimeSlotMode            string         `json:"time_slot_mode"`
	TimeSlotIntervalMinutes int            `json:"time_slot_interval_minutes"`
	NoStaffMode             string         `json:"no_staff_mode"`
	ShowNoStaffOption       bool           `json:"show_no_staff_option"`
	AdditionalFields        jsonRawOrEmpty `json:"additional_fields"`
	LineChannelID           string         `json:"line_channel_id"`
	LineChannelSecret       string         `json:"line_channel_secret"`
	LiffID                  string         `json:"liff_id"`
	LineAccessToken         string         `json:"line_access_token"`
}

// jsonRawOrEmpty は JSON フィールドを生の JSON として保持するためのエイリアス。
// 標準の []byte ではなく json.RawMessage を使うのは、encoding/json が []byte フィールドを
// base64 エンコード文字列として扱うため（BUG-LINE-007）。json.RawMessage は MarshalJSON /
// UnmarshalJSON が実装されており任意の JSON オブジェクト・配列をそのまま保持できる。
// 実体は []byte なので service 層の []byte 引数にはそのまま渡せる。
type jsonRawOrEmpty = json.RawMessage
