package model

import (
	"time"
)

// LstepTriggerPriority はクリニック単位のトリガー配信優先順位設定 (Q23)。
type LstepTriggerPriority struct {
	ID          uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" `
	ClinicID    uint64    `gorm:"column:clinic_id;type:bigint;not null" `
	TriggerType string    `gorm:"column:trigger_type;type:varchar(64);not null" `
	Priority    int       `gorm:"column:priority;type:integer;not null" `
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime" `
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:NOW();autoUpdateTime" `
}

func (LstepTriggerPriority) TableName() string { return "lstep_trigger_priorities" }

// DefaultTriggerPriorities はDB未設定時のシステムデフォルト優先順位マップ (Q23 PO-QA 2026-05-08)。
// 小さいほど優先 (1=最優先)。
var DefaultTriggerPriorities = map[string]int{
	TriggerTypeDormantPrevention365: 2,
	TriggerTypeCheckupFollowUp:      3,
	TriggerTypeFilariaAlert:         4,
	TriggerTypeFleaTickAlert:        4,
	TriggerTypeDormantPrevention240: 5,
	TriggerTypeDormantPrevention210: 6,
	TriggerTypeDormantPrevention180: 7,
	TriggerTypeVaccineDeadline30:    8,
	TriggerTypeVaccineDeadline60:    8,
	TriggerTypeFoodRefillReminder:   10,
	// 未マッピング (priority 11+) — EMR-308: tentative リマインダー 2 ステージは
	// legacy next_visit_reminder と同順位 11（旧型は歴史ログ/優先度互換のため登録維持）。
	TriggerTypeNextVisitReminder:            11,
	TriggerTypeNextVisitTentativeReminder3d: 11,
	TriggerTypeNextVisitTentativeReminder0d: 11,
	TriggerTypeBirthdayMessage:              12,
	TriggerTypeFirstVisitFollowUp3D:         13,
	TriggerTypeFirstVisitFollowUp7D:         13,
	TriggerTypeFirstVisitWelcome:            14,
}

// DefaultPriorityFallback は DefaultTriggerPriorities に未定義のトリガーに使う最後尾優先度。
const DefaultPriorityFallback = 99
