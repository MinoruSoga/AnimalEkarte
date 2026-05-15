package model

import "time"

type LstepSendPurposeTagPrefix struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"   json:"id"`
	Purpose     string    `gorm:"column:purpose;not null"    json:"purpose"`
	TagPrefix   string    `gorm:"column:tag_prefix;not null" json:"tag_prefix"`
	Description *string   `gorm:"column:description"         json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime"             json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"             json:"updated_at"`
}

func (LstepSendPurposeTagPrefix) TableName() string { return "lstep_send_purpose_tag_prefixes" }
