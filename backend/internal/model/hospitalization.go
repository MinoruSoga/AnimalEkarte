package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type HospitalizationType string

const (
	HospitalizationTypeInpatient HospitalizationType = "入院"
	HospitalizationTypeHotel     HospitalizationType = "ホテル"
)

type HospitalizationStatus string

const (
	HospitalizationStatusAdmitted   HospitalizationStatus = "入院中"
	HospitalizationStatusDischarged HospitalizationStatus = "退院済"
	HospitalizationStatusReserved   HospitalizationStatus = "予約"
)

type Hospitalization struct {
	ID                  uuid.UUID             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID             *uuid.UUID            `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	OwnerName           string                `gorm:"not null;default:''"                            json:"owner_name"`
	PetID               *uuid.UUID            `gorm:"type:uuid"                                      json:"pet_id,omitempty"`
	PetName             string                `gorm:"not null;default:''"                            json:"pet_name"`
	Species             PetSpecies            `gorm:"type:pet_species;not null"                      json:"species"`
	HospitalizationType HospitalizationType   `gorm:"type:hospitalization_type;not null"             json:"hospitalization_type"`
	StartDate           time.Time             `gorm:"type:date;not null"                             json:"start_date"`
	EndDate             time.Time             `gorm:"type:date;not null"                             json:"end_date"`
	Status              HospitalizationStatus `gorm:"type:hospitalization_status;default:'予約'"       json:"status"`
	CageID              *uuid.UUID            `gorm:"type:uuid"                                      json:"cage_id,omitempty"`
	DoctorID            *uuid.UUID            `gorm:"type:uuid"                                      json:"doctor_id,omitempty"`
	Memo                string                `gorm:"default:''"                                     json:"memo"`
	OwnerRequest        string                `gorm:"default:''"                                     json:"owner_request"`
	StaffNotes          string                `gorm:"default:''"                                     json:"staff_notes"`
	CreatedAt           time.Time             `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time             `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner          *Owner          `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet            *Pet            `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	Cage           *Cage           `gorm:"foreignKey:CageID"           json:"cage,omitempty"`
	Doctor         *Staff          `gorm:"foreignKey:DoctorID"         json:"doctor,omitempty"`
	CarePlanItems  []CarePlanItem  `gorm:"foreignKey:HospitalizationID" json:"care_plan_items,omitempty"`
	DailyRecords   []DailyRecord   `gorm:"foreignKey:HospitalizationID" json:"daily_records,omitempty"`
	TreatmentPlans []TreatmentPlan `gorm:"foreignKey:HospitalizationID" json:"treatment_plans,omitempty"`
}

func (Hospitalization) TableName() string { return "hospitalizations" }

type CarePlanType string

const (
	CarePlanTypeFood        CarePlanType = "food"
	CarePlanTypeMedicine    CarePlanType = "medicine"
	CarePlanTypeTreatment   CarePlanType = "treatment"
	CarePlanTypeInstruction CarePlanType = "instruction"
	CarePlanTypeItem        CarePlanType = "item"
)

type CarePlanStatus string

const (
	CarePlanStatusActive       CarePlanStatus = "active"
	CarePlanStatusCompleted    CarePlanStatus = "completed"
	CarePlanStatusDiscontinued CarePlanStatus = "discontinued"
)

type PlanTiming string

const (
	PlanTimingMorning PlanTiming = "morning"
	PlanTimingNoon    PlanTiming = "noon"
	PlanTimingNight   PlanTiming = "night"
)

type CarePlanItem struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID     uuid.UUID      `gorm:"type:uuid;not null"                             json:"hospitalization_id"`
	Type                  CarePlanType   `gorm:"type:care_plan_type;not null"                   json:"type"`
	Name                  string         `gorm:"not null;default:''"                            json:"name"`
	Description           string         `gorm:"default:''"                                     json:"description"`
	Timing                pq.StringArray `gorm:"type:plan_timing[]"                             json:"timing" swaggertype:"array,string"`
	Status                CarePlanStatus `gorm:"type:care_plan_status;default:'active'"         json:"status"`
	Notes                 string         `gorm:"default:''"                                     json:"notes"`
	MedicineID            *uuid.UUID     `gorm:"type:uuid"                                      json:"medicine_id,omitempty"`
	ProcedureID           *uuid.UUID     `gorm:"type:uuid"                                      json:"procedure_id,omitempty"`
	HospitalizationPlanID *uuid.UUID     `gorm:"type:uuid"                                      json:"hospitalization_plan_id,omitempty"`
	UnitPrice             float64        `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Category              string         `gorm:"default:''"                                     json:"category"`
	SortOrder             int            `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt             time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt             time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Medicine            *Medicine            `gorm:"foreignKey:MedicineID"            json:"medicine,omitempty"`
	Procedure           *Procedure           `gorm:"foreignKey:ProcedureID"           json:"procedure,omitempty"`
	HospitalizationPlan *HospitalizationPlan `gorm:"foreignKey:HospitalizationPlanID" json:"hospitalization_plan,omitempty"`
}

func (CarePlanItem) TableName() string { return "care_plan_items" }

type TreatmentPlan struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID uuid.UUID `gorm:"type:uuid;not null"                             json:"hospitalization_id"`
	TreatmentContent  string    `gorm:"not null;default:''"                            json:"treatment_content"`
	Memo              string    `gorm:"default:''"                                     json:"memo"`
	Insurance         bool      `gorm:"default:false"                                  json:"insurance"`
	UnitPrice         float64   `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Quantity          int       `gorm:"default:1"                                      json:"quantity"`
	DiscountRate      float64   `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount    float64   `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	Subtotal          float64   `gorm:"type:numeric(10,2);default:0"                   json:"subtotal"`
	SortOrder         int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt         time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TreatmentPlan) TableName() string { return "treatment_plans" }

