// Package domain contains all entities, interfaces, DTOs, validators, and errors.
package domain

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ---- Owner ----

type MembershipType string

const (
	MembershipTypeNonMember   MembershipType = "non_member"
	MembershipTypeMember      MembershipType = "member"
	MembershipTypeDeceased    MembershipType = "deceased"
	MembershipTypeTransferred MembershipType = "transferred"
)

type Owner struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID       uint64         `gorm:"not null"                                       json:"clinic_id"`
	OwnerName      string         `gorm:"not null"                                       json:"owner_name"`
	OwnerNameKana  string         `gorm:"default:''"                                     json:"owner_name_kana"`
	BirthDate      *time.Time     `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Company        string         `gorm:"default:''"                                     json:"company"`
	PostalCode     string         `gorm:"default:''"                                     json:"postal_code"`
	Address1       string         `gorm:"default:''"                                     json:"address1"`
	Address2       string         `gorm:"default:''"                                     json:"address2"`
	HomePostalCode string         `gorm:"default:''"                                     json:"home_postal_code"`
	HomeAddress1   string         `gorm:"default:''"                                     json:"home_address1"`
	HomeAddress2   string         `gorm:"default:''"                                     json:"home_address2"`
	Phone          string         `gorm:"default:''"                                     json:"phone"`
	CompanyPhone   string         `gorm:"default:''"                                     json:"company_phone"`
	Email          string         `gorm:"default:''"                                     json:"email"`
	Remarks        string         `gorm:"default:''"                                     json:"remarks"`
	IsDangerous    bool           `gorm:"default:false"                                  json:"is_dangerous"`
	DiscountRate   float64        `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	MembershipType MembershipType `gorm:"type:membership_type;default:'non_member'"      json:"membership_type"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt      gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Pets []Pet `gorm:"foreignKey:OwnerID" json:"pets,omitempty"`
}

func (Owner) TableName() string { return "owners" }

// ---- Pet ----

type PetStatus string

const (
	PetStatusAlive    PetStatus = "alive"
	PetStatusDeceased PetStatus = "deceased"
)

type PetGender string

const (
	PetGenderMale    PetGender = "male"
	PetGenderFemale  PetGender = "female"
	PetGenderUnknown PetGender = "unknown"
)

type AcquisitionType string

const (
	AcquisitionTypePurchase  AcquisitionType = "purchased"
	AcquisitionTypeTransfer  AcquisitionType = "transferred"
	AcquisitionTypeProtected AcquisitionType = "rescued"
	AcquisitionTypeOther     AcquisitionType = "other"
)

type DangerLevel string

const (
	DangerLevelLow    DangerLevel = "low"
	DangerLevelMedium DangerLevel = "medium"
	DangerLevelHigh   DangerLevel = "high"
)

type Pet struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64           `gorm:"not null"                                       json:"clinic_id"`
	OwnerID         uint64           `gorm:"not null"                                       json:"owner_id"`
	AnimalSpeciesID uint64           `gorm:"not null"                                       json:"animal_species_id"`
	PetNumber       string           `gorm:"default:''"                                     json:"pet_number"`
	Name            string           `gorm:"not null"                                       json:"name"`
	PetNameKana     string           `gorm:"default:''"                                     json:"pet_name_kana"`
	Gender          PetGender        `gorm:"type:pet_gender;default:'unknown'"               json:"gender"`
	Status          PetStatus        `gorm:"type:pet_status;default:'alive'"                 json:"status"`
	BirthDate       *time.Time       `gorm:"type:date"                                      json:"birth_date,omitempty"`
	Breed           string           `gorm:"default:''"                                     json:"breed"`
	Color           string           `gorm:"default:''"                                     json:"color"`
	Weight          *float64         `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	NeuteredDate    *time.Time       `gorm:"type:date"                                      json:"neutered_date,omitempty"`
	AcquisitionType *AcquisitionType `gorm:"type:acquisition_type"                          json:"acquisition_type,omitempty"`
	DangerLevel     DangerLevel      `gorm:"type:danger_level;default:'low'"                 json:"danger_level"`
	Food            string           `gorm:"default:''"                                     json:"food"`
	Environment     string           `gorm:"default:''"                                     json:"environment"`
	Phone           string           `gorm:"default:''"                                     json:"phone"`
	LastVisit       *time.Time       `gorm:"type:date"                                      json:"last_visit,omitempty"`
	InsuranceID     *uint64          `                                                      json:"insurance_id,omitempty"`
	Remarks         string           `gorm:"default:''"                                     json:"remarks"`
	CreatedAt       time.Time        `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time        `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt   `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Insurance     *Insurance     `gorm:"foreignKey:InsuranceID"      json:"insurance,omitempty"`
	AnimalSpecies *AnimalSpecies `gorm:"foreignKey:AnimalSpeciesID"  json:"animal_species,omitempty"`
}

func (Pet) TableName() string { return "pets" }

// ---- ReservationAppointment ----

type ReservationStatus string

const (
	ReservationStatusConfirmed      ReservationStatus = "confirmed"
	ReservationStatusPending        ReservationStatus = "pending"
	ReservationStatusCancelled      ReservationStatus = "cancelled"
	ReservationStatusCheckedIn      ReservationStatus = "checked_in"
	ReservationStatusInConsultation ReservationStatus = "in_consultation"
	ReservationStatusAccounting     ReservationStatus = "accounting"
	ReservationStatusCompleted      ReservationStatus = "completed"
)

type VisitType string

const (
	VisitTypeFirst   VisitType = "first"
	VisitTypeRevisit VisitType = "revisit"
)

type ReservationAppointment struct {
	ID            uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64            `gorm:"not null"                                       json:"clinic_id"`
	StartTime     time.Time         `gorm:"not null"                                       json:"start_time"`
	EndTime       time.Time         `gorm:"not null"                                       json:"end_time"`
	OwnerID       *uint64           `                                                      json:"owner_id,omitempty"`
	PetID         *uint64           `                                                      json:"pet_id,omitempty"`
	VisitType     VisitType         `gorm:"type:visit_type;not null;default:'revisit'"     json:"visit_type"`
	ServiceTypeID uint64            `gorm:"not null"                                       json:"service_type_id"`
	DoctorID      *uint64           `                                                      json:"doctor_id,omitempty"`
	IsDesignated  bool              `gorm:"default:false"                                  json:"is_designated"`
	Status        ReservationStatus `gorm:"type:reservation_status;default:'pending'"      json:"status"`
	Notes         string            `gorm:"default:''"                                     json:"notes"`
	DeletedAt     gorm.DeletedAt    `                                                      json:"-" swaggerignore:"true"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner       *Owner       `gorm:"foreignKey:OwnerID"       json:"owner,omitempty"`
	Pet         *Pet         `gorm:"foreignKey:PetID"         json:"pet,omitempty"`
	ServiceType *ServiceType `gorm:"foreignKey:ServiceTypeID" json:"service_type,omitempty"`
	Doctor      *Staff       `gorm:"foreignKey:DoctorID"      json:"doctor,omitempty"`
}

func (ReservationAppointment) TableName() string { return "reservation_appointments" }

// ---- MedicalRecord ----

type MedicalRecordStatus string

const (
	MedicalRecordStatusDraft     MedicalRecordStatus = "draft"
	MedicalRecordStatusFinalized MedicalRecordStatus = "finalized"
)

