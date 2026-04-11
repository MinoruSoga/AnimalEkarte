package model

type StaffExcludedReservationCategory struct {
	ID                    uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	StaffID               uint64 `gorm:"not null"                 json:"staff_id"`
	ReservationCategoryID uint64 `gorm:"not null"                 json:"reservation_category_id"`

	// Relations
	Staff               *Staff               `gorm:"foreignKey:StaffID"       json:"staff,omitempty"`
	ReservationCategory *ReservationCategory `gorm:"foreignKey:ReservationCategoryID" json:"reservation_category,omitempty"`
}

func (StaffExcludedReservationCategory) TableName() string { return "staff_reservation_exclusions" }
