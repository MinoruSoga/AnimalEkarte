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

	// グループ（カレンダー凡例用）
	GroupID *uint64               `gorm:"index"              json:"group_id"`
	Group   *ReservationTypeGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

func (ReservationType) TableName() string { return "reservation_types" }