type MedicalRecord struct {
	ID                       uint64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID                 uint64              `gorm:"not null"                                       json:"clinic_id"`
	RecordNo                 string              `gorm:"not null"                                       json:"record_no"`
	Date                     time.Time           `gorm:"type:date;not null"                             json:"date"`
	OwnerID                  *uint64             `                                                      json:"owner_id,omitempty"`
	PetID                    *uint64             `                                                      json:"pet_id,omitempty"`
	DoctorID                 *uint64             `                                                      json:"doctor_id,omitempty"`
	ReservationAppointmentID *uint64             `                                                      json:"reservation_appointment_id,omitempty"`
	Status                   MedicalRecordStatus `gorm:"type:medical_record_status;default:'draft'"      json:"status"`
	CreatedAt                time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt                gorm.DeletedAt      `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"         json:"doctor,omitempty"`
	ClinicalPlan  *ClinicalPlan  `gorm:"foreignKey:MedicalRecordID"  json:"clinical_plan,omitempty"`
	Inquiry       *Inquiry       `gorm:"foreignKey:MedicalRecordID"  json:"inquiry,omitempty"`
	Treatments    []Treatment    `gorm:"foreignKey:MedicalRecordID"  json:"treatments,omitempty"`
	Vitals        []Vital        `gorm:"foreignKey:MedicalRecordID"  json:"vitals,omitempty"`
	Exams         []Examination         `gorm:"foreignKey:MedicalRecordID"  json:"exams,omitempty"`
	Vaccinations  []Vaccination  `gorm:"foreignKey:MedicalRecordID"  json:"vaccinations,omitempty"`
	Checkups      []Checkup      `gorm:"foreignKey:MedicalRecordID"  json:"checkups,omitempty"`
	Estimates     []Estimate     `gorm:"foreignKey:MedicalRecordID"  json:"estimates,omitempty"`
	BillingReview *BillingReview `gorm:"foreignKey:MedicalRecordID"  json:"billing_review,omitempty"`
}

func (MedicalRecord) TableName() string { return "medical_records" }

// ---- ClinicalPlan ----

type ClinicalPlan struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID     uint64    `gorm:"not null;uniqueIndex"                           json:"medical_record_id"`
	PhysicalExam        string    `gorm:"default:''"                                     json:"physical_exam"`
	DiagnosisCategoryID *uint64   `                                                      json:"diagnosis_category_id,omitempty"`
	DiagnosisNameID     *uint64   `                                                      json:"diagnosis_name_id,omitempty"`
	DiagnosisDetails    string    `gorm:"default:''"                                     json:"diagnosis_details"`
	TreatmentPolicy     string    `gorm:"default:''"                                     json:"treatment_policy"`
	CreatedAt           time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord     *MedicalRecord     `gorm:"foreignKey:MedicalRecordID"     json:"medical_record,omitempty"`
	DiagnosisCategory *DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"diagnosis_category,omitempty"`
	DiagnosisName     *DiagnosisName     `gorm:"foreignKey:DiagnosisNameID"     json:"diagnosis_name,omitempty"`
}

func (ClinicalPlan) TableName() string { return "clinical_plans" }

// ---- Inquiry ----

type AppetiteLevel string

const (
	AppetiteLevelNormal    AppetiteLevel = "normal"
	AppetiteLevelIncreased AppetiteLevel = "increased"
	AppetiteLevelDecreased AppetiteLevel = "decreased"
	AppetiteLevelNone      AppetiteLevel = "none"
)

type WaterIntakeLevel string

const (
	WaterIntakeLevelNormal    WaterIntakeLevel = "normal"
	WaterIntakeLevelIncreased WaterIntakeLevel = "increased"
	WaterIntakeLevelDecreased WaterIntakeLevel = "decreased"
	WaterIntakeLevelNone      WaterIntakeLevel = "none"
)

type Inquiry struct {
	ID                       uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID          uint64            `gorm:"not null"                                       json:"medical_record_id"`
	ChiefComplaintCategoryID *uint64           `                                                      json:"chief_complaint_category_id,omitempty"`
	ChiefComplaint           string            `gorm:"default:''"                                     json:"chief_complaint"`
	History                  string            `gorm:"default:''"                                     json:"history"`
	CurrentMedications       string            `gorm:"default:''"                                     json:"current_medications"`
	AllergyInfo              string            `gorm:"default:''"                                     json:"allergy_info"`
	LastMeal                 string            `gorm:"default:''"                                     json:"last_meal"`
	LastDefecation           string            `gorm:"default:''"                                     json:"last_defecation"`
	LastUrination            string            `gorm:"default:''"                                     json:"last_urination"`
	Appetite                 *AppetiteLevel    `gorm:"type:appetite_level"                            json:"appetite,omitempty"`
	WaterIntake              *WaterIntakeLevel `gorm:"type:water_intake_level"                        json:"water_intake,omitempty"`
	OwnerObservations        string            `gorm:"default:''"                                     json:"owner_observations"`
	Notes                    string            `gorm:"default:''"                                     json:"notes"`
	StaffID                  *uint64           `                                                      json:"staff_id,omitempty"`
	CreatedAt                time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt                time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord          *MedicalRecord          `gorm:"foreignKey:MedicalRecordID"          json:"medical_record,omitempty"`
	ChiefComplaintCategory *ChiefComplaintCategory `gorm:"foreignKey:ChiefComplaintCategoryID" json:"chief_complaint_category,omitempty"`
	Staff                  *Staff                  `gorm:"foreignKey:StaffID"                  json:"staff,omitempty"`
}

func (Inquiry) TableName() string { return "inquiries" }

// ---- Treatment ----

type TreatmentItemType string

const (
	TreatmentItemTypeConsultation TreatmentItemType = "consultation"
	TreatmentItemTypeProcedure    TreatmentItemType = "procedure"
	TreatmentItemTypeMedicine     TreatmentItemType = "medicine"
	TreatmentItemTypeOther        TreatmentItemType = "other"
)

type TreatmentStatus string

const (
	TreatmentStatusPending       TreatmentStatus = "pending"
	TreatmentStatusCompleted     TreatmentStatus = "completed"
	TreatmentStatusNotApplicable TreatmentStatus = "not_applicable"
)

type Treatment struct {
	ID              uint64            `gorm:"primaryKey;autoIncrement"                          json:"id"`
	MedicalRecordID uint64            `gorm:"not null"                                          json:"medical_record_id"`
	ItemType        TreatmentItemType `gorm:"type:treatment_item_type;not null;default:'other'" json:"item_type"`
	ConsultationID  *uint64           `                                                         json:"consultation_id,omitempty"`
	ProcedureID     *uint64           `                                                         json:"procedure_id,omitempty"`
	MedicineID      *uint64           `                                                         json:"medicine_id,omitempty"`
	InventoryID     *uint64           `                                                         json:"inventory_id,omitempty"`
	UnitPrice       float64           `gorm:"type:numeric(10,2);default:0"                      json:"unit_price"`
	Quantity        int               `gorm:"default:1"                                         json:"quantity"`
	Selected        bool              `gorm:"default:false"                                     json:"selected"`
	Status          TreatmentStatus   `gorm:"type:treatment_status;default:'pending'"             json:"status"`
	Content         string            `gorm:"not null;default:''"                               json:"content"`
	Memo            string            `gorm:"default:''"                                        json:"memo"`
	Insurance       bool              `gorm:"default:false"                                     json:"insurance"`
	DiscountRate    float64           `gorm:"type:numeric(5,2);default:0"                       json:"discount_rate"`
	DiscountAmount  float64           `gorm:"type:numeric(10,2);default:0"                      json:"discount_amount"`
	SortOrder       int               `gorm:"default:0"                                         json:"sort_order"`
	DeletedAt       gorm.DeletedAt    `                                                         json:"-" swaggerignore:"true"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                    json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                    json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Consultation  *Consultation  `gorm:"foreignKey:ConsultationID"  json:"consultation,omitempty"`
	Procedure     *Procedure     `gorm:"foreignKey:ProcedureID"     json:"procedure,omitempty"`
	Medicine      *Medicine      `gorm:"foreignKey:MedicineID"      json:"medicine,omitempty"`
	Inventory     *InventoryItem `gorm:"foreignKey:InventoryID"     json:"inventory,omitempty"`
}

