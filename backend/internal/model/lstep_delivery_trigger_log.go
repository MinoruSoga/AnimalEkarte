package model

import "time"

// TriggerType は配信トリガーの種別を表す型エイリアス。
type TriggerType = string

// TriggerStatus は配信トリガーログのステータスを表す型エイリアス。
type TriggerStatus = string

const (
	TriggerTypeFirstVisitFollowUp3D TriggerType = "first_visit_followup_3d"
	TriggerTypeFirstVisitFollowUp7D TriggerType = "first_visit_followup_7d"
	TriggerTypeNextVisitReminder    TriggerType = "next_visit_reminder"
	TriggerTypeVaccineDeadline60    TriggerType = "vaccine_deadline_60d"
	TriggerTypeVaccineDeadline30    TriggerType = "vaccine_deadline_30d"
	TriggerTypeBirthdayMessage      TriggerType = "birthday_message"
	TriggerTypeDormantPrevention180 TriggerType = "dormant_prevention_180d"
	TriggerTypeDormantPrevention210 TriggerType = "dormant_prevention_210d"
	TriggerTypeDormantPrevention240 TriggerType = "dormant_prevention_240d"
	TriggerTypeDormantPrevention365 TriggerType = "dormant_prevention_365d"
	TriggerTypeFilariaAlert         TriggerType = "filaria_alert"
	TriggerTypeFleaTickAlert        TriggerType = "flea_tick_alert"
	TriggerTypeFoodRefillReminder   TriggerType = "food_refill_reminder"
	TriggerTypeSuppRefillReminder   TriggerType = "supp_refill_reminder"
	TriggerTypeFirstVisitWelcome    TriggerType = "first_visit_welcome"
	TriggerTypeCheckupFollowUp      TriggerType = "checkup_followup"

	TriggerStatusScheduled TriggerStatus = "scheduled"
	TriggerStatusFired     TriggerStatus = "fired"
	TriggerStatusExcluded  TriggerStatus = "excluded"
	TriggerStatusFailed    TriggerStatus = "failed"
)

// AllTriggerTypes は定義済みの全トリガー種別スライスを返す。
func AllTriggerTypes() []string {
	return []string{
		TriggerTypeFirstVisitFollowUp3D,
		TriggerTypeFirstVisitFollowUp7D,
		TriggerTypeNextVisitReminder,
		TriggerTypeVaccineDeadline60,
		TriggerTypeVaccineDeadline30,
		TriggerTypeBirthdayMessage,
		TriggerTypeDormantPrevention180,
		TriggerTypeDormantPrevention210,
		TriggerTypeDormantPrevention240,
		TriggerTypeDormantPrevention365,
		TriggerTypeFilariaAlert,
		TriggerTypeFleaTickAlert,
		TriggerTypeFoodRefillReminder,
		TriggerTypeSuppRefillReminder,
		TriggerTypeFirstVisitWelcome,
		TriggerTypeCheckupFollowUp,
	}
}

// LstepDeliveryTriggerLog は自動配信トリガーの実行ログ。
type LstepDeliveryTriggerLog struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement"`
	OwnerID              uint64    `gorm:"not null"`
	ClinicID             uint64    `gorm:"not null"`
	TriggerType          string    `gorm:"type:varchar(50);not null"`
	ScheduledAt          time.Time `gorm:"not null"`
	Status               string    `gorm:"type:varchar(20);not null;default:'scheduled'"`
	FiredAt              *time.Time
	ExcludedReason       *string   `gorm:"type:varchar(100)"`
	SuppressedByPriority bool      `gorm:"not null;default:false" json:"suppressed_by_priority"`
	SuppressionReason    *string   `gorm:"type:varchar(255)" json:"suppression_reason,omitempty"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`
}
