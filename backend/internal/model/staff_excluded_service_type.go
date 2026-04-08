package model

type StaffExcludedServiceType struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	StaffID       uint64 `gorm:"not null"                 json:"staff_id"`
	ServiceTypeID uint64 `gorm:"not null"                 json:"service_type_id"`

	// Relations
	Staff       *Staff       `gorm:"foreignKey:StaffID"       json:"staff,omitempty"`
	ServiceType *ServiceType `gorm:"foreignKey:ServiceTypeID" json:"service_type,omitempty"`
}

func (StaffExcludedServiceType) TableName() string { return "staff_excluded_service_types" }
