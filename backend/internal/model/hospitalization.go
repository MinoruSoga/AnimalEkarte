package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Hospitalization は入院/ホテル記録テーブルに対応するモデル
type Hospitalization struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID             *uuid.UUID `gorm:"column:owner_id;type:uuid" json:"owner_id,omitempty"`
	OwnerName           string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetID               *uuid.UUID `gorm:"column:pet_id;type:uuid" json:"pet_id,omitempty"`
	PetName             string     `gorm:"column:pet_name;not null" json:"pet_name"`
	Species             string     `gorm:"column:species;type:pet_species;not null" json:"species"`
	HospitalizationType string     `gorm:"column:hospitalization_type;type:hospitalization_type;not null" json:"hospitalization_type"`
	StartDate           time.Time  `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate             time.Time  `gorm:"column:end_date;type:date;not null" json:"end_date"`
	Status              string     `gorm:"column:status;type:hospitalization_status;default:'予約'" json:"status"`
	CageID              *uuid.UUID `gorm:"column:cage_id;type:uuid" json:"cage_id,omitempty"`
	DoctorName          *string    `gorm:"column:doctor_name" json:"doctor_name,omitempty"`
	Memo                string     `gorm:"column:memo;default:''" json:"memo"`
	OwnerRequest        string     `gorm:"column:owner_request;default:''" json:"owner_request"`
	StaffNotes          string     `gorm:"column:staff_notes;default:''" json:"staff_notes"`
	CreatedAt           time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	CarePlanItems  []CarePlanItem  `gorm:"foreignKey:HospitalizationID" json:"care_plan_items,omitempty"`
	DailyRecords   []DailyRecord   `gorm:"foreignKey:HospitalizationID" json:"daily_records,omitempty"`
	TreatmentPlans []TreatmentPlan `gorm:"foreignKey:HospitalizationID" json:"treatment_plans,omitempty"`
}

// CarePlanItem はケアプラン項目テーブルに対応するモデル
type CarePlanItem struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID uuid.UUID      `gorm:"column:hospitalization_id;type:uuid;not null" json:"hospitalization_id"`
	Type              string         `gorm:"column:type;type:care_plan_type;not null" json:"type"`
	Name              string         `gorm:"column:name;not null" json:"name"`
	Description       string         `gorm:"column:description;default:''" json:"description"`
	Timing            pq.StringArray `gorm:"column:timing;type:plan_timing[]" json:"timing" swaggertype:"array,string"`
	Status            string         `gorm:"column:status;type:care_plan_status;default:'active'" json:"status"`
	Notes             string         `gorm:"column:notes;default:''" json:"notes"`
	MasterID          *uuid.UUID     `gorm:"column:master_id;type:uuid" json:"master_id,omitempty"`
	UnitPrice         *float64       `gorm:"column:unit_price;type:numeric(10,2)" json:"unit_price,omitempty"`
	Category          *string        `gorm:"column:category" json:"category,omitempty"`
	SortOrder         int            `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// DailyRecord は日次記録コンテナテーブルに対応するモデル
type DailyRecord struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID uuid.UUID `gorm:"column:hospitalization_id;type:uuid;not null" json:"hospitalization_id"`
	Date              time.Time `gorm:"column:date;type:date;not null" json:"date"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	VitalRecords     []VitalRecord     `gorm:"foreignKey:DailyRecordID" json:"vital_records,omitempty"`
	CareLogRecords   []CareLogRecord   `gorm:"foreignKey:DailyRecordID" json:"care_log_records,omitempty"`
	StaffNoteRecords []StaffNoteRecord `gorm:"foreignKey:DailyRecordID" json:"staff_note_records,omitempty"`
}

// VitalRecord は入院バイタルサインテーブルに対応するモデル
type VitalRecord struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID   uuid.UUID `gorm:"column:daily_record_id;type:uuid;not null" json:"daily_record_id"`
	Time            string    `gorm:"column:time;not null" json:"time"`
	Temperature     *float64  `gorm:"column:temperature;type:numeric(4,1)" json:"temperature,omitempty"`
	HeartRate       *int      `gorm:"column:heart_rate" json:"heart_rate,omitempty"`
	RespirationRate *int      `gorm:"column:respiration_rate" json:"respiration_rate,omitempty"`
	Weight          *float64  `gorm:"column:weight;type:numeric(6,2)" json:"weight,omitempty"`
	Notes           string    `gorm:"column:notes;default:''" json:"notes"`
	Staff           string    `gorm:"column:staff;not null" json:"staff"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// CareLogRecord はケアログテーブルに対応するモデル
