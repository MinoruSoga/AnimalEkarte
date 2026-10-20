package model

import (
	"github.com/lib/pq"
	"time"
)

// ClinicSettings は診療所の締め時間設定
type ClinicSettings struct {
	ClinicID                     uint64        `gorm:"column:clinic_id;type:bigint;primaryKey;autoIncrement:false;not null" json:"clinic_id"`
	ClosingAmPmBoundary          string        `gorm:"column:closing_am_pm_boundary;type:time;not null;default:'14:00'" json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd            string        `gorm:"column:closing_weekday_end;type:time;not null;default:'18:30'" json:"closing_weekday_end"`
	ClosingSundayEnd             string        `gorm:"column:closing_sunday_end;type:time;not null;default:'17:30'" json:"closing_sunday_end"`
	ClosedWeekdays               pq.Int64Array `gorm:"column:closed_weekdays;type:smallint[];not null;default:'{}'" json:"closed_weekdays"`
	CPMVersion                   string        `gorm:"column:cpm_version;type:varchar(8);not null;default:'v1'" json:"cpm_version"`
	DormantPrevention180Days     int           `gorm:"column:dormant_prevention_180_days;type:integer;not null;default:180" json:"dormant_prevention_180_days"`
	DormantPrevention210Days     int           `gorm:"column:dormant_prevention_210_days;type:integer;not null;default:210" json:"dormant_prevention_210_days"`
	DormantPrevention240Days     int           `gorm:"column:dormant_prevention_240_days;type:integer;not null;default:240" json:"dormant_prevention_240_days"`
	DormantPrevention365Days     int           `gorm:"column:dormant_prevention_365_days;type:integer;not null;default:365" json:"dormant_prevention_365_days"`
	CPMV2ComingThreshold         int           `gorm:"column:cpm_v2_coming_threshold;type:int;not null;default:2" json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold           int           `gorm:"column:cpm_v2_good_threshold;type:int;not null;default:4" json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold         int           `gorm:"column:cpm_v2_family_threshold;type:int;not null;default:8" json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold           int           `gorm:"column:cpm_v2_noah_threshold;type:int;not null;default:13" json:"cpm_v2_noah_threshold"`
	CPMV1DormantDays             int           `gorm:"column:cpm_v1_dormant_days;type:int;not null;default:240" json:"cpm_v1_dormant_days"`
	CPMV1NoahDays                int           `gorm:"column:cpm_v1_noah_days;type:int;not null;default:365" json:"cpm_v1_noah_days"`
	CPMV1NoahAnnualVisits        int           `gorm:"column:cpm_v1_noah_annual_visits;type:int;not null;default:3" json:"cpm_v1_noah_annual_visits"`
	CPMV1NoahLTV                 int64         `gorm:"column:cpm_v1_noah_ltv;type:bigint;not null;default:80000" json:"cpm_v1_noah_ltv"`
	CPMV1CoreDays                int           `gorm:"column:cpm_v1_core_days;type:int;not null;default:180" json:"cpm_v1_core_days"`
	CPMV1CoreAnnualVisits        int           `gorm:"column:cpm_v1_core_annual_visits;type:int;not null;default:2" json:"cpm_v1_core_annual_visits"`
	CPMV1CoreLTV                 int64         `gorm:"column:cpm_v1_core_ltv;type:bigint;not null;default:50000" json:"cpm_v1_core_ltv"`
	CPMV1SpotMinAmount           int64         `gorm:"column:cpm_v1_spot_min_amount;type:bigint;not null;default:30000" json:"cpm_v1_spot_min_amount"`
	CPMV1SpotInactiveDays        int           `gorm:"column:cpm_v1_spot_inactive_days;type:int;not null;default:90" json:"cpm_v1_spot_inactive_days"`
	CPMV1GrowingMaxDays          int           `gorm:"column:cpm_v1_growing_max_days;type:int;not null;default:90" json:"cpm_v1_growing_max_days"`
	CPMV1GrowingMinVisits        int           `gorm:"column:cpm_v1_growing_min_visits;type:int;not null;default:2" json:"cpm_v1_growing_min_visits"`
	CPMV1GrowingMaxVisits        int           `gorm:"column:cpm_v1_growing_max_visits;type:int;not null;default:3" json:"cpm_v1_growing_max_visits"`
	CPMV1LTVBreakLow             int64         `gorm:"column:cpm_v1_ltv_break_low;type:bigint;not null;default:20000" json:"cpm_v1_ltv_break_low"`
	HealthPreventionLookbackDays int           `gorm:"column:health_prevention_lookback_days;type:int;not null;default:365" json:"health_prevention_lookback_days"`
	VaccineDeadlineDays          int           `gorm:"column:vaccine_deadline_days;type:int;not null;default:60" json:"vaccine_deadline_days"`
	CreatedAt                    time.Time     `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt                    time.Time     `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	ClosingAmStart               string        `gorm:"column:closing_am_start;type:time;not null;default:'09:00'" json:"closing_am_start"`
}

func (ClinicSettings) TableName() string { return "clinic_settings" }
