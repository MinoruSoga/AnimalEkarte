package model

import (
	"encoding/json"
	"time"
)

type LineReservationSetting struct {
	ID                      uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID                uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Status                  string           `gorm:"column:status;type:text;not null;default:'stopped'" json:"status"`
	HeaderText              string           `gorm:"column:header_text;type:text;not null;default:''" json:"header_text"`
	ReservationNotice       string           `gorm:"column:reservation_notice;type:text;not null;default:''" json:"reservation_notice"`
	CancelNotice            string           `gorm:"column:cancel_notice;type:text;not null;default:''" json:"cancel_notice"`
	PrivacyPolicy           string           `gorm:"column:privacy_policy;type:text;not null;default:''" json:"privacy_policy"`
	ClosedWeekdays          json.RawMessage  `gorm:"column:closed_weekdays;type:jsonb;not null;default:'[]'" json:"closed_weekdays"`
	ClosedDates             json.RawMessage  `gorm:"column:closed_dates;type:jsonb;not null;default:'[]'" json:"closed_dates"`
	NationalHolidayClosed   bool             `gorm:"column:national_holiday_closed;type:boolean;not null;default:false" json:"national_holiday_closed"`
	BusinessHours           json.RawMessage  `gorm:"column:business_hours;type:jsonb;not null;default:'{\"start\":\"0900\",\"end\":\"1900\"}'" json:"business_hours"`
	BusinessHoursByWeekday  *json.RawMessage `gorm:"column:business_hours_by_weekday;type:jsonb" json:"business_hours_by_weekday,omitempty"`
	BreakHours              json.RawMessage  `gorm:"column:break_hours;type:jsonb;not null;default:'[{\"start\":\"1200\",\"end\":\"1300\"}]'" json:"break_hours"`
	DailyLimit              *int             `gorm:"column:daily_limit;type:int;default:1" json:"daily_limit,omitempty"`
	MonthlyLimit            *int             `gorm:"column:monthly_limit;type:int" json:"monthly_limit,omitempty"`
	BookingWindowMaxDays    int              `gorm:"column:booking_window_max_days;type:int;not null;default:30" json:"booking_window_max_days"`
	BookingWindowMinDays    int              `gorm:"column:booking_window_min_days;type:int;not null;default:2" json:"booking_window_min_days"`
	CalendarMonths          int              `gorm:"column:calendar_months;type:int;not null;default:2" json:"calendar_months"`
	PhoneNumber             string           `gorm:"column:phone_number;type:text;not null;default:''" json:"phone_number"`
	NotificationEmail       string           `gorm:"column:notification_email;type:text;not null;default:''" json:"notification_email"`
	RequestExample          string           `gorm:"column:request_example;type:text;not null;default:''" json:"request_example"`
	TimeSlotMode            string           `gorm:"column:time_slot_mode;type:text;not null;default:'minimize_gaps'" json:"time_slot_mode"`
	TimeSlotIntervalMinutes int              `gorm:"column:time_slot_interval_minutes;type:int;not null;default:15" json:"time_slot_interval_minutes"`
	NoStaffMode             string           `gorm:"column:no_staff_mode;type:text;not null;default:'first_available'" json:"no_staff_mode"`
	ShowNoStaffOption       bool             `gorm:"column:show_no_staff_option;type:boolean;not null;default:true" json:"show_no_staff_option"`
	AdditionalFields        json.RawMessage  `gorm:"column:additional_fields;type:jsonb;not null;default:'[\n        {\"key\":\"phone\",\"label\":\"電話番号\",\"required\":true,\"placeholder\":\"例) 090-1234-5678\"},\n        {\"key\":\"owner_name\",\"label\":\"飼い主名\",\"required\":true,\"placeholder\":\"\"},\n        {\"key\":\"pet_info\",\"label\":\"ペットの名前と種類\",\"required\":true,\"placeholder\":\"例) ポチ（柴犬）\"},\n        {\"key\":\"symptoms\",\"label\":\"診察内容\",\"required\":false,\"placeholder\":\"\"}\n    ]'" json:"additional_fields"`
	CreatedAt               time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
}

func (LineReservationSetting) TableName() string { return "line_reservation_settings" }
