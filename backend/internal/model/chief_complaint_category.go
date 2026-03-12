package model

import (
	"time"

	"github.com/google/uuid"
)

// ChiefComplaintCategory は主訴区分マスタ（v11.0追加）
type ChiefComplaintCategory struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ClinicID  uuid.UUID    `gorm:"type:uuid;not null"                             json:"clinic_id"`
	Code      string       `gorm:"default:''"                                     json:"code"`
	Name      string       `gorm:"not null"                                       json:"name"`
	IsActive  bool         `gorm:"default:true"                                   json:"is_active"`
	SortOrder int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ChiefComplaintCategory) TableName() string { return "chief_complaint_categories" }
