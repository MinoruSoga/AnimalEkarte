package model

import (
	"time"

	"github.com/google/uuid"
)

// MedicalRecord は電子カルテテーブルに対応するモデル
type MedicalRecord struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecordNo              string     `gorm:"column:record_no;uniqueIndex;not null" json:"record_no"`
	Date                  time.Time  `gorm:"column:date;type:date;not null" json:"date"`
	OwnerID               *uuid.UUID `gorm:"column:owner_id;type:uuid" json:"owner_id,omitempty"`
	OwnerName             string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetID                 *uuid.UUID `gorm:"column:pet_id;type:uuid" json:"pet_id,omitempty"`
	PetName               string     `gorm:"column:pet_name;not null" json:"pet_name"`
	Species               string     `gorm:"column:species;type:pet_species;not null" json:"species"`
	ChiefComplaint        string     `gorm:"column:chief_complaint;default:''" json:"chief_complaint"`
	TreatmentPolicy       string     `gorm:"column:treatment_policy;default:''" json:"treatment_policy"`
	PhysicalExam          string     `gorm:"column:physical_exam;default:''" json:"physical_exam"`
	DiagnosisDetails      string     `gorm:"column:diagnosis_details;default:''" json:"diagnosis_details"`
	Diagnosis1CategoryID  *uuid.UUID `gorm:"column:diagnosis1_category_id;type:uuid" json:"diagnosis1_category_id,omitempty"`
	Diagnosis1NameID      *uuid.UUID `gorm:"column:diagnosis1_name_id;type:uuid" json:"diagnosis1_name_id,omitempty"`
	Diagnosis2CategoryID  *uuid.UUID `gorm:"column:diagnosis2_category_id;type:uuid" json:"diagnosis2_category_id,omitempty"`
	Diagnosis2NameID      *uuid.UUID `gorm:"column:diagnosis2_name_id;type:uuid" json:"diagnosis2_name_id,omitempty"`
	Doctor                string     `gorm:"column:doctor;not null" json:"doctor"`
	Status                string     `gorm:"column:status;type:medical_record_status;default:'作成中'" json:"status"`
	CreatedAt             time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	TreatmentItems     []TreatmentItem     `gorm:"foreignKey:MedicalRecordID" json:"treatment_items,omitempty"`
	VitalEntries       []VitalEntry        `gorm:"foreignKey:MedicalRecordID" json:"vital_entries,omitempty"`
	ExaminationRecords []ExaminationRecord `gorm:"foreignKey:MedicalRecordID" json:"examination_records,omitempty"`
	VaccinationRecords []VaccinationRecord `gorm:"foreignKey:MedicalRecordID" json:"vaccination_records,omitempty"`
	CheckupRecords     []CheckupRecord     `gorm:"foreignKey:MedicalRecordID" json:"checkup_records,omitempty"`
}

// TreatmentItem は治療/処置項目テーブルに対応するモデル
type TreatmentItem struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID  `gorm:"column:medical_record_id;type:uuid;not null" json:"medical_record_id"`
	Selected        bool       `gorm:"column:selected;default:false" json:"selected"`
	Status          string     `gorm:"column:status;type:treatment_status;default:'-'" json:"status"`
	Content         string     `gorm:"column:content;not null" json:"content"`
	Memo            string     `gorm:"column:memo;default:''" json:"memo"`
	Insurance       bool       `gorm:"column:insurance;default:false" json:"insurance"`
	UnitPrice       float64    `gorm:"column:unit_price;type:numeric(10,2);default:0" json:"unit_price"`
	Quantity        int        `gorm:"column:quantity;default:1" json:"quantity"`
	DiscountRate    float64    `gorm:"column:discount_rate;type:numeric(5,2);default:0" json:"discount_rate"`
	DiscountAmount  float64    `gorm:"column:discount_amount;type:numeric(10,2);default:0" json:"discount_amount"`
	InventoryID     *uuid.UUID `gorm:"column:inventory_id;type:uuid" json:"inventory_id,omitempty"`
	SortOrder       int        `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// VitalEntry はカルテ内バイタルテーブルに対応するモデル
type VitalEntry struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID `gorm:"column:medical_record_id;type:uuid;not null" json:"medical_record_id"`
	RecordedAt      time.Time `gorm:"column:recorded_at;not null" json:"recorded_at"`
	Staff           string    `gorm:"column:staff;not null" json:"staff"`
	Temperature     *float64  `gorm:"column:temperature;type:numeric(4,1)" json:"temperature,omitempty"`
	HeartRate       *int      `gorm:"column:heart_rate" json:"heart_rate,omitempty"`
	RespirationRate *int      `gorm:"column:respiration_rate" json:"respiration_rate,omitempty"`
	Weight          *float64  `gorm:"column:weight;type:numeric(6,2)" json:"weight,omitempty"`
	Notes           string    `gorm:"column:notes;default:''" json:"notes"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// ExaminationRecord は検査記録テーブルに対応するモデル
