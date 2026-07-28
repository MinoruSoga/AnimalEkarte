package reservation

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type lineReservationSettingResponse struct {
	ID                      uint64          `json:"id"`
	ClinicID                uint64          `json:"clinic_id"`
	Status                  string          `json:"status"`
	HeaderText              string          `json:"header_text"`
	ReservationNotice       string          `json:"reservation_notice"`
	CancelNotice            string          `json:"cancel_notice"`
	PrivacyPolicy           string          `json:"privacy_policy"`
	ClosedWeekdays          json.RawMessage `json:"closed_weekdays"`
	ClosedDates             json.RawMessage `json:"closed_dates"`
	NationalHolidayClosed   bool            `json:"national_holiday_closed"`
	BusinessHours           json.RawMessage `json:"business_hours"`
	BusinessHoursByWeekday  json.RawMessage `json:"business_hours_by_weekday,omitempty"`
	BreakHours              json.RawMessage `json:"break_hours"`
	DailyLimit              *int            `json:"daily_limit,omitempty"`
	MonthlyLimit            *int            `json:"monthly_limit,omitempty"`
	BookingWindowMaxDays    int             `json:"booking_window_max_days"`
	BookingWindowMinDays    int             `json:"booking_window_min_days"`
	CalendarMonths          int             `json:"calendar_months"`
	PhoneNumber             string          `json:"phone_number"`
	NotificationEmail       string          `json:"notification_email"`
	RequestExample          string          `json:"request_example"`
	TimeSlotMode            string          `json:"time_slot_mode"`
	TimeSlotIntervalMinutes int             `json:"time_slot_interval_minutes"`
	NoStaffMode             string          `json:"no_staff_mode"`
	ShowNoStaffOption       bool            `json:"show_no_staff_option"`
	AdditionalFields        json.RawMessage `json:"additional_fields"`
	LineChannelID           string          `json:"line_channel_id"`
	// LineChannelSecret と LineAccessToken はセキュリティ上レスポンスに含めない
	LiffID    string    `json:"liff_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toLineReservationSettingResponse(s *model.LineReservationSetting) lineReservationSettingResponse {
	return lineReservationSettingResponse{
		ID:                      s.ID,
		ClinicID:                s.ClinicID,
		Status:                  s.Status,
		HeaderText:              s.HeaderText,
		ReservationNotice:       s.ReservationNotice,
		CancelNotice:            s.CancelNotice,
		PrivacyPolicy:           s.PrivacyPolicy,
		ClosedWeekdays:          json.RawMessage(s.ClosedWeekdays),
		ClosedDates:             json.RawMessage(s.ClosedDates),
		NationalHolidayClosed:   s.NationalHolidayClosed,
		BusinessHours:           json.RawMessage(s.BusinessHours),
		BusinessHoursByWeekday:  json.RawMessage(s.BusinessHoursByWeekday),
		BreakHours:              json.RawMessage(s.BreakHours),
		DailyLimit:              s.DailyLimit,
		MonthlyLimit:            s.MonthlyLimit,
		BookingWindowMaxDays:    s.BookingWindowMaxDays,
		BookingWindowMinDays:    s.BookingWindowMinDays,
		CalendarMonths:          s.CalendarMonths,
		PhoneNumber:             s.PhoneNumber,
		NotificationEmail:       s.NotificationEmail,
		RequestExample:          s.RequestExample,
		TimeSlotMode:            s.TimeSlotMode,
		TimeSlotIntervalMinutes: s.TimeSlotIntervalMinutes,
		NoStaffMode:             s.NoStaffMode,
		ShowNoStaffOption:       s.ShowNoStaffOption,
		AdditionalFields:        json.RawMessage(s.AdditionalFields),
		LineChannelID:           s.LineChannelID,
		LiffID:                  s.LiffID,
		CreatedAt:               httpapi.LocalTime(s.CreatedAt),
		UpdatedAt:               httpapi.LocalTime(s.UpdatedAt),
	}
}
