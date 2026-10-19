package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

// SupportChatMessageRole は保存済みチャット履歴の発言者（LLM へ送る system は保存しない）
type SupportChatMessageRole string

const (
	SupportChatRoleUser      SupportChatMessageRole = "user"
	SupportChatRoleAssistant SupportChatMessageRole = "assistant"
)

// SupportChatMessage はヘルプチャットの会話履歴1件。
// clinic_id × staff_id で「選択clinic内のスタッフ個人」にスコープされる。
type SupportChatMessage struct {
	ID        uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID  uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	StaffID   uint64           `gorm:"column:staff_id;type:bigint;not null" json:"staff_id"`
	Role      string           `gorm:"column:role;type:varchar(16);not null" json:"role"`
	Content   string           `gorm:"column:content;type:text;not null" json:"content"`
	Sources   *json.RawMessage `gorm:"column:sources;type:jsonb" json:"sources,omitempty"`
	CreatedAt time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at,omitempty"`
}

// TableName は SupportChatMessage のテーブル名
func (SupportChatMessage) TableName() string { return "support_chat_messages" }
