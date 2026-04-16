package model

import (
	"time"
)

// ReservationDayOption defines which weekdays a service type is available for LINE reservation.
type ReservationDayOption string

const (
	DayOptionNone     ReservationDayOption = "none"
	DayOptionSaturday ReservationDayOption = "saturday"
	DayOptionWeekday  ReservationDayOption = "weekday"
	DayOptionAnyday   ReservationDayOption = "anyday"
)

// ReservationTypeCategory は予約区分のカテゴリ（'general' | 'trimming'）
type ReservationTypeCategory string

const (
	ReservationTypeCategoryGeneral  ReservationTypeCategory = "general"
	ReservationTypeCategoryTrimming ReservationTypeCategory = "trimming"
)

// UnavailableType は予約不可時間の種別
type UnavailableType string

const (
	UnavailableTypeWeekly   UnavailableType = "weekly"
	UnavailableTypeSpecific UnavailableType = "specific"
)

// ReservationType はサービス種別（予約区分）マスタ
type ReservationType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                              json:"description"`
	Color       string    `gorm:"default:'#3B82F6'"                              json:"color"`
	SortOrder   int       `gorm:"type:integer;default:0"                         json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// LINE予約用フィールド
	ReservationDisplayName string               `gorm:"not null;default:''"                    json:"reservation_display_name"`
	DurationMinutes        int                  `gorm:"not null;default:15"                        json:"duration_minutes"`
	ShortName              string               `gorm:"not null;default:''"                    json:"short_name"`
	ShowShortName          bool                 `gorm:"not null;default:false"                     json:"show_short_name"`
	ReservationVisible     bool                 `gorm:"not null;default:true"                      json:"reservation_visible"`
	ReservationComment     string               `gorm:"not null;default:''"                    json:"reservation_comment"`
	ReservationImageURL    string               `gorm:"not null;default:''"                    json:"reservation_image_url"`
	ReservationDayOption   ReservationDayOption `gorm:"not null;default:'none'"                    json:"reservation_day_option"`
	IsInternal             bool                 `gorm:"not null;default:false"                     json:"is_internal"`

	// カテゴリ（'general' | 'trimming'）
	Category ReservationTypeCategory `gorm:"type:reservation_type_category;not null;default:'general'" json:"category"`

	// グループ（カレンダー凡例用）
	GroupID *uint64               `gorm:"index"              json:"group_id"`
	Group   *ReservationTypeGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`

	// Relations（BE-115）
	UnavailableTimes []ReservationTypeUnavailableTime `gorm:"foreignKey:ReservationTypeID" json:"unavailable_times,omitempty"`
	Occupations      []ReservationTypeOccupation      `gorm:"foreignKey:ReservationTypeID" json:"occupations,omitempty"`
}

func (ReservationType) TableName() string { return "reservation_types" }

// ReservationTypeUnavailableTime は予約区分の予約不可時間帯（BE-115）
type ReservationTypeUnavailableTime struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement"         json:"id"`
	ClinicID          uint64          `gorm:"not null"                         json:"clinic_id"`
	ReservationTypeID uint64          `gorm:"not null"                         json:"reservation_type_id"`
	UnavailableType   UnavailableType `gorm:"not null"                         json:"unavailable_type"`
	DayOfWeek         *int8           `                                        json:"day_of_week,omitempty"`   // 0=Sun..6=Sat（weekly のみ）
	SpecificDate      *time.Time      `gorm:"type:date"                        json:"specific_date,omitempty"` // specific のみ
	StartTime         string          `gorm:"type:varchar(5);not null"         json:"start_time"`              // "HH:MM"
	EndTime           string          `gorm:"type:varchar(5);not null"         json:"end_time"`                // "HH:MM"
	CreatedAt         time.Time       `gorm:"autoCreateTime"                   json:"created_at"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime"                   json:"updated_at"`
}

func (ReservationTypeUnavailableTime) TableName() string { return "reservation_type_unavailable_times" }

// ReservationTypeOccupation は予約区分と職種の紐付け（M:N）（BE-115）
type ReservationTypeOccupation struct {
	ID                uint64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ClinicID          uint64      `gorm:"not null"                 json:"clinic_id"`
	ReservationTypeID uint64      `gorm:"not null"                 json:"reservation_type_id"`
	OccupationID      uint64      `gorm:"not null"                 json:"occupation_id"`
	Occupation        *Occupation `gorm:"foreignKey:OccupationID"  json:"occupation,omitempty"`
	CreatedAt         time.Time   `gorm:"autoCreateTime"           json:"created_at"`
}

func (ReservationTypeOccupation) TableName() string { return "reservation_type_occupations" }
