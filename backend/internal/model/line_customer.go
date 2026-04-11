package model

import "time"

type LineCustomer struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID         uint64    `gorm:"not null"                                       json:"clinic_id"`
	LineUserID       string    `gorm:"not null"                                       json:"line_user_id"`
	DisplayName      string    `gorm:"not null;default:''"                            json:"display_name"`
	RealName         string    `gorm:"not null;default:''"                            json:"real_name"`
	AdditionalFields []byte    `gorm:"type:jsonb;not null;default:'{}'"               json:"additional_fields"`
	OwnerID          *uint64   `                                                      json:"owner_id,omitempty"`
	CreatedAt        time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner *Owner `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}

func (LineCustomer) TableName() string { return "line_customers" }
