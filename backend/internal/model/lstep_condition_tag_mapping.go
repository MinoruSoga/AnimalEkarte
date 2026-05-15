package model

import "time"

type LstepConditionTagMapping struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"       json:"id"`
	ConditionCode string    `gorm:"column:condition_code;not null" json:"condition_code"`
	TagName       string    `gorm:"column:tag_name;not null"       json:"tag_name"`
	Description   *string   `gorm:"column:description"             json:"description,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime"                 json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"                 json:"updated_at"`
}

func (LstepConditionTagMapping) TableName() string { return "lstep_condition_tag_mappings" }
