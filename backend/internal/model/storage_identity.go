package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// Current migration is authoritative. These are persistence values, not public DTOs or workflow implementations.

// CommonOwnerBasicRevision maps the current common_owner_basic_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonOwnerBasicRevision struct {
	ID                 uint64           `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	CommonOwnerID      uuid.UUID        `gorm:"column:common_owner_id;type:uuid;not null" json:"-"`
	Revision           int64            `gorm:"column:revision;type:bigint;not null" json:"-"`
	SourceClinicID     uint64           `gorm:"column:source_clinic_id;type:bigint;not null" json:"-"`
	SourceLocalOwnerID uint64           `gorm:"column:source_local_owner_id;type:bigint;not null" json:"-"`
	OperationID        uint64           `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	Name               string           `gorm:"column:name;type:text;not null" json:"-"`
	NameKana           *string          `gorm:"column:name_kana;type:text" json:"-"`
	BirthDate          *time.Time       `gorm:"column:birth_date;type:date" json:"-"`
	Company            *string          `gorm:"column:company;type:text" json:"-"`
	PostalCode         *string          `gorm:"column:postal_code;type:text" json:"-"`
	Address1           *string          `gorm:"column:address1;type:text" json:"-"`
	Address2           *string          `gorm:"column:address2;type:text" json:"-"`
	HomePostalCode     *string          `gorm:"column:home_postal_code;type:text" json:"-"`
	HomeAddress1       *string          `gorm:"column:home_address1;type:text" json:"-"`
	HomeAddress2       *string          `gorm:"column:home_address2;type:text" json:"-"`
	Phone              *string          `gorm:"column:phone;type:text" json:"-"`
	CompanyPhone       *string          `gorm:"column:company_phone;type:text" json:"-"`
	Email              *string          `gorm:"column:email;type:text" json:"-"`
	CygramResult       *json.RawMessage `gorm:"column:cygram_result;type:jsonb" json:"-"`
	CygramRuleVersion  *string          `gorm:"column:cygram_rule_version;type:text" json:"-"`
	Provenance         json.RawMessage  `gorm:"column:provenance;type:jsonb;not null" json:"-"`
	RecordedBy         uint64           `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt         time.Time        `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CommonOwnerBasicRevision) TableName() string { return "common_owner_basic_revisions" }

// CommonOwnerClinicBinding maps the current common_owner_clinic_bindings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonOwnerClinicBinding struct {
	ID            uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID      uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	CommonOwnerID uuid.UUID `gorm:"column:common_owner_id;type:uuid;not null" json:"-"`
	LocalOwnerID  uint64    `gorm:"column:local_owner_id;type:bigint;not null" json:"-"`
	ConfirmedBy   uint64    `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt   time.Time `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
	OperationID   uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (CommonOwnerClinicBinding) TableName() string { return "common_owner_clinic_bindings" }

// CommonOwnerSubject maps the current common_owner_subjects storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonOwnerSubject struct {
	ID                     uuid.UUID `gorm:"column:id;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	Version                int64     `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CurrentBasicRevisionID *uint64   `gorm:"column:current_basic_revision_id;type:bigint" json:"-"`
	CreatedAt              time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CommonOwnerSubject) TableName() string { return "common_owner_subjects" }

