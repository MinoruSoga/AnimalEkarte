package model

import "time"

type LstepAutoManagedPrefix struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Prefix      string    `gorm:"column:prefix;not null"   json:"prefix"`
	Category    string    `gorm:"column:category;not null" json:"category"`
	Description *string   `gorm:"column:description"       json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"           json:"updated_at"`
}

func (LstepAutoManagedPrefix) TableName() string { return "lstep_auto_managed_prefixes" }