type DailyRecord struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	HospitalizationID uuid.UUID `gorm:"type:uuid;not null"                             json:"hospitalization_id"`
	Date              time.Time `gorm:"type:date;not null"                             json:"date"`
	CreatedAt         time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	VitalRecords     []VitalRecord     `gorm:"foreignKey:DailyRecordID" json:"vital_records,omitempty"`
	CareLogRecords   []CareLogRecord   `gorm:"foreignKey:DailyRecordID" json:"care_log_records,omitempty"`
	StaffNoteRecords []StaffNoteRecord `gorm:"foreignKey:DailyRecordID" json:"staff_note_records,omitempty"`
}

func (DailyRecord) TableName() string { return "daily_records" }

type VitalRecord struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID   uuid.UUID  `gorm:"type:uuid;not null"                             json:"daily_record_id"`
	Time            string     `gorm:"not null;default:''"                            json:"time"`
	Temperature     *float64   `gorm:"type:numeric(4,1)"                              json:"temperature,omitempty"`
	HeartRate       *int       `gorm:"column:heart_rate"                              json:"heart_rate,omitempty"`
	RespirationRate *int       `gorm:"column:respiration_rate"                        json:"respiration_rate,omitempty"`
	Weight          *float64   `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	Notes           string     `gorm:"default:''"                                     json:"notes"`
	StaffID         *uuid.UUID `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (VitalRecord) TableName() string { return "vital_records" }

type CareLogType string

const (
	CareLogTypeFood      CareLogType = "food"
	CareLogTypeExcretion CareLogType = "excretion"
	CareLogTypeMedicine  CareLogType = "medicine"
	CareLogTypeTreatment CareLogType = "treatment"
	CareLogTypeOther     CareLogType = "other"
)

type CareLogStatus string

const (
	CareLogStatusCompleted CareLogStatus = "completed"
	CareLogStatusPartial   CareLogStatus = "partial"
	CareLogStatusSkipped   CareLogStatus = "skipped"
)

type CareLogRecord struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID uuid.UUID     `gorm:"type:uuid;not null"                             json:"daily_record_id"`
	Time          string        `gorm:"not null;default:''"                            json:"time"`
	Type          CareLogType   `gorm:"type:care_log_type;not null"                    json:"type"`
	Status        CareLogStatus `gorm:"type:care_log_status;not null;default:'completed'" json:"status"`
	Value         string        `gorm:"default:''"                                     json:"value"`
	StaffID       *uuid.UUID    `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	Notes         string        `gorm:"default:''"                                     json:"notes"`
	CreatedAt     time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (CareLogRecord) TableName() string { return "care_log_records" }

type StaffNoteRecord struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DailyRecordID uuid.UUID  `gorm:"type:uuid;not null"                             json:"daily_record_id"`
	Time          string     `gorm:"not null;default:''"                            json:"time"`
	Content       string     `gorm:"not null;default:''"                            json:"content"`
	StaffID       *uuid.UUID `gorm:"type:uuid"                                      json:"staff_id,omitempty"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (StaffNoteRecord) TableName() string { return "staff_note_records" }
