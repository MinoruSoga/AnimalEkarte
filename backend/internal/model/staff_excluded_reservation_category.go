package model

// StaffReservationExclusion records which reservation types a staff member cannot handle.
type StaffReservationExclusion struct {
	ID                uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	StaffID           uint64 `gorm:"not null"                 json:"staff_id"`
	ReservationTypeID uint64 `gorm:"not null"                 json:"reservation_type_id"`

	// Relations
	Staff           *Staff           `gorm:"foreignKey:StaffID"            json:"staff,omitempty"`
	ReservationType *ReservationType `gorm:"foreignKey:ReservationTypeID"  json:"reservation_type,omitempty"`
}

func (StaffReservationExclusion) TableName() string { return "staff_reservation_exclusions" }

// StaffExcludedReservationCategory は StaffReservationExclusion の後方互換エイリアス。
type StaffExcludedReservationCategory = StaffReservationExclusion
