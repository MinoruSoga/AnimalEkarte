package model

import (
	"time"

	"github.com/google/uuid"
)

// Hospitalization 入院/ホテルモデル
type Hospitalization struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HospitalizationNo string     `json:"hospitalization_no" gorm:"type:varchar(20)"`
	PetID             uuid.UUID  `json:"pet_id" gorm:"type:uuid;not null;index:idx_hosp_pet_id"`
	OwnerID           uuid.UUID  `json:"owner_id" gorm:"type:uuid;not null"`
	CageID            *uuid.UUID `json:"cage_id" gorm:"type:uuid"`
	Type              string     `json:"type" gorm:"type:varchar(20)"` // 入院, ホテル
	StartDate         time.Time  `json:"start_date" gorm:"type:date"`
	EndDate           time.Time  `json:"end_date" gorm:"type:date"`
	Status            string     `json:"status" gorm:"type:varchar(20);index:idx_hosp_status;default:'予約'"` // 入院中, 退院済, 予約, 一時帰宅
	OwnerRequest      string     `json:"owner_request" gorm:"type:text"`
	StaffNotes        string     `json:"staff_notes" gorm:"type:text"`
	Memo              string     `json:"memo" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relations
	Pet           *Pet           `json:"pet,omitempty" gorm:"foreignKey:PetID"`
	Owner         *Owner         `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Cage          *Cage          `json:"cage,omitempty" gorm:"foreignKey:CageID"`
	CarePlanItems []CarePlanItem `json:"care_plan_items,omitempty" gorm:"foreignKey:HospitalizationID"`
	DailyRecords  []DailyRecord  `json:"daily_records,omitempty" gorm:"foreignKey:HospitalizationID"`
}

// TableName テーブル名を指定
func (Hospitalization) TableName() string {
	return "hospitalizations"
}

// Cage ケージマスタモデル
type Cage struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Code        string    `json:"code" gorm:"type:varchar(20)"`
	Name        string    `json:"name" gorm:"type:varchar(100)"`
	Size        string    `json:"size" gorm:"type:varchar(50)"` // S, M, L, XL
	Type        string    `json:"type" gorm:"type:varchar(50)"` // 犬用, 猫用, 共用
	IsAvailable bool      `json:"is_available" gorm:"default:true"`
	Notes       string    `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName テーブル名を指定
func (Cage) TableName() string {
	return "cages"
}

// CarePlanItem ケアプラン項目モデル
type CarePlanItem struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HospitalizationID uuid.UUID  `json:"hospitalization_id" gorm:"type:uuid;not null"`
	MasterID          *uuid.UUID `json:"master_id" gorm:"type:uuid"`
	Type              string     `json:"type" gorm:"type:varchar(30)"` // food, medicine, treatment, instruction, item
	Name              string     `json:"name" gorm:"type:varchar(100)"`
	Description       string     `json:"description" gorm:"type:text"`
	Timing            string     `json:"timing" gorm:"type:json"`                         // JSON array
	Status            string     `json:"status" gorm:"type:varchar(20);default:'active'"` // active, completed, discontinued
	UnitPrice         *float64   `json:"unit_price" gorm:"type:decimal(10,2)"`
	Category          string     `json:"category" gorm:"type:varchar(50)"`
	Notes             string     `json:"notes" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// TableName テーブル名を指定
func (CarePlanItem) TableName() string {
	return "care_plan_items"
}

// DailyRecord 日次記録モデル
type DailyRecord struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HospitalizationID uuid.UUID `json:"hospitalization_id" gorm:"type:uuid;not null"`
	RecordDate        time.Time `json:"record_date" gorm:"type:date"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relations
	Vitals     []Vital     `json:"vitals,omitempty" gorm:"foreignKey:DailyRecordID"`
	CareLogs   []CareLog   `json:"care_logs,omitempty" gorm:"foreignKey:DailyRecordID"`
	StaffNotes []StaffNote `json:"staff_notes,omitempty" gorm:"foreignKey:DailyRecordID"`
}

// TableName テーブル名を指定
func (DailyRecord) TableName() string {
	return "daily_records"
}

// Vital バイタル記録モデル
type Vital struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	DailyRecordID   uuid.UUID  `json:"daily_record_id" gorm:"type:uuid;not null"`
	StaffID         *uuid.UUID `json:"staff_id" gorm:"type:uuid"`
	RecordedTime    string     `json:"recorded_time" gorm:"type:time"`
	Temperature     *float64   `json:"temperature" gorm:"type:decimal(4,1)"`
	HeartRate       *int       `json:"heart_rate"`
	RespirationRate *int       `json:"respiration_rate"`
	Weight          *float64   `json:"weight" gorm:"type:decimal(5,2)"`
	Notes           string     `json:"notes" gorm:"type:text"`
	CreatedAt       time.Time  `json:"created_at"`
}

// TableName テーブル名を指定
func (Vital) TableName() string {
	return "vitals"
}

// CareLog ケアログモデル
type CareLog struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	DailyRecordID uuid.UUID  `json:"daily_record_id" gorm:"type:uuid;not null"`
	StaffID       *uuid.UUID `json:"staff_id" gorm:"type:uuid"`
	RecordedTime  string     `json:"recorded_time" gorm:"type:time"`
	Type          string     `json:"type" gorm:"type:varchar(30)"`   // food, excretion, medicine, treatment, other
	Status        string     `json:"status" gorm:"type:varchar(20)"` // completed, partial, skipped
	Value         string     `json:"value" gorm:"type:varchar(100)"`
	Notes         string     `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TableName テーブル名を指定
func (CareLog) TableName() string {
	return "care_logs"
}

// StaffNote スタッフメモモデル
type StaffNote struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	DailyRecordID uuid.UUID  `json:"daily_record_id" gorm:"type:uuid;not null"`
	StaffID       *uuid.UUID `json:"staff_id" gorm:"type:uuid"`
	RecordedTime  string     `json:"recorded_time" gorm:"type:time"`
	Content       string     `json:"content" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at"`
}

// TableName テーブル名を指定
func (StaffNote) TableName() string {
	return "staff_notes"
}
