package model

import (
	"time"
)

type LstepSendPurposeTagPrefix struct {
	ID          uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Purpose     string    `gorm:"column:purpose;type:varchar(100);not null" json:"purpose"`
	TagPrefix   string    `gorm:"column:tag_prefix;type:varchar(100);not null" json:"tag_prefix"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:NOW();autoUpdateTime" json:"updated_at"`
}

func (LstepSendPurposeTagPrefix) TableName() string { return "lstep_send_purpose_tag_prefixes" }