func (Treatment) TableName() string { return "treatments" }

// ---- Vital ----

type Vital struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64    `gorm:"not null"                                       json:"medical_record_id"`
	RecordedAt      time.Time `gorm:"not null;default:now()"                         json:"recorded_at"`
	StaffID         *uint64   `                                                      json:"staff_id,omitempty"`
	Temperature     *float64  `gorm:"type:numeric(4,1)"                              json:"temperature,omitempty"`
	HeartRate       *int      `gorm:"column:heart_rate"                              json:"heart_rate,omitempty"`
	RespirationRate *int      `gorm:"column:respiration_rate"                        json:"respiration_rate,omitempty"`
	Weight          *float64  `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	Notes           string    `gorm:"default:''"                                     json:"notes"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Staff         *Staff         `gorm:"foreignKey:StaffID"         json:"staff,omitempty"`
}

func (Vital) TableName() string { return "vitals" }

// ---- Checkup ----

type Checkup struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64         `gorm:"not null"                                       json:"medical_record_id"`
	PetID           *uint64        `                                                      json:"pet_id,omitempty"`
	CheckupTypeID   uint64         `gorm:"not null"                                       json:"checkup_type_id"`
	Date            time.Time      `gorm:"type:date;not null"                             json:"date"`
	DoctorID        *uint64        `                                                      json:"doctor_id,omitempty"`
	Result          string         `gorm:"default:''"                                     json:"result"`
	NextDate        *time.Time     `gorm:"type:date"                                      json:"next_date,omitempty"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	CheckupType   *CheckupType   `gorm:"foreignKey:CheckupTypeID"   json:"checkup_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Checkup) TableName() string { return "checkups" }

// ---- Estimate ----

type EstimateStatus string

const (
	EstimateStatusDraft    EstimateStatus = "draft"
	EstimateStatusSent     EstimateStatus = "sent"
	EstimateStatusApproved EstimateStatus = "approved"
	EstimateStatusRejected EstimateStatus = "rejected"
)

type Estimate struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64         `gorm:"not null"                                       json:"clinic_id"`
	EstimateNo      string         `gorm:"not null;default:''"                            json:"estimate_no"`
	MedicalRecordID *uint64        `                                                      json:"medical_record_id,omitempty"`
	Title           string         `gorm:"default:''"                                     json:"title"`
	OwnerID         *uint64        `                                                      json:"owner_id,omitempty"`
	Status          EstimateStatus `gorm:"type:estimate_status;default:'draft'"           json:"status"`
	Subtotal        float64        `gorm:"type:numeric(10,2);default:0"                   json:"subtotal"`
	TaxTotal        float64        `gorm:"type:numeric(10,2);default:0"                   json:"tax_total"`
	TotalAmount     float64        `gorm:"type:numeric(10,2);default:0"                   json:"total_amount"`
	InsuranceAmount float64        `gorm:"type:numeric(10,2);default:0"                   json:"insurance_amount"`
	DiscountAmount  float64        `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	ValidUntil      *time.Time     `gorm:"type:date"                                      json:"valid_until,omitempty"`
	Comment         string         `gorm:"default:''"                                     json:"comment"`
	Notes           string         `gorm:"default:''"                                     json:"notes"`
	CreatedBy       *uint64        `                                                      json:"created_by,omitempty"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Owner         *Owner         `gorm:"foreignKey:OwnerID"         json:"owner,omitempty"`
	CreatedStaff  *Staff         `gorm:"foreignKey:CreatedBy"       json:"created_staff,omitempty"`
	Items         []EstimateItem `gorm:"foreignKey:EstimateID"      json:"items,omitempty"`
}

func (Estimate) TableName() string { return "estimates" }

type EstimateItem struct {
	ID                    uint64       `gorm:"primaryKey;autoIncrement"                       json:"id"`
	EstimateID            uint64       `gorm:"not null"                                       json:"estimate_id"`
	Name                  string       `gorm:"not null;default:''"                            json:"name"`
	Category              ItemCategory `gorm:"type:item_category;not null"                    json:"category"`
	UnitPrice             float64      `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Quantity              int          `gorm:"default:1"                                      json:"quantity"`
	TaxRate               float64      `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	DiscountRate          float64      `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount        float64      `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	IsInsuranceApplicable bool         `gorm:"default:false"                                  json:"is_insurance_applicable"`
	ConsultationID        *uint64      `                                                      json:"consultation_id,omitempty"`
	ProcedureID           *uint64      `                                                      json:"procedure_id,omitempty"`
	MedicineID            *uint64      `                                                      json:"medicine_id,omitempty"`
	SortOrder             int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt             time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Consultation *Consultation `gorm:"foreignKey:ConsultationID" json:"consultation,omitempty"`
	Procedure    *Procedure    `gorm:"foreignKey:ProcedureID"    json:"procedure,omitempty"`
	Medicine     *Medicine     `gorm:"foreignKey:MedicineID"     json:"medicine,omitempty"`
}

func (EstimateItem) TableName() string { return "estimate_items" }

// ---- BillingReview ----

type BillingReviewStatus string

const (
	BillingReviewStatusPending   BillingReviewStatus = "pending"
	BillingReviewStatusConfirmed BillingReviewStatus = "confirmed"
	BillingReviewStatusReturned  BillingReviewStatus = "returned"
)

type BillingReview struct {
	ID              uint64              `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64              `gorm:"not null;uniqueIndex"                           json:"medical_record_id"`
	Status          BillingReviewStatus `gorm:"type:billing_review_status;default:'pending'"   json:"status"`
	ConfirmedBy     *uint64             `                                                      json:"confirmed_by,omitempty"`
	ConfirmedAt     *time.Time          `gorm:"column:confirmed_at"                            json:"confirmed_at,omitempty"`
	ReturnedBy      *uint64             `                                                      json:"returned_by,omitempty"`
	ReturnedAt      *time.Time          `gorm:"column:returned_at"                             json:"returned_at,omitempty"`
	ReturnReason    string              `gorm:"default:''"                                     json:"return_reason"`
	Memo            string              `gorm:"default:''"                                     json:"memo"`
	CreatedAt       time.Time           `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord  *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	ConfirmedStaff *Staff         `gorm:"foreignKey:ConfirmedBy"     json:"confirmed_staff,omitempty"`
	ReturnedStaff  *Staff         `gorm:"foreignKey:ReturnedBy"      json:"returned_staff,omitempty"`
}

