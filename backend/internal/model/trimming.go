package model

import (
	"time"
)

// AppointmentTrimmingDetail は appointments の1:1拡張テーブル（トリミング詳細）
type AppointmentTrimmingDetail struct {
	ID              uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID        uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	AppointmentID   uint64          `gorm:"column:appointment_id;type:bigint;not null" json:"appointment_id"`
	StyleRequest    string          `gorm:"column:style_request;type:text;not null;default:''" json:"style_request"`
	BodyWeight      *ExactNumeric   `gorm:"column:body_weight;type:numeric(6,2)" json:"body_weight,omitempty"`
	BWUnit          *BodyWeightUnit `gorm:"column:bw_unit;type:body_weight_unit;default:'Kg'" json:"bw_unit"`
	BodyTemperature *ExactNumeric   `gorm:"column:body_temperature;type:numeric(4,1)" json:"body_temperature,omitempty"`
	UsedShampoo     string          `gorm:"column:used_shampoo;type:text;not null;default:''" json:"used_shampoo"`
	UsedRibbon      string          `gorm:"column:used_ribbon;type:text;not null;default:''" json:"used_ribbon"`
	Remarks         string          `gorm:"column:remarks;type:text;not null;default:''" json:"remarks"`
	StyleImage      string          `gorm:"column:style_image;type:text;not null;default:''" json:"style_image"`
	CompletedImage  string          `gorm:"column:completed_image;type:text;not null;default:''" json:"completed_image"`
	CreatedAt       time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	Course          *TrimmingCourse `gorm:"-"                                                                                                          json:"course,omitempty"`
}

func (AppointmentTrimmingDetail) TableName() string { return "appointment_trimming_details" }

// AppointmentTrimmingOption は appointments と trimming_options の中間テーブル
