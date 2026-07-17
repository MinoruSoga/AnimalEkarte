package model

import "time"

// AppointmentTrimmingDetail は appointments の1:1拡張テーブル（トリミング詳細）
type AppointmentTrimmingDetail struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64         `gorm:"not null"                                       json:"clinic_id"`
	AppointmentID   uint64         `gorm:"uniqueIndex;not null"                           json:"appointment_id"`
	CourseID        *uint64        `                                                      json:"course_id,omitempty"`
	StyleRequest    string         `gorm:"default:''"                                     json:"style_request"`
	BodyWeight      *float64       `gorm:"column:body_weight;type:numeric(6,2)"           json:"body_weight,omitempty"`
	BWUnit          BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'"             json:"bw_unit"`
	BodyTemperature *float64       `gorm:"column:body_temperature;type:numeric(4,1)"      json:"body_temperature,omitempty"`
	UsedShampoo     string         `gorm:"default:''"                                     json:"used_shampoo"`
	UsedRibbon      string         `gorm:"default:''"                                     json:"used_ribbon"`
	Remarks         string         `gorm:"default:''"                                     json:"remarks"`
	StyleImage      string         `gorm:"default:''"                                     json:"style_image"`
	CompletedImage  string         `gorm:"default:''"                                     json:"completed_image"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Course *TrimmingCourse `gorm:"foreignKey:CourseID"                                                                                                          json:"course,omitempty"`
	// Options: foreignKey/references を明示しないと GORM は自モデルの primaryKey(ID) をソース側
	// 結合キーとして使う（デフォルト解決）。本モデルは ID(サロゲートPK) と AppointmentID(実FK) が
	// 別の値空間のため、明示しないと Preload/Association が appointment_trimming_options.appointment_id
	// を appointment_trimming_details.id と突合してしまい、常に空/失敗する（#212）。
	Options []TrimmingOption `gorm:"many2many:appointment_trimming_options;foreignKey:AppointmentID;joinForeignKey:AppointmentID;references:ID;joinReferences:OptionID" json:"options,omitempty"`
}

func (AppointmentTrimmingDetail) TableName() string { return "appointment_trimming_details" }

// AppointmentTrimmingOption は appointments と trimming_options の中間テーブル
type AppointmentTrimmingOption struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	AppointmentID uint64    `gorm:"not null"                 json:"appointment_id"`
	OptionID      uint64    `gorm:"not null"                 json:"option_id"`
	SortOrder     int       `gorm:"default:0"               json:"sort_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime"           json:"created_at"`
}

func (AppointmentTrimmingOption) TableName() string { return "appointment_trimming_options" }
