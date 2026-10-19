package model

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
	"time"
)

type HospitalizationType string

const (
	HospitalizationTypeInpatient HospitalizationType = "hospitalization"
	HospitalizationTypeHotel     HospitalizationType = "hotel"
)

type HospitalizationStatus string

const (
	HospitalizationStatusAdmitted   HospitalizationStatus = "admitted"
	HospitalizationStatusDischarged HospitalizationStatus = "discharged"
	HospitalizationStatusReserved   HospitalizationStatus = "reserved"
)

type Hospitalization struct {
	ID                   uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID             uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	OwnerID              uint64          `gorm:"column:owner_id;type:bigint;not null" json:"owner_id"`
	PetID                uint64          `gorm:"column:pet_id;type:bigint;not null" json:"pet_id"`
	CageID               *uint64         `gorm:"column:cage_id;type:bigint" json:"cage_id,omitempty"`
	DoctorID             *uint64         `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	Memo                 string          `gorm:"column:memo;type:text;not null;default:''" json:"memo"`
	OwnerRequest         string          `gorm:"column:owner_request;type:text;not null;default:''" json:"owner_request"`
	StaffNotes           string          `gorm:"column:staff_notes;type:text;not null;default:''" json:"staff_notes"`
	InsuranceCompanyName *string         `gorm:"column:insurance_company_name;type:varchar(100)" json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string         `gorm:"column:insurance_number;type:varchar(50)" json:"insurance_number,omitempty"`
	CreatedAt            time.Time       `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	NextVisit            *time.Time      `gorm:"column:next_visit;type:date" json:"next_visit,omitempty"`
	StayCaseID           *uint64         `gorm:"column:stay_case_id;type:bigint" json:"-"`
	Version              int64           `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	Owner                *Owner          `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID"           json:"owner,omitempty"`
	Pet                  *Pet            `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"             json:"pet,omitempty"`
	Cage                 *Cage           `gorm:"foreignKey:CageID,ClinicID;references:ID,ClinicID"            json:"cage,omitempty"`
	Doctor               *Staff          `gorm:"foreignKey:DoctorID;references:ID"          json:"doctor,omitempty"`
	CarePlanItems        []CarePlanItem  `gorm:"foreignKey:HospitalizationID;references:ID" json:"care_plan_items,omitempty"`
	DailyRecords         []DailyRecord   `gorm:"foreignKey:HospitalizationID;references:ID" json:"daily_records,omitempty"`
	TreatmentPlans       []TreatmentPlan `gorm:"foreignKey:HospitalizationID;references:ID" json:"treatment_plans,omitempty"`
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
	ID                    uint64               `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	HospitalizationID     uint64               `gorm:"column:hospitalization_id;type:bigint;not null" json:"hospitalization_id"`
	Type                  CarePlanType         `gorm:"column:type;type:care_plan_type;not null" json:"type"`
	Name                  string               `gorm:"column:name;type:text;not null;default:''" json:"name"`
	Description           string               `gorm:"column:description;type:text;not null;default:''" json:"description"`
	Timing                *pq.StringArray      `gorm:"column:timing;type:plan_timing[];default:'{}'" json:"timing"`
	Status                *CarePlanStatus      `gorm:"column:status;type:care_plan_status;default:'active'" json:"status"`
	Notes                 string               `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	MedicineID            *uint64              `gorm:"column:medicine_id;type:bigint" json:"medicine_id,omitempty"`
	ProcedureID           *uint64              `gorm:"column:procedure_id;type:bigint" json:"procedure_id,omitempty"`
	HospitalizationPlanID *uint64              `gorm:"column:hospitalization_plan_id;type:bigint" json:"hospitalization_plan_id,omitempty"`
	UnitPrice             *int64               `gorm:"column:unit_price;type:bigint;default:0" json:"unit_price"`
	Category              string               `gorm:"column:category;type:text;not null;default:''" json:"category"`
	SortOrder             *int                 `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt             time.Time            `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time            `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	OtherReason           string               `gorm:"column:other_reason;type:text;not null;default:''" json:"other_reason"`
	Medicine              *Medicine            `gorm:"foreignKey:MedicineID;references:ID"            json:"medicine,omitempty"`
	Procedure             *Procedure           `gorm:"foreignKey:ProcedureID;references:ID"           json:"procedure,omitempty"`
	HospitalizationPlan   *HospitalizationPlan `gorm:"foreignKey:HospitalizationPlanID;references:ID" json:"hospitalization_plan,omitempty"`
}

func (CarePlanItem) TableName() string { return "care_plan_items" }

type TreatmentPlan struct {
	ID                uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID          uint64           `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	MedicalRecordID   *uint64          `gorm:"column:medical_record_id;type:bigint" json:"medical_record_id,omitempty"`
	HospitalizationID *uint64          `gorm:"column:hospitalization_id;type:bigint" json:"hospitalization_id,omitempty"`
	TreatmentContent  string           `gorm:"column:treatment_content;type:text;not null;default:''" json:"treatment_content"`
	Memo              string           `gorm:"column:memo;type:text;not null;default:''" json:"memo"`
	IsInsurance       *bool            `gorm:"column:is_insurance;type:boolean;default:false" json:"is_insurance"`
	UnitPrice         *int64           `gorm:"column:unit_price;type:bigint;default:0" json:"unit_price"`
	Quantity          *ExactNumeric    `gorm:"column:quantity;type:numeric(10,1);default:1" json:"quantity"`
	DiscountRate      *ExactNumeric    `gorm:"column:discount_rate;type:numeric(5,2);default:0" json:"discount_rate"`
	DiscountAmount    *int64           `gorm:"column:discount_amount;type:bigint;default:0" json:"discount_amount"`
	SortOrder         *int             `gorm:"column:sort_order;type:integer;default:0" json:"sort_order"`
	CreatedAt         time.Time        `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time        `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	MedicalRecord     *MedicalRecord   `gorm:"foreignKey:MedicalRecordID;references:ID"   json:"medical_record,omitempty"`
	Hospitalization   *Hospitalization `gorm:"foreignKey:HospitalizationID;references:ID" json:"hospitalization,omitempty"`
}