type CareLogRecord struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID uuid.UUID `gorm:"column:daily_record_id;type:uuid;not null" json:"daily_record_id"`
	Time          string    `gorm:"column:time;not null" json:"time"`
	Type          string    `gorm:"column:type;type:care_log_type;not null" json:"type"`
	Status        string    `gorm:"column:status;type:care_log_status;not null" json:"status"`
	Value         string    `gorm:"column:value;default:''" json:"value"`
	Staff         string    `gorm:"column:staff;not null" json:"staff"`
	Notes         string    `gorm:"column:notes;default:''" json:"notes"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// StaffNoteRecord はスタッフメモテーブルに対応するモデル
type StaffNoteRecord struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID uuid.UUID `gorm:"column:daily_record_id;type:uuid;not null" json:"daily_record_id"`
	Time          string    `gorm:"column:time;not null" json:"time"`
	Content       string    `gorm:"column:content;not null" json:"content"`
	Staff         string    `gorm:"column:staff;not null" json:"staff"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TreatmentPlan は入院治療プランテーブルに対応するモデル
type TreatmentPlan struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID uuid.UUID `gorm:"column:hospitalization_id;type:uuid;not null" json:"hospitalization_id"`
	TreatmentContent  string    `gorm:"column:treatment_content;not null" json:"treatment_content"`
	Memo              string    `gorm:"column:memo;default:''" json:"memo"`
	Insurance         bool      `gorm:"column:insurance;default:false" json:"insurance"`
	UnitPrice         float64   `gorm:"column:unit_price;type:numeric(10,2);default:0" json:"unit_price"`
	Quantity          int       `gorm:"column:quantity;default:1" json:"quantity"`
	DiscountRate      float64   `gorm:"column:discount_rate;type:numeric(5,2);default:0" json:"discount_rate"`
	DiscountAmount    float64   `gorm:"column:discount_amount;type:numeric(10,2);default:0" json:"discount_amount"`
	Subtotal          float64   `gorm:"column:subtotal;type:numeric(10,2);default:0" json:"subtotal"`
	SortOrder         int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateHospitalizationRequest は入院作成リクエスト
type CreateHospitalizationRequest struct {
	OwnerID             *string `json:"owner_id"`
	OwnerName           string  `json:"owner_name" binding:"required"`
	PetID               *string `json:"pet_id"`
	PetName             string  `json:"pet_name" binding:"required"`
	Species             string  `json:"species" binding:"required"`
	HospitalizationType string  `json:"hospitalization_type" binding:"required"`
	StartDate           string  `json:"start_date" binding:"required"`
	EndDate             string  `json:"end_date" binding:"required"`
	Status              string  `json:"status"`
	CageID              *string `json:"cage_id"`
	DoctorName          *string `json:"doctor_name"`
	Memo                string  `json:"memo"`
	OwnerRequest        string  `json:"owner_request"`
	StaffNotes          string  `json:"staff_notes"`
}

// UpdateHospitalizationRequest は入院更新リクエスト
type UpdateHospitalizationRequest struct {
	OwnerID             *string `json:"owner_id"`
	OwnerName           *string `json:"owner_name"`
	PetID               *string `json:"pet_id"`
	PetName             *string `json:"pet_name"`
	Species             *string `json:"species"`
	HospitalizationType *string `json:"hospitalization_type"`
	StartDate           *string `json:"start_date"`
	EndDate             *string `json:"end_date"`
	Status              *string `json:"status"`
	CageID              *string `json:"cage_id"`
	DoctorName          *string `json:"doctor_name"`
	Memo                *string `json:"memo"`
	OwnerRequest        *string `json:"owner_request"`
	StaffNotes          *string `json:"staff_notes"`
}