func (BillingReview) TableName() string { return "billing_reviews" }

// ---- Hospitalization ----

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
	ID                  uint64                `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID            uint64                `gorm:"not null"                                       json:"clinic_id"`
	OwnerID             uint64                `gorm:"not null"                                       json:"owner_id"`
	PetID               uint64                `gorm:"not null"                                       json:"pet_id"`
	HospitalizationType HospitalizationType   `gorm:"type:hospitalization_type;not null"             json:"hospitalization_type"`
	StartDate           time.Time             `gorm:"type:date;not null"                             json:"start_date"`
	EndDate             time.Time             `gorm:"type:date;not null"                             json:"end_date"`
	Status              HospitalizationStatus `gorm:"type:hospitalization_status;default:'reserved'"  json:"status"`
	CageID              *uint64               `                                                      json:"cage_id,omitempty"`
	DoctorID            *uint64               `                                                      json:"doctor_id,omitempty"`
	Memo                string                `gorm:"default:''"                                     json:"memo"`
	OwnerRequest        string                `gorm:"default:''"                                     json:"owner_request"`
	StaffNotes          string                `gorm:"default:''"                                     json:"staff_notes"`
	CreatedAt           time.Time             `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time             `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt           gorm.DeletedAt        `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Owner          *Owner          `gorm:"foreignKey:OwnerID"           json:"owner,omitempty"`
	Pet            *Pet            `gorm:"foreignKey:PetID"             json:"pet,omitempty"`
	Cage           *Cage           `gorm:"foreignKey:CageID"            json:"cage,omitempty"`
	Doctor         *Staff          `gorm:"foreignKey:DoctorID"          json:"doctor,omitempty"`
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
	ID                    uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	HospitalizationID     uint64         `gorm:"not null"                                       json:"hospitalization_id"`
	Type                  CarePlanType   `gorm:"type:care_plan_type;not null"                   json:"type"`
	Name                  string         `gorm:"not null;default:''"                            json:"name"`
	Description           string         `gorm:"default:''"                                     json:"description"`
	Timing                pq.StringArray `gorm:"type:plan_timing[]"                             json:"timing" swaggertype:"array,string"`
	Status                CarePlanStatus `gorm:"type:care_plan_status;default:'active'"         json:"status"`
	Notes                 string         `gorm:"default:''"                                     json:"notes"`
	MedicineID            *uint64        `                                                      json:"medicine_id,omitempty"`
	ProcedureID           *uint64        `                                                      json:"procedure_id,omitempty"`
	HospitalizationPlanID *uint64        `                                                      json:"hospitalization_plan_id,omitempty"`
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
	ID                uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID   *uint64        `                                                      json:"medical_record_id,omitempty"`
	HospitalizationID *uint64        `                                                      json:"hospitalization_id,omitempty"`
	TreatmentContent  string         `gorm:"not null;default:''"                            json:"treatment_content"`
	Memo              string         `gorm:"default:''"                                     json:"memo"`
	Insurance         bool           `gorm:"default:false"                                  json:"insurance"`
	UnitPrice         float64        `gorm:"type:numeric(10,2);default:0"                   json:"unit_price"`
	Quantity          int            `gorm:"default:1"                                      json:"quantity"`
	DiscountRate      float64        `gorm:"type:numeric(5,2);default:0"                    json:"discount_rate"`
	DiscountAmount    float64        `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	Subtotal          float64        `gorm:"type:numeric(10,2);default:0"                   json:"subtotal"`
	SortOrder         int            `gorm:"default:0"                                      json:"sort_order"`
	DeletedAt         gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord   *MedicalRecord   `gorm:"foreignKey:MedicalRecordID"   json:"medical_record,omitempty"`
	Hospitalization *Hospitalization `gorm:"foreignKey:HospitalizationID" json:"hospitalization,omitempty"`
}

func (TreatmentPlan) TableName() string { return "treatment_plans" }

type DailyRecord struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	HospitalizationID uint64    `gorm:"not null"                                       json:"hospitalization_id"`
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
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	DailyRecordID   uint64    `gorm:"not null"                                       json:"daily_record_id"`
	Time            string    `gorm:"not null;default:''"                            json:"time"`
	Temperature     *float64  `gorm:"type:numeric(4,1)"                              json:"temperature,omitempty"`
	HeartRate       *int      `gorm:"column:heart_rate"                              json:"heart_rate,omitempty"`
	RespirationRate *int      `gorm:"column:respiration_rate"                        json:"respiration_rate,omitempty"`
	Weight          *float64  `gorm:"type:numeric(6,2)"                              json:"weight,omitempty"`
	Notes           string    `gorm:"default:''"                                     json:"notes"`
	StaffID         *uint64   `                                                      json:"staff_id,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`

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
	ID            uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	DailyRecordID uint64        `gorm:"not null"                                       json:"daily_record_id"`
	Time          string        `gorm:"not null;default:''"                            json:"time"`
	Type          CareLogType   `gorm:"type:care_log_type;not null"                    json:"type"`
	Status        CareLogStatus `gorm:"type:care_log_status;not null;default:'completed'" json:"status"`
	Value         string        `gorm:"default:''"                                     json:"value"`
	StaffID       *uint64       `                                                      json:"staff_id,omitempty"`
	Notes         string        `gorm:"default:''"                                     json:"notes"`
	CreatedAt     time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (CareLogRecord) TableName() string { return "care_log_records" }

type StaffNoteRecord struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	DailyRecordID uint64    `gorm:"not null"                                       json:"daily_record_id"`
	Time          string    `gorm:"not null;default:''"                            json:"time"`
	Content       string    `gorm:"not null;default:''"                            json:"content"`
	StaffID       *uint64   `                                                      json:"staff_id,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime"                                 json:"created_at"`

	// Relations
	Staff *Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (StaffNoteRecord) TableName() string { return "staff_note_records" }

// ---- Billing / Accounting ----

type BillingStatus string

const (
	BillingStatusWaiting   BillingStatus = "waiting"
	BillingStatusCompleted BillingStatus = "completed"
	BillingStatusCancelled BillingStatus = "cancelled"
	BillingStatusPending   BillingStatus = "pending"
)

type PaymentMethod string

const (
	PaymentMethodCash            PaymentMethod = "cash"
	PaymentMethodCreditCard      PaymentMethod = "credit_card"
	PaymentMethodElectronicMoney PaymentMethod = "electronic_money"
)

type ItemCategory string

const (
	ItemCategoryExamination ItemCategory = "examination"
	ItemCategoryTest        ItemCategory = "test"
	ItemCategoryProcedure   ItemCategory = "procedure"
	ItemCategorySurgery     ItemCategory = "surgery"
	ItemCategoryMedicine    ItemCategory = "medicine"
	ItemCategoryFood        ItemCategory = "food"
	ItemCategoryGoods       ItemCategory = "goods"
	ItemCategoryOther       ItemCategory = "other"
)

type ItemSource string

const (
	ItemSourceMedicalRecord   ItemSource = "medical_record"
	ItemSourceManual          ItemSource = "manual"
	ItemSourceHospitalization ItemSource = "hospitalization"
)

