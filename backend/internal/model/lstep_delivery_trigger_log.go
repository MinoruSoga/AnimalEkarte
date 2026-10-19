package model

import (
	"time"
)

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
	TriggerTypeFirstVisitWelcome    TriggerType = "first_visit_welcome"
	TriggerTypeCheckupFollowUp      TriggerType = "checkup_followup"
	// EMR-308: 次回予定（tentative appointment）リマインダーの 2 ステージ。
	// trigger_type 自体がステージ識別子となり、(owner, date, trigger_type) の
	// 日次 claim がステージ単位の抑制・冪等性を保証する（stage 列追加なし）。
	TriggerTypeNextVisitTentativeReminder3d TriggerType = "next_visit_tentative_reminder_3d"
	TriggerTypeNextVisitTentativeReminder0d TriggerType = "next_visit_tentative_reminder_0d"
)

// FE6-1: tygo の detectEnumGroup は 1 つの const() ブロック（1 GenDecl）につき
// 先頭定数の型のみを基準にグループ化するため、TriggerType/TriggerStatus が
// 同一ブロックに混在していると後発の TriggerStatus が enum_style: "union" の対象から
// 脱落する。ブロックを分離するだけで両方が正しく union 化される（値・型・エクスポートは不変）。
const (
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
		TriggerTypeFirstVisitWelcome,
		TriggerTypeCheckupFollowUp,
		TriggerTypeNextVisitTentativeReminder3d,
		TriggerTypeNextVisitTentativeReminder0d,
	}
}

// LstepDeliveryTriggerLog は自動配信トリガーの実行ログ。
type LstepDeliveryTriggerLog struct {
	ID                   uint64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" `
	OwnerID              uint64     `gorm:"column:owner_id;type:bigint;not null" `
	ClinicID             uint64     `gorm:"column:clinic_id;type:bigint;not null" `
	TriggerType          string     `gorm:"column:trigger_type;type:varchar(50);not null" `
	ScheduledAt          time.Time  `gorm:"column:scheduled_at;type:timestamptz;not null" `
	Status               string     `gorm:"column:status;type:varchar(20);not null;default:'scheduled'" `
	FiredAt              *time.Time `gorm:"column:fired_at;type:timestamptz" json:"-"`
	ExcludedReason       *string    `gorm:"column:excluded_reason;type:varchar(100)" `
	SuppressedByPriority bool       `gorm:"column:suppressed_by_priority;type:boolean;not null;default:FALSE" json:"suppressed_by_priority"`
	SuppressionReason    *string    `gorm:"column:suppression_reason;type:varchar(255)" json:"suppression_reason,omitempty"`
	CreatedAt            time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" `
	UpdatedAt            time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" `
}

func (LstepDeliveryTriggerLog) TableName() string { return "lstep_delivery_trigger_log" }
