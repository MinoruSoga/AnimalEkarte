package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
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
	ID        uint64                 `gorm:"primaryKey;autoIncrement"     json:"id"`
	ClinicID  uint64                 `gorm:"not null;index"              json:"clinic_id"`
	StaffID   uint64                 `gorm:"not null"                    json:"staff_id"`
	Role      SupportChatMessageRole `gorm:"type:varchar(16);not null"   json:"role"`
	Content   string                 `gorm:"not null;type:text"          json:"content"`
	Sources   json.RawMessage        `gorm:"type:jsonb"                  json:"sources,omitempty"`
	CreatedAt time.Time              `gorm:"autoCreateTime"              json:"created_at"`
	DeletedAt gorm.DeletedAt         `gorm:"index"                       json:"deleted_at,omitempty"`
}

// TableName は SupportChatMessage のテーブル名
func (SupportChatMessage) TableName() string { return "support_chat_messages" }