func (TreatmentPlan) TableName() string { return "treatment_plans" }

type DailyRecord struct {
	ID                uint64        `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	HospitalizationID uint64        `gorm:"column:hospitalization_id;type:bigint;not null" json:"hospitalization_id"`
	ClinicID          uint64        `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	Date              time.Time     `gorm:"column:date;type:date;not null" json:"date"`
	CreatedAt         time.Time     `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time     `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	VitalRecords      []VitalRecord `gorm:"foreignKey:DailyRecordID;references:ID" json:"vital_records,omitempty"`
	CareLogs          []CareLog     `gorm:"foreignKey:DailyRecordID;references:ID" json:"care_logs,omitempty"`
	StaffNotes        []StaffNote   `gorm:"foreignKey:DailyRecordID;references:ID" json:"staff_notes,omitempty"`
}

func (DailyRecord) TableName() string { return "daily_records" }

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

type CareLog struct {
	ID            uint64        `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID      uint64        `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	DailyRecordID uint64        `gorm:"column:daily_record_id;type:bigint;not null" json:"daily_record_id"`
	Time          string        `gorm:"column:time;type:time;not null" json:"time"`
	Type          CareLogType   `gorm:"column:type;type:care_log_type;not null" json:"type"`
	Status        CareLogStatus `gorm:"column:status;type:care_log_status;not null;default:'completed'" json:"status"`
	Value         string        `gorm:"column:value;type:text;not null;default:''" json:"value"`
	StaffID       *uint64       `gorm:"column:staff_id;type:bigint" json:"staff_id,omitempty"`
	Notes         string        `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	CreatedAt     time.Time     `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time     `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	Staff         *Staff        `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
}

func (CareLog) TableName() string { return "care_logs" }

type StaffNote struct {
	ID            uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	DailyRecordID uint64    `gorm:"column:daily_record_id;type:bigint;not null" json:"daily_record_id"`
	Time          string    `gorm:"column:time;type:time;not null" json:"time"`
	Content       string    `gorm:"column:content;type:text;not null;default:''" json:"content"`
	StaffID       *uint64   `gorm:"column:staff_id;type:bigint" json:"staff_id,omitempty"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	Staff         *Staff    `gorm:"foreignKey:StaffID;references:ID" json:"staff,omitempty"`
}

func (StaffNote) TableName() string { return "staff_notes" }
