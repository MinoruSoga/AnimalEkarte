package model

type ShiftEntryBreak struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ShiftEntryID uint64 `gorm:"not null"                 json:"shift_entry_id"`
	BreakStart   string `gorm:"type:time;not null"       json:"break_start"`
	BreakEnd     string `gorm:"type:time;not null"       json:"break_end"`

	// Relations
	ShiftEntry *ShiftEntry `gorm:"foreignKey:ShiftEntryID" json:"shift_entry,omitempty"`
}

func (ShiftEntryBreak) TableName() string { return "shift_entry_breaks" }
