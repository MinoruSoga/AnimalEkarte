package model

import (
	"time"

	"github.com/google/uuid"
)

type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "作成中"
	MedicalRecordStatusFinalized MedicalRecordStatus = "確定済"
)

type MedicalRecord struct {
	ID                   uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecordNo             string              `gorm:"not null;uniqueIndex"                           json:"record_no"`
	Date                 time.Time           `gorm:"type:date;not null"                             json:"date"`
	OwnerID              *uuid.UUID          `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	OwnerName            string              `gorm:"not null;default:''"                            json:"owner_name"`
	PetID                *uuid.UUID          `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	PetName              string              `gorm:"not null;default:''"                            json:"pet_name"`
	Species              PetSpecies          `gorm:"type:pet_species;not null"                      json:"species"`
	ChiefComplaint       string              `gorm:"default:''"                                     json:"chief_complaint"`
	TreatmentPolicy      string              `gorm:"default:''"                                     json:"treatment_policy"`
	PhysicalExam         string              `gorm:"default:''"                                     json:"physical_exam"`
	DiagnosisDetails     string              `gorm:"default:''"                                     json:"diagnosis_details"`
	Diagnosis1CategoryID *uuid.UUID          `gorm:"type:uuid"                                      json:"diagnosis1_category_id,omitempty"`
	Diagnosis1NameID     *uuid.UUID          `gorm:"type:uuid"                                      json:"diagnosis1_name_id,omitempty"`
	Diagnosis2CategoryID *uuid.UUID          `gorm:"type:uuid"                                      json:"diagnosis2_category_id,omitempty"`
	Diagnosis2NameID     *uuid.UUID          `gorm:"type:uuid"                                      json:"diagnosis2_name_id,omitempty"`
	DoctorID             *uuid.UUID          `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	Status               MedicalRecordStatus `gorm:"type:medical_record_status;default:'作成中'"       json:"status"`
	CreatedAt            time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt            time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner              *Owner              `gorm:"foreignKey:OwnerID"              json:"owner,omitempty"`
	Pet                *Pet                `gorm:"foreignKey:PetID"                json:"pet,omitempty"`
	Doctor             *Staff              `gorm:"foreignKey:DoctorID"             json:"doctor,omitempty"`
	Diagnosis1Category *DiagnosisCategory  `gorm:"foreignKey:Diagnosis1CategoryID" json:"diagnosis1_category,omitempty"`
	Diagnosis1Name     *DiagnosisName      `gorm:"foreignKey:Diagnosis1NameID"     json:"diagnosis1_name,omitempty"`
	Diagnosis2Category *DiagnosisCategory  `gorm:"foreignKey:Diagnosis2CategoryID" json:"diagnosis2_category,omitempty"`
	Diagnosis2Name     *DiagnosisName      `gorm:"foreignKey:Diagnosis2NameID"     json:"diagnosis2_name,omitempty"`
	TreatmentItems     []TreatmentItem     `gorm:"foreignKey:MedicalRecordID"      json:"treatment_items,omitempty"`
	VitalEntries       []VitalEntry        `gorm:"foreignKey:MedicalRecordID"      json:"vital_entries,omitempty"`
	ExaminationRecords []ExaminationRecord `gorm:"foreignKey:MedicalRecordID"      json:"examination_records,omitempty"`
	VaccinationRecords []VaccinationRecord `gorm:"foreignKey:MedicalRecordID"      json:"vaccination_records,omitempty"`
	CheckupRecords     []CheckupRecord     `gorm:"foreignKey:MedicalRecordID"      json:"checkup_records,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_records" }

type TreatmentItemType string

const (
	TreatmentItemTypeConsultation TreatmentItemType = "consultation"
	TreatmentItemTypeProcedure    TreatmentItemType = "procedure"
	TreatmentItemTypeMedicine     TreatmentItemType = "medicine"
	TreatmentItemTypeOther        TreatmentItemType = "other"
)

type TreatmentStatus string

const (
	TreatmentStatusIncomplete TreatmentStatus = "未完了"
	TreatmentStatusComplete   TreatmentStatus = "完了"
	TreatmentStatusNA         TreatmentStatus = "-"
)

type TreatmentItem struct {
	ID              uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID         `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	ItemType        TreatmentItemType `gorm:"type:treatment_item_type;not null;default:'other'" json:"item_type"`
	ConsultationID  *uuid.UUID        `gorm:"type:uuid"                                      json:"consultation_id,omitempty"`
	ProcedureID     *uuid.UUID        `gorm:"type:uuid"                                      json:"procedure_id,omitempty"`
	MedicineID      *uuid.UUID        `gorm:"type:uuid"                                      json:"medicine_id,omitempty"`
	Selected        bool              `gorm:"default:false"                                  json:"selected"`
	Status          TreatmentStatus   `gorm:"type:treatment_status;default:'未完了'"            json:"status"`
	Content         string            `gorm:"not null;default:''"                            json:"content"`
	Memo            string            `gorm:"default:''"                                     json:"memo"`
	Insurance       bool              `gorm:"default:false"                                  json:"insurance"`
	UnitPrice       float64           `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Quantity        int               `gorm:"default:1"                                      json:"quantity"`
	DiscountRate    float64           `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount  float64           `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	InventoryID     *uuid.UUID        `gorm:"type:uuid"                                      json:"inventory_id,omitempty"`
	SortOrder       int               `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Consultation *Consultation  `gorm:"foreignKey:ConsultationID" json:"consultation,omitempty"`
	Procedure    *Procedure     `gorm:"foreignKey:ProcedureID"    json:"procedure,omitempty"`
	Medicine     *Medicine      `gorm:"foreignKey:MedicineID"     json:"medicine,omitempty"`
	Inventory    *InventoryItem `gorm:"foreignKey:InventoryID"    json:"inventory,omitempty"`
}

func (TreatmentItem) TableName() string { return "treatment_items" }

type VitalEntry struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MedicalRecordID uuid.UUID  `gorm:"type:uuid;not null"                             json:"medical_record_id"`
	RecordedAt      time.Time  `gorm:"not null;default:now()"                         json:"recorded_at"`
	StaffID         *uuid.UUID `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	Temperature     *float64   `gorm:"type:numeric(4,1)"                              json:"temperature,omitempty"`
	HeartRate       *int       `gorm:"column:heart_rate"                              json:"heart_rate,omitempty"`
	RespirationRate *int       `gorm:"column:respiration_rate"                        json:"respiration_rate,omitempty"`
	Weight          *float64   `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	Notes           string     `gorm:"default:''"                                     json:"notes"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (VitalEntry) TableName() string { return "vital_entries" }
