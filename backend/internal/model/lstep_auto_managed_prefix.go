package model

import (
	"time"
)

type LstepAutoManagedPrefix struct {
	ID          uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	Prefix      string    `gorm:"column:prefix;type:varchar(100);not null" json:"prefix"`
	Category    string    `gorm:"column:category;type:varchar(20);not null" json:"category"`
	Description *string   `gorm:"column:description;type:text" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:NOW();autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:NOW();autoUpdateTime" json:"updated_at"`
}

func (LstepAutoManagedPrefix) TableName() string { return "lstep_auto_managed_prefixes" }
