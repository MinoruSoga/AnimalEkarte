package model

import (
	"time"

	"github.com/google/uuid"
)

// TrimmingRecord はトリミング記録テーブルに対応するモデル
// DB テーブル名: trimming_records
type TrimmingRecord struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Date           time.Time  `gorm:"column:date;type:date;not null" json:"date"`
	PetID          uuid.UUID  `gorm:"column:pet_id;type:uuid;not null" json:"pet_id"`
	PetNumber      string     `gorm:"column:pet_number;not null" json:"pet_number"`
	PetName        string     `gorm:"column:pet_name;not null" json:"pet_name"`
	OwnerName      string     `gorm:"column:owner_name;not null" json:"owner_name"`
	Species        string     `gorm:"column:species;type:pet_species;not null" json:"species"`
	Weight         string     `gorm:"column:weight;default:''" json:"weight"`
	StyleRequest   string     `gorm:"column:style_request;default:''" json:"style_request"`
	Staff          string     `gorm:"column:staff;not null" json:"staff"`
	Status         string     `gorm:"column:status;type:trimming_status;default:'予約'" json:"status"`
	CourseID       *uuid.UUID `gorm:"column:course_id;type:uuid" json:"course_id,omitempty"`
	Bw             string     `gorm:"column:bw;default:''" json:"bw"`
	BwUnit         string     `gorm:"column:bw_unit;type:body_weight_unit;default:'Kg'" json:"bw_unit"`
	Bt             string     `gorm:"column:bt;default:''" json:"bt"`
	UsedShampoo    string     `gorm:"column:used_shampoo;default:''" json:"used_shampoo"`
	UsedRibbon     string     `gorm:"column:used_ribbon;default:''" json:"used_ribbon"`
	Treatment      string     `gorm:"column:treatment;default:''" json:"treatment"`
	Medicine       string     `gorm:"column:medicine;default:''" json:"medicine"`
	Charge         string     `gorm:"column:charge;default:''" json:"charge"`
	Remarks        string     `gorm:"column:remarks;default:''" json:"remarks"`
	FinalCheck     string     `gorm:"column:final_check;default:''" json:"final_check"`
	StyleImage     *string    `gorm:"column:style_image" json:"style_image,omitempty"`
	CompletedImage *string    `gorm:"column:completed_image" json:"completed_image,omitempty"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Options []TrimmingRecordOption `gorm:"foreignKey:TrimmingRecordID" json:"options,omitempty"`
}

// TableName は GORM に対してテーブル名を明示する
func (TrimmingRecord) TableName() string {
	return "trimming_records"
}

// TrimmingRecordOption はトリミング記録オプション中間テーブルに対応するモデル
type TrimmingRecordOption struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TrimmingRecordID uuid.UUID `gorm:"column:trimming_record_id;type:uuid;not null" json:"trimming_record_id"`
	OptionID         uuid.UUID `gorm:"column:option_id;type:uuid;not null" json:"option_id"`
	SortOrder        int       `gorm:"column:sort_order;default:0" json:"sort_order"`
}

// TableName は GORM に対してテーブル名を明示する
func (TrimmingRecordOption) TableName() string {
	return "trimming_record_options"
}

// CreateTrimmingRequest はトリミング作成リクエスト
type CreateTrimmingRequest struct {
	Date         string   `json:"date" binding:"required"`
	PetID        string   `json:"pet_id" binding:"required"`
	PetNumber    string   `json:"pet_number" binding:"required"`
	PetName      string   `json:"pet_name" binding:"required"`
	OwnerName    string   `json:"owner_name" binding:"required"`
	Species      string   `json:"species" binding:"required"`
	Weight       string   `json:"weight"`
	StyleRequest string   `json:"style_request"`
	Staff        string   `json:"staff" binding:"required"`
	Status       string   `json:"status"`
	CourseID     *string  `json:"course_id"`
	OptionIDs    []string `json:"option_ids"`
	Bw           string   `json:"bw"`
	BwUnit       string   `json:"bw_unit"`
	Bt           string   `json:"bt"`
	UsedShampoo  string   `json:"used_shampoo"`
	UsedRibbon   string   `json:"used_ribbon"`
	Treatment    string   `json:"treatment"`
	Medicine     string   `json:"medicine"`
	Charge       string   `json:"charge"`
	Remarks      string   `json:"remarks"`
	FinalCheck   string   `json:"final_check"`
}

// UpdateTrimmingRequest はトリミング更新リクエスト
type UpdateTrimmingRequest struct {
	Date         *string  `json:"date"`
	PetID        *string  `json:"pet_id"`
	PetNumber    *string  `json:"pet_number"`
	PetName      *string  `json:"pet_name"`
	OwnerName    *string  `json:"owner_name"`
	Species      *string  `json:"species"`
	Weight       *string  `json:"weight"`
	StyleRequest *string  `json:"style_request"`
	Staff        *string  `json:"staff"`
	Status       *string  `json:"status"`
	CourseID     *string  `json:"course_id"`
	OptionIDs    []string `json:"option_ids"`
	Bw           *string  `json:"bw"`
	BwUnit       *string  `json:"bw_unit"`
	Bt           *string  `json:"bt"`
	UsedShampoo  *string  `json:"used_shampoo"`
	UsedRibbon   *string  `json:"used_ribbon"`
	Treatment    *string  `json:"treatment"`
	Medicine     *string  `json:"medicine"`
	Charge       *string  `json:"charge"`
	Remarks      *string  `json:"remarks"`
	FinalCheck   *string  `json:"final_check"`
}
