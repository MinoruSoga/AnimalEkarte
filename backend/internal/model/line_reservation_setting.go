package model

import "time"

type LineReservationSetting struct {
	ID                      uint64 `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID                uint64 `gorm:"not null;uniqueIndex"                           json:"clinic_id"`
	Status                  string `gorm:"not null;default:'stopped'"                     json:"status"`
	HeaderText              string `gorm:"not null;default:''"                            json:"header_text"`
	ReservationNotice       string `gorm:"not null;default:''"                            json:"reservation_notice"`
	CancelNotice            string `gorm:"not null;default:''"                            json:"cancel_notice"`
	PrivacyPolicy           string `gorm:"not null;default:''"                            json:"privacy_policy"`
	ClosedWeekdays          []byte `gorm:"type:jsonb;not null;default:'[]'"               json:"closed_weekdays"`
	ClosedDates             []byte `gorm:"type:jsonb;not null;default:'[]'"               json:"closed_dates"`
	NationalHolidayClosed   bool   `gorm:"not null;default:false"                         json:"national_holiday_closed"`
	BusinessHours           []byte `gorm:"type:jsonb;not null;default:'{\"start\":\"0900\",\"end\":\"1900\"}'" json:"business_hours"`
	BusinessHoursByWeekday  []byte `gorm:"type:jsonb"                                     json:"business_hours_by_weekday,omitempty"`
	BreakHours              []byte `gorm:"type:jsonb;not null;default:'[{\"start\":\"1200\",\"end\":\"1300\"}]'" json:"break_hours"`
	DailyLimit              *int   `                                                      json:"daily_limit,omitempty"`
	MonthlyLimit            *int   `                                                      json:"monthly_limit,omitempty"`
	BookingWindowMaxDays    int    `gorm:"not null;default:30"                            json:"booking_window_max_days"`
	BookingWindowMinDays    int    `gorm:"not null;default:2"                             json:"booking_window_min_days"`
	CalendarMonths          int    `gorm:"not null;default:2"                             json:"calendar_months"`
	PhoneNumber             string `gorm:"not null;default:''"                            json:"phone_number"`
	NotificationEmail       string `gorm:"not null;default:''"                            json:"notification_email"`
	RequestExample          string `gorm:"not null;default:''"                            json:"request_example"`
	TimeSlotMode            string `gorm:"not null;default:'minimize_gaps'"               json:"time_slot_mode"`
	TimeSlotIntervalMinutes int    `gorm:"not null;default:15"                            json:"time_slot_interval_minutes"`
	NoStaffMode             string `gorm:"not null;default:'first_available'"             json:"no_staff_mode"`
	ShowNoStaffOption       bool   `gorm:"not null;default:true"                          json:"show_no_staff_option"`
	AdditionalFields        []byte `gorm:"type:jsonb;not null"                            json:"additional_fields"`
	LineChannelID           string `gorm:"not null;default:''"                            json:"line_channel_id"`
	LineChannelSecret       string `gorm:"not null;default:''"                            json:"-"`
	// LineBotUserID is the LINE Messaging API bot user ID (webhook destination).
	// Used for O(1) webhook signature routing (SEC-CS-F05-R1). Empty until provisioned.
	LineBotUserID   string    `gorm:"not null;default:''"                            json:"line_bot_user_id,omitempty"`
	LiffID          string    `gorm:"not null;default:''"                            json:"liff_id"`
	LineAccessToken string    `gorm:"not null;default:''"                            json:"-"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (LineReservationSetting) TableName() string { return "line_reservation_settings" }