// CommonPetBasicRevision maps the current common_pet_basic_revisions storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonPetBasicRevision struct {
	ID                     uint64            `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	CommonPetID            uuid.UUID         `gorm:"column:common_pet_id;type:uuid;not null" json:"-"`
	Revision               int64             `gorm:"column:revision;type:bigint;not null" json:"-"`
	SourceClinicID         uint64            `gorm:"column:source_clinic_id;type:bigint;not null" json:"-"`
	SourceLocalPetID       uint64            `gorm:"column:source_local_pet_id;type:bigint;not null" json:"-"`
	OperationID            uint64            `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	Name                   string            `gorm:"column:name;type:text;not null" json:"-"`
	NameKana               *string           `gorm:"column:name_kana;type:text" json:"-"`
	AnimalSpeciesID        uint64            `gorm:"column:animal_species_id;type:bigint;not null" json:"-"`
	Breed                  *string           `gorm:"column:breed;type:text" json:"-"`
	Gender                 *PetGender        `gorm:"column:gender;type:pet_gender" json:"-"`
	Status                 *PetStatus        `gorm:"column:status;type:pet_status" json:"-"`
	BirthDate              *time.Time        `gorm:"column:birth_date;type:date" json:"-"`
	Color                  *string           `gorm:"column:color;type:text" json:"-"`
	RegistrationWeight     *ExactNumeric     `gorm:"column:registration_weight;type:numeric" json:"-"`
	RegistrationWeightUnit *string           `gorm:"column:registration_weight_unit;type:text" json:"-"`
	NeuteredStatus         PetNeuteredStatus `gorm:"column:neutered_status;type:pet_neutered_status;not null;default:'unknown'" json:"-"`
	NeuteredDate           *time.Time        `gorm:"column:neutered_date;type:date" json:"-"`
	AcquisitionType        *AcquisitionType  `gorm:"column:acquisition_type;type:acquisition_type" json:"-"`
	DangerLevel            *DangerLevel      `gorm:"column:danger_level;type:danger_level" json:"-"`
	DangerReason           *string           `gorm:"column:danger_reason;type:text" json:"-"`
	Food                   *string           `gorm:"column:food;type:text" json:"-"`
	Environment            *string           `gorm:"column:environment;type:text" json:"-"`
	Phone                  *string           `gorm:"column:phone;type:text" json:"-"`
	Remarks                *string           `gorm:"column:remarks;type:text" json:"-"`
	DeceasedAt             *time.Time        `gorm:"column:deceased_at;type:timestamptz" json:"-"`
	DeceasedReason         *string           `gorm:"column:deceased_reason;type:text" json:"-"`
	BloodType              *string           `gorm:"column:blood_type;type:text" json:"-"`
	MicrochipNumber        *string           `gorm:"column:microchip_number;type:text" json:"-"`
	NameOrigin             *string           `gorm:"column:name_origin;type:text" json:"-"`
	MeetingStory           *string           `gorm:"column:meeting_story;type:text" json:"-"`
	Provenance             json.RawMessage   `gorm:"column:provenance;type:jsonb;not null" json:"-"`
	RecordedBy             uint64            `gorm:"column:recorded_by;type:bigint;not null" json:"-"`
	RecordedAt             time.Time         `gorm:"column:recorded_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CommonPetBasicRevision) TableName() string { return "common_pet_basic_revisions" }

// CommonPetClinicBinding maps the current common_pet_clinic_bindings storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonPetClinicBinding struct {
	ID          uint64    `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID    uint64    `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	CommonPetID uuid.UUID `gorm:"column:common_pet_id;type:uuid;not null" json:"-"`
	LocalPetID  uint64    `gorm:"column:local_pet_id;type:bigint;not null" json:"-"`
	ConfirmedBy uint64    `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt time.Time `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
	OperationID uint64    `gorm:"column:operation_id;type:bigint;not null" json:"-"`
}

func (CommonPetClinicBinding) TableName() string { return "common_pet_clinic_bindings" }

// CommonPetSubject maps the current common_pet_subjects storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type CommonPetSubject struct {
	ID                     uuid.UUID `gorm:"column:id;type:uuid;primaryKey;autoIncrement:false;not null" json:"-"`
	Version                int64     `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	CurrentBasicRevisionID *uint64   `gorm:"column:current_basic_revision_id;type:bigint" json:"-"`
	CreatedAt              time.Time `gorm:"column:created_at;type:timestamptz;not null;default:clock_timestamp();autoCreateTime:false;autoUpdateTime:false" json:"-"`
}

func (CommonPetSubject) TableName() string { return "common_pet_subjects" }

// OwnerIdentityConfirmation maps the current owner_identity_confirmations storage contract. Required write-owner/authorization/receipt rules remain in the owning workflow.
type OwnerIdentityConfirmation struct {
	ID               uint64          `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"-"`
	ClinicID         uint64          `gorm:"column:clinic_id;type:bigint;not null" json:"-"`
	OwnerID          uint64          `gorm:"column:owner_id;type:bigint;not null" json:"-"`
	CandidateOwnerID uint64          `gorm:"column:candidate_owner_id;type:bigint;not null" json:"-"`
	Determination    string          `gorm:"column:determination;type:text;not null" json:"-"`
	Evidence         json.RawMessage `gorm:"column:evidence;type:jsonb;not null" json:"-"`
	OperationID      uint64          `gorm:"column:operation_id;type:bigint;not null" json:"-"`
	ConfirmedBy      uint64          `gorm:"column:confirmed_by;type:bigint;not null" json:"-"`
	ConfirmedAt      time.Time       `gorm:"column:confirmed_at;type:timestamptz;not null" json:"-"`
}

func (OwnerIdentityConfirmation) TableName() string { return "owner_identity_confirmations" }
