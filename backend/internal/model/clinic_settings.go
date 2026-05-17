package model

import (
	"time"

	"github.com/lib/pq"
)

// ClinicSettings は診療所の締め時間設定
type ClinicSettings struct {
	ClinicID            uint64        `gorm:"primaryKey"                            json:"clinic_id"`
	ClosingAmPmBoundary string        `gorm:"type:time;not null;default:'14:00'"    json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   string        `gorm:"type:time;not null;default:'18:30'"    json:"closing_weekday_end"`
	ClosingSundayEnd    string        `gorm:"type:time;not null;default:'17:30'"    json:"closing_sunday_end"`
	ClosedWeekdays      pq.Int64Array `gorm:"type:smallint[];not null;default:'{}'" json:"closed_weekdays"`
	LstepFireHourJST    int           `gorm:"not null;default:10"                  json:"lstep_fire_hour_jst"`
	CPMVersion          string        `gorm:"type:varchar(8);not null;default:'v1'" json:"cpm_version"`
	// Q21 SPEC-004 dormant prevention 閾値 (clinic 単位調整可能)
	DormantPrevention180Days int `gorm:"not null;default:180" json:"dormant_prevention_180_days"`
	DormantPrevention210Days int `gorm:"not null;default:210" json:"dormant_prevention_210_days"`
	DormantPrevention240Days int `gorm:"not null;default:240" json:"dormant_prevention_240_days"`
	DormantPrevention365Days int `gorm:"not null;default:365" json:"dormant_prevention_365_days"`
	// P1 CPM V2 来院回数閾値 (clinic 単位調整可能、migration 007)
	CPMV2ComingThreshold int       `gorm:"not null;default:2"   json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold   int       `gorm:"not null;default:4"   json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold int       `gorm:"not null;default:8"   json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold   int       `gorm:"not null;default:13"  json:"cpm_v2_noah_threshold"`
	CreatedAt            time.Time `gorm:"autoCreateTime"       json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"       json:"updated_at"`
}

func (ClinicSettings) TableName() string { return "clinic_settings" }