type Billing struct {
	ID                uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64         `gorm:"not null"                                       json:"clinic_id"`
	MedicalRecordID   *uint64        `                                                      json:"medical_record_id,omitempty"`
	HospitalizationID *uint64        `                                                      json:"hospitalization_id,omitempty"`
	OwnerID           *uint64        `                                                      json:"owner_id,omitempty"`
	PetID             *uint64        `                                                      json:"pet_id,omitempty"`
	Subtotal          int            `gorm:"default:0"                                      json:"subtotal"`
	TaxTotal          int            `gorm:"default:0"                                      json:"tax_total"`
	TotalAmount       int            `gorm:"default:0"                                      json:"total_amount"`
	HasInsurance      bool           `gorm:"default:false"                                  json:"has_insurance"`
	Status            BillingStatus  `gorm:"type:billing_status;default:'waiting'"          json:"status"`
	ScheduledDate     time.Time      `gorm:"type:date;not null"                             json:"scheduled_date"`
	CompletedAt       *time.Time     `                                                      json:"completed_at,omitempty"`
	Memo              string         `gorm:"default:''"                                     json:"memo"`
	DeletedAt         gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt         time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Owner         *Owner         `gorm:"foreignKey:OwnerID"          json:"owner,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"            json:"pet,omitempty"`
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID"  json:"medical_record,omitempty"`
	Items         []BillingItem  `gorm:"foreignKey:BillingID"        json:"items,omitempty"`
	Payments      []Payment      `gorm:"foreignKey:BillingID"        json:"payments,omitempty"`
}

func (Billing) TableName() string { return "billings" }

type BillingItem struct {
	ID                    uint64       `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID             uint64       `gorm:"not null"                                       json:"billing_id"`
	Category              ItemCategory `gorm:"type:item_category;not null"                    json:"category"`
	Name                  string       `gorm:"not null;default:''"                            json:"name"`
	UnitPrice             float64      `gorm:"type:numeric(10,2);not null;default:0"          json:"unit_price"`
	Quantity              int          `gorm:"not null;default:1"                             json:"quantity"`
	TaxRate               float64      `gorm:"type:numeric(3,2);default:0.10"                 json:"tax_rate"`
	IsInsuranceApplicable bool         `gorm:"default:false"                                  json:"is_insurance_applicable"`
	Source                ItemSource   `gorm:"type:item_source;default:'manual'"              json:"source"`
	SortOrder             int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt             time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (BillingItem) TableName() string { return "billing_items" }

type Payment struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	BillingID       uint64        `gorm:"not null;uniqueIndex"                           json:"billing_id"`
	Subtotal        float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"subtotal"`
	TaxTotal        float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"tax_total"`
	TotalAmount     float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"total_amount"`
	InsuranceName   string        `gorm:"default:''"                                     json:"insurance_name"`
	InsuranceRatio  float64       `gorm:"type:numeric(3,2);default:0"                    json:"insurance_ratio"`
	InsuranceAmount float64       `gorm:"type:numeric(10,2);default:0"                   json:"insurance_amount"`
	DiscountAmount  float64       `gorm:"type:numeric(10,2);default:0"                   json:"discount_amount"`
	BillingAmount   float64       `gorm:"type:numeric(10,2);not null;default:0"          json:"billing_amount"`
	ReceivedAmount  float64       `gorm:"type:numeric(10,2);default:0"                   json:"received_amount"`
	ChangeAmount    float64       `gorm:"type:numeric(10,2);default:0"                   json:"change_amount"`
	Method          PaymentMethod `gorm:"type:payment_method;default:'cash'"             json:"method"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }

// ---- Trimming ----

type TrimmingStatus string

const (
	TrimmingStatusCompleted  TrimmingStatus = "completed"
	TrimmingStatusReserved   TrimmingStatus = "reserved"
	TrimmingStatusInProgress TrimmingStatus = "in_progress"
)

type BodyWeightUnit string

const (
	BodyWeightUnitKg BodyWeightUnit = "Kg"
	BodyWeightUnitG  BodyWeightUnit = "g"
)

type TrimmingRecord struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID       uint64         `gorm:"not null"                                       json:"clinic_id"`
	Date           time.Time      `gorm:"type:date;not null"                             json:"date"`
	PetID          *uint64        `                                                      json:"pet_id,omitempty"`
	StaffID        *uint64        `                                                      json:"staff_id,omitempty"`
	CourseID       *uint64        `                                                      json:"course_id,omitempty"`
	Weight         string         `gorm:"default:''"                                     json:"weight"`
	Status         TrimmingStatus `gorm:"type:trimming_status;default:'reserved'"         json:"status"`
	StyleRequest   string         `gorm:"default:''"                                     json:"style_request"`
	BW             string         `gorm:"default:''"                                     json:"bw"`
	BWUnit         BodyWeightUnit `gorm:"type:body_weight_unit;default:'Kg'"             json:"bw_unit"`
	BT             string         `gorm:"default:''"                                     json:"bt"`
	UsedShampoo    string         `gorm:"default:''"                                     json:"used_shampoo"`
	UsedRibbon     string         `gorm:"default:''"                                     json:"used_ribbon"`
	Remarks        string         `gorm:"default:''"                                     json:"remarks"`
	StyleImage     string         `gorm:"default:''"                                     json:"style_image"`
	CompletedImage string         `gorm:"default:''"                                     json:"completed_image"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt      gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Pet     *Pet             `gorm:"foreignKey:PetID"    json:"pet,omitempty"`
	Staff   *Staff           `gorm:"foreignKey:StaffID"  json:"staff,omitempty"`
	Course  *TrimmingCourse  `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Options []TrimmingOption `gorm:"many2many:trimming_record_options;joinForeignKey:TrimmingRecordID;joinReferences:OptionID" json:"options,omitempty"`
}

func (TrimmingRecord) TableName() string { return "trimming_records" }

type TrimmingRecordOption struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement"                       json:"id"`
	TrimmingRecordID uint64 `gorm:"not null"                                       json:"trimming_record_id"`
	OptionID         uint64 `gorm:"not null"                                       json:"option_id"`
	SortOrder        int    `gorm:"default:0"                                      json:"sort_order"`
}

func (TrimmingRecordOption) TableName() string { return "trimming_record_options" }

// ---- Examination ----

type ExaminationStatus string

const (
	ExaminationStatusPending    ExaminationStatus = "pending"
	ExaminationStatusInProgress ExaminationStatus = "in_progress"
	ExaminationStatusCompleted  ExaminationStatus = "completed"
)

type ExaminationResultStatus string

const (
	ExaminationResultStatusNormal ExaminationResultStatus = "normal"
	ExaminationResultStatusHigh   ExaminationResultStatus = "high"
	ExaminationResultStatusLow    ExaminationResultStatus = "low"
)