type ExaminationRecord struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID `gorm:"column:medical_record_id;type:uuid;not null" json:"medical_record_id"`
	PetID           uuid.UUID `gorm:"column:pet_id;type:uuid;not null" json:"pet_id"`
	Date            time.Time `gorm:"column:date;type:date;not null" json:"date"`
	OwnerName       string    `gorm:"column:owner_name;not null" json:"owner_name"`
	PetName         string    `gorm:"column:pet_name;not null" json:"pet_name"`
	TestType        string    `gorm:"column:test_type;not null" json:"test_type"`
	Doctor          string    `gorm:"column:doctor;not null" json:"doctor"`
	Status          string    `gorm:"column:status;type:examination_status;default:'依頼中'" json:"status"`
	ResultSummary   string    `gorm:"column:result_summary;default:''" json:"result_summary"`
	Machine         string    `gorm:"column:machine;default:''" json:"machine"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	// Relations
	Items []ExaminationRecordItem `gorm:"foreignKey:ExaminationRecordID" json:"items,omitempty"`
}

// ExaminationRecordItem は検査結果項目テーブルに対応するモデル
type ExaminationRecordItem struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ExaminationRecordID uuid.UUID `gorm:"column:examination_record_id;type:uuid;not null" json:"examination_record_id"`
	Name                string    `gorm:"column:name;not null" json:"name"`
	InspectionValue     string    `gorm:"column:inspection_value;default:''" json:"inspection_value"`
	NormalValue         string    `gorm:"column:normal_value;default:''" json:"normal_value"`
	Result              string    `gorm:"column:result;default:''" json:"result"`
	Unit                string    `gorm:"column:unit;default:''" json:"unit"`
	Ref                 string    `gorm:"column:ref;default:''" json:"ref"`
	Status              *string   `gorm:"column:status;type:examination_result_status" json:"status,omitempty"`
	SortOrder           int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// VaccinationRecord は予防接種記録テーブルに対応するモデル
type VaccinationRecord struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID  uuid.UUID  `gorm:"column:medical_record_id;type:uuid;not null" json:"medical_record_id"`
	PetID            uuid.UUID  `gorm:"column:pet_id;type:uuid;not null" json:"pet_id"`
	OwnerName        string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetName          string     `gorm:"column:pet_name;not null" json:"pet_name"`
	VaccineName      string     `gorm:"column:vaccine_name;not null" json:"vaccine_name"`
	Date             time.Time  `gorm:"column:date;type:date;not null" json:"date"`
	NextDate         *time.Time `gorm:"column:next_date;type:date" json:"next_date,omitempty"`
	NextScheduleType *string    `gorm:"column:next_schedule_type;type:next_schedule_type" json:"next_schedule_type,omitempty"`
	Doctor           string     `gorm:"column:doctor;default:''" json:"doctor"`
	Supplemental     string     `gorm:"column:supplemental;default:''" json:"supplemental"`
	Lot1             string     `gorm:"column:lot1;default:''" json:"lot1"`
	Lot2             string     `gorm:"column:lot2;default:''" json:"lot2"`
	Lot3             string     `gorm:"column:lot3;default:''" json:"lot3"`
	Lot4             string     `gorm:"column:lot4;default:''" json:"lot4"`
	Remarks          string     `gorm:"column:remarks;default:''" json:"remarks"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CheckupRecord は定期健診記録テーブルに対応するモデル
