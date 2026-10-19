package model

import (
	"time"
)

type ShiftEntryBreak struct {
	ID           uint64      `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ShiftEntryID uint64      `gorm:"column:shift_entry_id;type:bigint;not null" json:"shift_entry_id"`
	StartsAt     *time.Time  `gorm:"column:starts_at;type:timestamptz" json:"-"`
	EndsAt       *time.Time  `gorm:"column:ends_at;type:timestamptz" json:"-"`
	ShiftEntry   *ShiftEntry `gorm:"foreignKey:ShiftEntryID;references:ID" json:"shift_entry,omitempty"`
}

func (ShiftEntryBreak) TableName() string { return "shift_entry_breaks" }