type Examination struct {
	ID              uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID uint64            `gorm:"not null"                                       json:"medical_record_id"`
	PetID           *uint64           `                                                      json:"pet_id,omitempty"`
	ExamTypeID      uint64            `gorm:"not null"                                       json:"exam_type_id"`
	DoctorID        *uint64           `                                                      json:"doctor_id,omitempty"`
	Date            time.Time         `gorm:"type:date;not null"                             json:"date"`
	ResultSummary   string            `gorm:"default:''"                                     json:"result_summary"`
	Machine         string            `gorm:"default:''"                                     json:"machine"`
	Status          ExaminationStatus `gorm:"type:examination_status;default:'pending'"        json:"status"`
	DeletedAt       gorm.DeletedAt    `                                                      json:"-" swaggerignore:"true"`
	CreatedAt       time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	ExaminationType      *ExaminationType      `gorm:"foreignKey:ExamTypeID"      json:"exam_type,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
	Items         []ExaminationItem     `gorm:"foreignKey:ExamID"          json:"items,omitempty"`
}

func (Examination) TableName() string { return "exams" }

type ExaminationItem struct {
	ID              uint64                  `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamID          uint64                  `gorm:"not null"                                       json:"exam_id"`
	Name            string                  `gorm:"not null;default:''"                            json:"name"`
	InspectionValue string                  `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string                  `gorm:"default:''"                                     json:"normal_value"`
	Result          string                  `gorm:"default:''"                                     json:"result"`
	Unit            string                  `gorm:"default:''"                                     json:"unit"`
	Ref             string                  `gorm:"default:''"                                     json:"ref"`
	Status          ExaminationResultStatus `gorm:"type:examination_result_status;default:'normal'" json:"status"`
	SortOrder       int                     `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time               `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (ExaminationItem) TableName() string { return "exam_items" }

// ---- Vaccination ----

type NextScheduleType string

const (
	NextScheduleType3Weeks NextScheduleType = "3weeks"
	NextScheduleType4Weeks NextScheduleType = "4weeks"
	NextScheduleType1Year  NextScheduleType = "1year"
	NextScheduleTypeOther  NextScheduleType = "other"
)

type Vaccination struct {
	ID               uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	MedicalRecordID  uint64            `gorm:"not null"                                       json:"medical_record_id"`
	PetID            *uint64           `                                                      json:"pet_id,omitempty"`
	VaccineID        uint64            `gorm:"not null"                                       json:"vaccine_id"`
	Date             time.Time         `gorm:"type:date;not null"                             json:"date"`
	DoctorID         *uint64           `                                                      json:"doctor_id,omitempty"`
	NextDate         *time.Time        `gorm:"type:date"                                      json:"next_date,omitempty"`
	NextScheduleType *NextScheduleType `gorm:"type:next_schedule_type"                       json:"next_schedule_type,omitempty"`
	Supplemental     string            `gorm:"default:''"                                     json:"supplemental"`
	Lot1             string            `gorm:"default:''"                                     json:"lot1"`
	Lot2             string            `gorm:"default:''"                                     json:"lot2"`
	Lot3             string            `gorm:"default:''"                                     json:"lot3"`
	Lot4             string            `gorm:"default:''"                                     json:"lot4"`
	Remarks          string            `gorm:"default:''"                                     json:"remarks"`
	DeletedAt        gorm.DeletedAt    `                                                      json:"-" swaggerignore:"true"`
	CreatedAt        time.Time         `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt        time.Time         `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	MedicalRecord *MedicalRecord `gorm:"foreignKey:MedicalRecordID" json:"medical_record,omitempty"`
	Pet           *Pet           `gorm:"foreignKey:PetID"           json:"pet,omitempty"`
	Vaccine       *Vaccine       `gorm:"foreignKey:VaccineID"       json:"vaccine,omitempty"`
	Doctor        *Staff         `gorm:"foreignKey:DoctorID"        json:"doctor,omitempty"`
}

func (Vaccination) TableName() string { return "vaccinations" }

// ---- Inventory ----

type InventoryCategory string

const (
	InventoryCategoryMedicine   InventoryCategory = "medicine"
	InventoryCategoryConsumable InventoryCategory = "consumable"
	InventoryCategoryFood       InventoryCategory = "food"
	InventoryCategoryOther      InventoryCategory = "other"
)

type InventoryStatus string

const (
	InventoryStatusSufficient InventoryStatus = "sufficient"
	InventoryStatusLow        InventoryStatus = "low"
	InventoryStatusOutOfStock InventoryStatus = "out_of_stock"
)

type InventoryItem struct {
	ID            uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64            `gorm:"not null"                                       json:"clinic_id"`
	Name          string            `gorm:"not null"                                        json:"name"`
	Category      InventoryCategory `gorm:"type:inventory_category;not null"                json:"category"`
	Quantity      int               `gorm:"default:0"                                       json:"quantity"`
	Unit          string            `gorm:"not null;default:''"                             json:"unit"`
	MinStockLevel int               `gorm:"default:0"                                       json:"min_stock_level"`
	Location      string            `gorm:"default:''"                                      json:"location"`
	ExpiryDate    *time.Time        `gorm:"type:date"                                       json:"expiry_date,omitempty"`
	Supplier      string            `gorm:"default:''"                                      json:"supplier"`
	LastRestocked *time.Time        `gorm:"type:date"                                       json:"last_restocked,omitempty"`
	Status        InventoryStatus   `gorm:"type:inventory_status;default:'sufficient'"      json:"status"`
	CreatedAt     time.Time         `gorm:"autoCreateTime"                                  json:"created_at"`
	UpdatedAt     time.Time         `gorm:"autoUpdateTime"                                  json:"updated_at"`
}

func (InventoryItem) TableName() string { return "inventory_items" }

// ---- Staff ----

type StaffRole string

const (
	StaffRoleVeterinarian StaffRole = "veterinarian"
	StaffRoleNurse        StaffRole = "nurse"
	StaffRoleTrimmer      StaffRole = "trimmer"
	StaffRoleReception    StaffRole = "reception"
	StaffRoleManager      StaffRole = "manager"
)

type Staff struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name          string         `gorm:"not null"                                       json:"name"`
	IsActive      bool           `gorm:"default:true"                                   json:"is_active"`
	StaffRole     StaffRole      `gorm:"type:staff_role;not null"                       json:"staff_role"`
	JobTitleID    *uint64        `                                                      json:"job_title_id,omitempty"`
	LicenseNumber string         `gorm:"default:''"                                     json:"license_number"`
	SortOrder     int            `gorm:"default:0"                                      json:"sort_order"`
	DeletedAt     gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	JobTitle *JobTitle `gorm:"foreignKey:JobTitleID" json:"job_title,omitempty"`
}

func (Staff) TableName() string { return "staffs" }

type ShiftType string

const (
	ShiftTypeFull      ShiftType = "full"
	ShiftTypeMorning   ShiftType = "morning"
	ShiftTypeAfternoon ShiftType = "afternoon"
	ShiftTypeOff       ShiftType = "off"
	ShiftTypePaidLeave ShiftType = "paid_leave"
)

type ShiftEntry struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	StaffID   uint64    `gorm:"not null"                                       json:"staff_id"`
	Date      time.Time `gorm:"type:date;not null"                             json:"date"`
	ShiftType ShiftType `gorm:"type:shift_type;not null"                       json:"shift_type"`
	StartTime string    `gorm:"default:''"                                     json:"start_time"`
	EndTime   string    `gorm:"default:''"                                     json:"end_time"`
	Note      string    `gorm:"default:''"                                     json:"note"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Staff Staff `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
}

func (ShiftEntry) TableName() string { return "shift_entries" }

// ---- Cage ----

type CageType string

const (
	CageTypeICU     CageType = "icu"
	CageTypeDog     CageType = "dog"
	CageTypeCat     CageType = "cat"
	CageTypeGeneral CageType = "general"
)

type CageSize string

const (
	CageSizeSmall  CageSize = "small"
	CageSizeMedium CageSize = "medium"
	CageSizeLarge  CageSize = "large"
)

type Cage struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	CageType    CageType  `gorm:"type:cage_type;not null"                        json:"cage_type"`
	CageSize    CageSize  `gorm:"type:cage_size;not null"                        json:"cage_size"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Cage) TableName() string { return "cages" }