type CheckupRecord struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID  `gorm:"column:medical_record_id;type:uuid;not null" json:"medical_record_id"`
	PetID           uuid.UUID  `gorm:"column:pet_id;type:uuid;not null" json:"pet_id"`
	OwnerName       string     `gorm:"column:owner_name;not null" json:"owner_name"`
	PetName         string     `gorm:"column:pet_name;not null" json:"pet_name"`
	CheckupType     string     `gorm:"column:checkup_type;not null" json:"checkup_type"`
	Date            time.Time  `gorm:"column:date;type:date;not null" json:"date"`
	NextDate        *time.Time `gorm:"column:next_date;type:date" json:"next_date,omitempty"`
	Doctor          string     `gorm:"column:doctor;default:''" json:"doctor"`
	Result          string     `gorm:"column:result;default:''" json:"result"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// CreateMedicalRecordRequest は電子カルテ作成リクエスト
type CreateMedicalRecordRequest struct {
	RecordNo             string  `json:"record_no"`
	Date                 string  `json:"date" binding:"required"`
	OwnerID              *string `json:"owner_id"`
	OwnerName            string  `json:"owner_name" binding:"required"`
	PetID                *string `json:"pet_id"`
	PetName              string  `json:"pet_name" binding:"required"`
	Species              string  `json:"species" binding:"required"`
	ChiefComplaint       string  `json:"chief_complaint"`
	TreatmentPolicy      string  `json:"treatment_policy"`
	PhysicalExam         string  `json:"physical_exam"`
	DiagnosisDetails     string  `json:"diagnosis_details"`
	Diagnosis1CategoryID *string `json:"diagnosis1_category_id"`
	Diagnosis1NameID     *string `json:"diagnosis1_name_id"`
	Diagnosis2CategoryID *string `json:"diagnosis2_category_id"`
	Diagnosis2NameID     *string `json:"diagnosis2_name_id"`
	Doctor               string  `json:"doctor" binding:"required"`
	Status               string  `json:"status"`
}

// UpdateMedicalRecordRequest は電子カルテ更新リクエスト
type UpdateMedicalRecordRequest struct {
	Date                 *string `json:"date"`
	OwnerID              *string `json:"owner_id"`
	OwnerName            *string `json:"owner_name"`
	PetID                *string `json:"pet_id"`
	PetName              *string `json:"pet_name"`
	Species              *string `json:"species"`
	ChiefComplaint       *string `json:"chief_complaint"`
	TreatmentPolicy      *string `json:"treatment_policy"`
	PhysicalExam         *string `json:"physical_exam"`
	DiagnosisDetails     *string `json:"diagnosis_details"`
	Diagnosis1CategoryID *string `json:"diagnosis1_category_id"`
	Diagnosis1NameID     *string `json:"diagnosis1_name_id"`
	Diagnosis2CategoryID *string `json:"diagnosis2_category_id"`
	Diagnosis2NameID     *string `json:"diagnosis2_name_id"`
	Doctor               *string `json:"doctor"`
	Status               *string `json:"status"`
}

// CreateVaccinationRecordRequest はワクチン接種記録作成リクエスト
type CreateVaccinationRecordRequest struct {
	MedicalRecordID  string  `json:"medical_record_id" binding:"required"`
	PetID            string  `json:"pet_id" binding:"required"`
	OwnerName        string  `json:"owner_name" binding:"required"`
	PetName          string  `json:"pet_name" binding:"required"`
	VaccineName      string  `json:"vaccine_name" binding:"required"`
	Date             string  `json:"date" binding:"required"`
	NextDate         *string `json:"next_date"`
	NextScheduleType *string `json:"next_schedule_type"`
	Doctor           string  `json:"doctor"`
	Supplemental     string  `json:"supplemental"`
	Lot1             string  `json:"lot1"`
	Lot2             string  `json:"lot2"`
	Lot3             string  `json:"lot3"`
	Lot4             string  `json:"lot4"`
	Remarks          string  `json:"remarks"`
}

// UpdateVaccinationRecordRequest はワクチン接種記録更新リクエスト
type UpdateVaccinationRecordRequest struct {
	VaccineName      *string `json:"vaccine_name"`
	Date             *string `json:"date"`
	NextDate         *string `json:"next_date"`
	NextScheduleType *string `json:"next_schedule_type"`
	Doctor           *string `json:"doctor"`
	Supplemental     *string `json:"supplemental"`
	Lot1             *string `json:"lot1"`
	Lot2             *string `json:"lot2"`
	Lot3             *string `json:"lot3"`
	Lot4             *string `json:"lot4"`
	Remarks          *string `json:"remarks"`
}

// CreateExaminationRecordRequest は検査記録作成リクエスト
type CreateExaminationRecordRequest struct {
	MedicalRecordID string `json:"medical_record_id" binding:"required"`
	PetID           string `json:"pet_id" binding:"required"`
	Date            string `json:"date" binding:"required"`
	OwnerName       string `json:"owner_name" binding:"required"`
	PetName         string `json:"pet_name" binding:"required"`
	TestType        string `json:"test_type" binding:"required"`
	Doctor          string `json:"doctor" binding:"required"`
	Status          string `json:"status"`
	ResultSummary   string `json:"result_summary"`
	Machine         string `json:"machine"`
}

// UpdateExaminationRecordRequest は検査記録更新リクエスト
type UpdateExaminationRecordRequest struct {
	Date          *string `json:"date"`
	TestType      *string `json:"test_type"`
	Doctor        *string `json:"doctor"`
	Status        *string `json:"status"`
	ResultSummary *string `json:"result_summary"`
	Machine       *string `json:"machine"`
}

// PaginatedMedicalRecords はページネーション付き電子カルテレスポンス
type PaginatedMedicalRecords struct {
	Data       []MedicalRecord `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
