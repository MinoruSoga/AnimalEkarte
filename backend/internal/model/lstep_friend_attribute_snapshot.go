package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// LstepFriendAttributeSnapshot は Lステップ友だち属性スナップショット。
type LstepFriendAttributeSnapshot struct {
	ID              uint64         `gorm:"column:id;primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64         `gorm:"column:clinic_id;not null"                                json:"clinic_id"`
	LineUserID      string         `gorm:"column:line_user_id;type:varchar(50);not null"            json:"line_user_id"`
	DisplayName     *string        `gorm:"column:display_name;type:varchar(255)"                    json:"display_name,omitempty"`
	RegisteredAt    *time.Time     `gorm:"column:registered_at"                                     json:"registered_at,omitempty"`
	Tags            datatypes.JSON `gorm:"column:tags;type:jsonb"                                   json:"tags,omitempty"`
	Scenarios       datatypes.JSON `gorm:"column:scenarios;type:jsonb"                              json:"scenarios,omitempty"`
	TrafficSource   *string        `gorm:"column:traffic_source;type:varchar(100)"                  json:"traffic_source,omitempty"`
	BlockStatus     *string        `gorm:"column:block_status;type:varchar(20)"                     json:"block_status,omitempty"`
	LastMessageAt   *time.Time     `gorm:"column:last_message_at"                                   json:"last_message_at,omitempty"`
	SnapshotTakenAt time.Time      `gorm:"column:snapshot_taken_at;not null"                        json:"snapshot_taken_at"`
	CsvImportID     *uuid.UUID     `gorm:"column:csv_import_id;type:uuid"                           json:"csv_import_id,omitempty"`
	CreatedAt       time.Time      `gorm:"column:created_at;not null;autoCreateTime"                json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;not null;autoUpdateTime"                json:"updated_at"`
}

func (LstepFriendAttributeSnapshot) TableName() string {
	return "lstep_friend_attribute_snapshots"
}