// ---- Master entities ----

type AnimalSpecies struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Name      string    `gorm:"not null"                                       json:"name"`
	IsActive  bool      `gorm:"default:true"                                   json:"is_active"`
	SortOrder int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (AnimalSpecies) TableName() string { return "animal_species" }

type VaccineSpecies string

const (
	VaccineSpeciesDog  VaccineSpecies = "dog"
	VaccineSpeciesCat  VaccineSpecies = "cat"
	VaccineSpeciesBoth VaccineSpecies = "both"
)

type Vaccine struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64          `gorm:"not null"                                       json:"clinic_id"`
	Name        string          `gorm:"not null"                                       json:"name"`
	Price       *float64        `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool            `gorm:"default:true"                                   json:"is_active"`
	Description string          `gorm:"default:''"                                     json:"description"`
	Species     *VaccineSpecies `gorm:"type:vaccine_species"                           json:"species,omitempty"`
	Interval    string          `gorm:"default:''"                                     json:"interval"`
	SortOrder   int             `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Vaccine) TableName() string { return "vaccines" }

type Insurance struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID     uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name         string    `gorm:"not null"                                       json:"name"`
	IsActive     bool      `gorm:"default:true"                                   json:"is_active"`
	Description  string    `gorm:"default:''"                                     json:"description"`
	CoverageRate *int      `gorm:"column:coverage_rate"                           json:"coverage_rate,omitempty"`
	ContactPhone string    `gorm:"default:''"                                     json:"contact_phone"`
	SortOrder    int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt    time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Insurance) TableName() string { return "insurances" }

type DosageForm string

const (
	DosageFormTablet    DosageForm = "tablet"
	DosageFormLiquid    DosageForm = "liquid"
	DosageFormInjection DosageForm = "injection"
	DosageFormTopical   DosageForm = "topical"
	DosageFormPowder    DosageForm = "powder"
)

type MedicineUnit string

const (
	MedicineUnitPerTablet MedicineUnit = "per_tablet"
	MedicineUnitPerML     MedicineUnit = "per_ml"
	MedicineUnitPerDose   MedicineUnit = "per_dose"
	MedicineUnitPerGram   MedicineUnit = "per_gram"
)

type Medicine struct {
	ID              uint64        `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID        uint64        `gorm:"not null"                                       json:"clinic_id"`
	Name            string        `gorm:"not null"                                       json:"name"`
	Price           *float64      `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive        bool          `gorm:"default:true"                                   json:"is_active"`
	Description     string        `gorm:"default:''"                                     json:"description"`
	DosageForm      *DosageForm   `gorm:"type:dosage_form"                               json:"dosage_form,omitempty"`
	MedicineUnit    *MedicineUnit `gorm:"type:medicine_unit"                             json:"medicine_unit,omitempty"`
	InventoryID     *uint64       `                                                      json:"inventory_id,omitempty"`
	DefaultQuantity int           `gorm:"default:1"                                      json:"default_quantity"`
	SortOrder       int           `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time     `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time     `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Inventory *InventoryItem `gorm:"foreignKey:InventoryID" json:"inventory,omitempty"`
}

func (Medicine) TableName() string { return "medicines" }

type ServiceType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	Color       string    `gorm:"default:'#3B82F6'"                              json:"color"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ServiceType) TableName() string { return "service_types" }

type Consultation struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID      uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name          string    `gorm:"not null"                                       json:"name"`
	Price         *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive      bool      `gorm:"default:true"                                   json:"is_active"`
	Description   string    `gorm:"default:''"                                     json:"description"`
	TimeCondition string    `gorm:"default:''"                                     json:"time_condition"`
	Duration      *int      `gorm:"type:integer"                                   json:"duration,omitempty"`
	SortOrder     int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt     time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Consultation) TableName() string { return "consultations" }

type AnesthesiaType string

const (
	AnesthesiaTypeNone     AnesthesiaType = "none"
	AnesthesiaTypeLocal    AnesthesiaType = "local"
	AnesthesiaTypeSedation AnesthesiaType = "sedation"
	AnesthesiaTypeGeneral  AnesthesiaType = "general"
)

type Procedure struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64         `gorm:"not null"                                       json:"clinic_id"`
	Name        string         `gorm:"not null"                                       json:"name"`
	Price       *float64       `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool           `gorm:"default:true"                                   json:"is_active"`
	Description string         `gorm:"default:''"                                     json:"description"`
	Duration    *int           `gorm:"type:integer"                                   json:"duration,omitempty"`
	Anesthesia  AnesthesiaType `gorm:"type:anesthesia_type;default:'none'"            json:"anesthesia"`
	SortOrder   int            `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Procedure) TableName() string { return "procedures" }

type BodySize string

const (
	BodySizeSmall  BodySize = "small"
	BodySizeMedium BodySize = "medium"
	BodySizeLarge  BodySize = "large"
)

type BillingUnit string

const (
	BillingUnitPerDay   BillingUnit = "per_day"
	BillingUnitPerNight BillingUnit = "per_night"
)

type HospitalizationPlan struct {
	ID          uint64       `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64       `gorm:"not null"                                       json:"clinic_id"`
	Name        string       `gorm:"not null"                                       json:"name"`
	Price       *float64     `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool         `gorm:"default:true"                                   json:"is_active"`
	Description string       `gorm:"default:''"                                     json:"description"`
	BodySize    *BodySize    `gorm:"type:body_size"                                 json:"body_size,omitempty"`
	BillingUnit *BillingUnit `gorm:"type:billing_unit"                              json:"billing_unit,omitempty"`
	SortOrder   int          `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (HospitalizationPlan) TableName() string { return "hospitalization_plans" }

type TargetSize string

const (
	TargetSizeSmall  TargetSize = "small"
	TargetSizeMedium TargetSize = "medium"
	TargetSizeLarge  TargetSize = "large"
	TargetSizeCat    TargetSize = "cat"
)

type TrimmingCourse struct {
	ID          uint64      `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64      `gorm:"not null"                                       json:"clinic_id"`
	Name        string      `gorm:"not null"                                       json:"name"`
	Price       *float64    `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool        `gorm:"default:true"                                   json:"is_active"`
	Description string      `gorm:"default:''"                                     json:"description"`
	TargetSize  *TargetSize `gorm:"type:target_size"                               json:"target_size,omitempty"`
	Duration    string      `gorm:"default:''"                                     json:"duration"`
	SortOrder   int         `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingCourse) TableName() string { return "trimming_courses" }

