package model

import (
	"time"

	"github.com/lib/pq"
)

// ClinicSettings は診療所の締め時間設定
type ClinicSettings struct {
	ClinicID            uint64 `gorm:"primaryKey"                            json:"clinic_id"`
	ClosingAmPmBoundary string `gorm:"type:time;not null;default:'14:00'"    json:"closing_am_pm_boundary"`
	ClosingWeekdayEnd   string `gorm:"type:time;not null;default:'18:30'"    json:"closing_weekday_end"`
	ClosingSundayEnd    string `gorm:"type:time;not null;default:'17:30'"    json:"closing_sunday_end"`
	// #215: AM 開始時刻。AM=[am_start, boundary) / EMG=[pm_end, 翌日 am_start) の越日レンジに使う（migration 011）。
	ClosingAmStart string        `gorm:"type:time;not null;default:'09:00'"    json:"closing_am_start"`
	ClosedWeekdays pq.Int64Array `gorm:"type:smallint[];not null;default:'{}'" json:"closed_weekdays"`
	CPMVersion     string        `gorm:"type:varchar(8);not null;default:'v1'" json:"cpm_version"`
	// Q21 SPEC-004 dormant prevention 閾値 (clinic 単位調整可能)
	// column: 必須 — GORM デフォルトは Prevention180 → prevention180（数字前に _ なし）となり
	// 実スキーマ dormant_prevention_180_days と不一致で SQLSTATE 42703 になる（#236 BUG#3）。
	DormantPrevention180Days int `gorm:"column:dormant_prevention_180_days;not null;default:180" json:"dormant_prevention_180_days"`
	DormantPrevention210Days int `gorm:"column:dormant_prevention_210_days;not null;default:210" json:"dormant_prevention_210_days"`
	DormantPrevention240Days int `gorm:"column:dormant_prevention_240_days;not null;default:240" json:"dormant_prevention_240_days"`
	DormantPrevention365Days int `gorm:"column:dormant_prevention_365_days;not null;default:365" json:"dormant_prevention_365_days"`
	// P1 CPM V2 来院回数閾値 (clinic 単位調整可能、migration 007)
	// column: 必須 — CPMV2 → cpmv2（_v2 にならない）となり実スキーマ cpm_v2_* と不一致（#236）。
	CPMV2ComingThreshold int `gorm:"column:cpm_v2_coming_threshold;not null;default:2"  json:"cpm_v2_coming_threshold"`
	CPMV2GoodThreshold   int `gorm:"column:cpm_v2_good_threshold;not null;default:4"  json:"cpm_v2_good_threshold"`
	CPMV2FamilyThreshold int `gorm:"column:cpm_v2_family_threshold;not null;default:8"  json:"cpm_v2_family_threshold"`
	CPMV2NoahThreshold   int `gorm:"column:cpm_v2_noah_threshold;not null;default:13" json:"cpm_v2_noah_threshold"`
	// P2 CPM V1 判定閾値 (clinic 単位調整可能、migration 008)
	CPMV1DormantDays      int   `gorm:"column:cpm_v1_dormant_days;not null;default:240"   json:"cpm_v1_dormant_days"`
	CPMV1NoahDays         int   `gorm:"column:cpm_v1_noah_days;not null;default:365"   json:"cpm_v1_noah_days"`
	CPMV1NoahAnnualVisits int   `gorm:"column:cpm_v1_noah_annual_visits;not null;default:3"     json:"cpm_v1_noah_annual_visits"`
	CPMV1NoahLTV          int64 `gorm:"column:cpm_v1_noah_ltv;not null;default:80000" json:"cpm_v1_noah_ltv"`
	CPMV1CoreDays         int   `gorm:"column:cpm_v1_core_days;not null;default:180"   json:"cpm_v1_core_days"`
	CPMV1CoreAnnualVisits int   `gorm:"column:cpm_v1_core_annual_visits;not null;default:2"     json:"cpm_v1_core_annual_visits"`
	CPMV1CoreLTV          int64 `gorm:"column:cpm_v1_core_ltv;not null;default:50000" json:"cpm_v1_core_ltv"`
	CPMV1SpotMinAmount    int64 `gorm:"column:cpm_v1_spot_min_amount;not null;default:30000" json:"cpm_v1_spot_min_amount"`
	CPMV1SpotInactiveDays int   `gorm:"column:cpm_v1_spot_inactive_days;not null;default:90"    json:"cpm_v1_spot_inactive_days"`
	CPMV1GrowingMaxDays   int   `gorm:"column:cpm_v1_growing_max_days;not null;default:90"    json:"cpm_v1_growing_max_days"`
	CPMV1GrowingMinVisits int   `gorm:"column:cpm_v1_growing_min_visits;not null;default:2"     json:"cpm_v1_growing_min_visits"`
	CPMV1GrowingMaxVisits int   `gorm:"column:cpm_v1_growing_max_visits;not null;default:3"     json:"cpm_v1_growing_max_visits"`
	CPMV1LTVBreakLow      int64 `gorm:"column:cpm_v1_ltv_break_low;not null;default:20000" json:"cpm_v1_ltv_break_low"`
	// P9 健診・予防タグ判定閾値 (clinic 単位調整可能、migration 009)
	HealthPreventionLookbackDays int       `gorm:"column:health_prevention_lookback_days;not null;default:365" json:"health_prevention_lookback_days"`
	VaccineDeadlineDays          int       `gorm:"column:vaccine_deadline_days;not null;default:60"  json:"vaccine_deadline_days"`
	CreatedAt                    time.Time `gorm:"autoCreateTime"     json:"created_at"`
	UpdatedAt                    time.Time `gorm:"autoUpdateTime"     json:"updated_at"`
}

func (ClinicSettings) TableName() string { return "clinic_settings" }