type TrimmingOption struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	Duration    string    `gorm:"default:''"                                     json:"duration"`
	Combinable  bool      `gorm:"default:true"                                   json:"combinable"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (TrimmingOption) TableName() string { return "trimming_options" }

type DiagnosisCategory struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Names []DiagnosisName `gorm:"foreignKey:DiagnosisCategoryID" json:"names,omitempty"`
}

func (DiagnosisCategory) TableName() string { return "diagnosis_categories" }

type DiagnosisName struct {
	ID                  uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID            uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name                string    `gorm:"not null"                                       json:"name"`
	IsActive            bool      `gorm:"default:true"                                   json:"is_active"`
	Description         string    `gorm:"default:''"                                     json:"description"`
	DiagnosisCategoryID uint64    `gorm:"not null"                                       json:"diagnosis_category_id"`
	SortOrder           int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt           time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Category *DiagnosisCategory `gorm:"foreignKey:DiagnosisCategoryID" json:"category,omitempty"`
}

func (DiagnosisName) TableName() string { return "diagnosis_names" }

type ExaminationType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// Relations
	Items []ExaminationTypeItem `gorm:"foreignKey:ExamTypeID" json:"items,omitempty"`
}

func (ExaminationType) TableName() string { return "exam_types" }

type ExaminationTypeItem struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ExamTypeID      uint64    `gorm:"not null"                                       json:"exam_type_id"`
	Name            string    `gorm:"not null"                                       json:"name"`
	InspectionValue string    `gorm:"default:''"                                     json:"inspection_value"`
	NormalValue     string    `gorm:"default:''"                                     json:"normal_value"`
	SortOrder       int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt       time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
}

func (ExaminationTypeItem) TableName() string { return "exam_type_items" }

type CheckupType struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID    uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name        string    `gorm:"not null"                                       json:"name"`
	Price       *float64  `gorm:"type:numeric(10,2)"                             json:"price,omitempty"`
	IsActive    bool      `gorm:"default:true"                                   json:"is_active"`
	Description string    `gorm:"default:''"                                     json:"description"`
	Interval    string    `gorm:"default:''"                                     json:"interval"`
	TargetAge   string    `gorm:"default:''"                                     json:"target_age"`
	SortOrder   int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (CheckupType) TableName() string { return "checkup_types" }

// ---- Clinic / Company / UserAccount ----

type UserType string

const (
	UserTypeSystemAdmin UserType = "system_admin"
	UserTypeClinicAdmin UserType = "clinic_admin"
	UserTypeStaff       UserType = "staff"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusLocked   AccountStatus = "locked"
)

type PermissionType string

const (
	PermissionAccountAdmin    PermissionType = "account_admin"
	PermissionMedical         PermissionType = "medical"
	PermissionMedicalRead     PermissionType = "medical_read"
	PermissionTrimming        PermissionType = "trimming"
	PermissionBilling         PermissionType = "billing"
	PermissionReception       PermissionType = "reception"
	PermissionHospitalization PermissionType = "hospitalization"
	PermissionMasterAdmin     PermissionType = "master_admin"
	PermissionShiftAdmin      PermissionType = "shift_admin"
	PermissionInventory       PermissionType = "inventory"
)

type Clinic struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	CompanyID          uint64    `gorm:"not null"                                       json:"company_id"`
	Name               string    `gorm:"not null"                                       json:"name"`
	PostalCode         string    `gorm:"default:''"                                     json:"postal_code"`
	Address            string    `gorm:"default:''"                                     json:"address"`
	PhoneNumber        string    `gorm:"default:''"                                     json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                                     json:"fax_number"`
	RegistrationNumber string    `gorm:"default:''"                                     json:"registration_number"`
	DirectorName       string    `gorm:"default:''"                                     json:"director_name"`
	Email              string    `gorm:"default:''"                                     json:"email"`
	Website            string    `gorm:"default:''"                                     json:"website"`
	LogoURL            string    `gorm:"default:''"                                     json:"logo_url"`
	IsActive           bool      `gorm:"default:true"                                   json:"is_active"`
	CreatedAt          time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (Clinic) TableName() string { return "clinics" }

type UserAccount struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	Email           string         `gorm:"not null;uniqueIndex"                           json:"email"`
	DisplayName     string         `gorm:"not null"                                       json:"display_name"`
	DisplayNameKana string         `gorm:"default:''"                                     json:"display_name_kana"`
	UserType        UserType       `gorm:"type:user_type;not null;default:'staff'"        json:"user_type"`
	JobTitleID      *uint64        `                                                      json:"job_title_id,omitempty"`
	Status          AccountStatus  `gorm:"type:account_status;default:'active'"           json:"status"`
	AvatarURL       string         `gorm:"default:''"                                     json:"avatar_url"`
	StaffID         *uint64        `                                                      json:"staff_id,omitempty"`
	PasswordHash    string         `gorm:"not null;default:''"                            json:"-"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`
	DeletedAt       gorm.DeletedAt `                                                      json:"-" swaggerignore:"true"`

	// Relations
	Staff    *Staff    `gorm:"foreignKey:StaffID"    json:"staff,omitempty"`
	JobTitle *JobTitle `gorm:"foreignKey:JobTitleID" json:"job_title,omitempty"`
}

func (UserAccount) TableName() string { return "user_accounts" }

type UserClinicMembership struct {
	ID       uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	UserID   uint64    `gorm:"not null"                                       json:"user_id"`
	ClinicID uint64    `gorm:"not null"                                       json:"clinic_id"`
	IsMain   bool      `gorm:"default:false"                                  json:"is_main"`
	JoinedAt time.Time `gorm:"default:now()"                                  json:"joined_at"`
}

func (UserClinicMembership) TableName() string { return "user_clinic_memberships" }

type UserPermission struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement"                       json:"id"`
	UserID     uint64         `gorm:"not null"                                       json:"user_id"`
	ClinicID   uint64         `gorm:"not null"                                       json:"clinic_id"`
	Permission PermissionType `gorm:"type:permission_type;not null"                  json:"permission"`
	GrantedBy  *uint64        `                                                      json:"granted_by,omitempty"`
	GrantedAt  time.Time      `gorm:"default:now()"                                  json:"granted_at"`
}

func (UserPermission) TableName() string { return "user_permissions" }

type Company struct {
	ID                 uint64    `gorm:"primaryKey;autoIncrement"  json:"id"`
	Name               string    `gorm:"not null"                  json:"name"`
	PostalCode         string    `gorm:"default:''"                json:"postal_code"`
	Address            string    `gorm:"default:''"                json:"address"`
	PhoneNumber        string    `gorm:"default:''"                json:"phone_number"`
	FaxNumber          string    `gorm:"default:''"                json:"fax_number"`
	Email              string    `gorm:"default:''"                json:"email"`
	Website            string    `gorm:"default:''"                json:"website"`
	DirectorName       string    `gorm:"default:''"                json:"director_name"`
	RegistrationNumber string    `gorm:"default:''"                json:"registration_number"`
	LogoURL            string    `gorm:"default:''"                json:"logo_url"`
	CreatedAt          time.Time `gorm:"autoCreateTime"            json:"created_at"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"            json:"updated_at"`
}

func (Company) TableName() string { return "company" }

type JobTitle struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name      string    `gorm:"not null"                                       json:"name"`
	SortOrder int       `gorm:"default:0"                                      json:"sort_order"`
	IsActive  bool      `gorm:"default:true"                                   json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (JobTitle) TableName() string { return "job_titles" }

type ChiefComplaintCategory struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID  uint64    `gorm:"not null"                                       json:"clinic_id"`
	Name      string    `gorm:"not null"                                       json:"name"`
	IsActive  bool      `gorm:"default:true"                                   json:"is_active"`
	SortOrder int       `gorm:"default:0"                                      json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`
}

func (ChiefComplaintCategory) TableName() string { return "chief_complaint_categories" }
