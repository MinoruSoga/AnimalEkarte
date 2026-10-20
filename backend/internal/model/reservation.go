package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type ReservationStatus string

const (
	ReservationStatusConfirmed      ReservationStatus = "confirmed"
	ReservationStatusPending        ReservationStatus = "pending"
	ReservationStatusCancelled      ReservationStatus = "cancelled"
	ReservationStatusCheckedIn      ReservationStatus = "checked_in"
	ReservationStatusInConsultation ReservationStatus = "in_consultation"
	ReservationStatusAccounting     ReservationStatus = "accounting"
	ReservationStatusCompleted      ReservationStatus = "completed"
	ReservationStatusNoShow         ReservationStatus = "no_show"
)

// EMR-308: tentative（次回予定）= date known, time/slot/doctor undecided。hierarchy: tentative < pending < confirmed。
// 故意に const ブロック外の var で宣言する: tygo enum_style:"union" は const 列挙を
// ReservationStatus union として生成するため、ここに置くと FE の exhaustive union に
// "tentative" が混入し既存の予約ステータス分岐（全 status を網羅する前提の switch/map）を
// 型レベルで壊す。tentative 行は FE へ isTentative / intended_date で届ける別経路。
var ReservationStatusTentative ReservationStatus = "tentative"

type VisitType string

const (
	VisitTypeFirst   VisitType = "first"
	VisitTypeRevisit VisitType = "revisit"
)

// ReservationSource indicates whether a reservation was created via LINE or manually.
type ReservationSource string

const (
	ReservationSourceManual ReservationSource = "manual"
	ReservationSourceLine   ReservationSource = "line"
)

type Reservation struct {
	ID                      uint64                     `gorm:"column:id;type:bigint;primaryKey;autoIncrement:true;not null" json:"id"`
	ClinicID                uint64                     `gorm:"column:clinic_id;type:bigint;not null" json:"clinic_id"`
	StartTime               time.Time                  `gorm:"column:start_time;type:timestamptz;not null" json:"start_time"`
	EndTime                 time.Time                  `gorm:"column:end_time;type:timestamptz;not null" json:"end_time"`
	OwnerID                 *uint64                    `gorm:"column:owner_id;type:bigint" json:"owner_id,omitempty"`
	PetID                   *uint64                    `gorm:"column:pet_id;type:bigint" json:"pet_id,omitempty"`
	VisitType               VisitType                  `gorm:"column:visit_type;type:visit_type;not null;default:'revisit'" json:"visit_type"`
	ReservationTypeID       uint64                     `gorm:"column:reservation_type_id;type:bigint;not null" json:"reservation_type_id"`
	DoctorID                *uint64                    `gorm:"column:doctor_id;type:bigint" json:"doctor_id,omitempty"`
	IsDesignated            *bool                      `gorm:"column:is_designated;type:boolean;default:false" json:"is_designated"`
	Status                  *ReservationStatus         `gorm:"column:status;type:reservation_status;default:'pending'" json:"status"`
	Notes                   string                     `gorm:"column:notes;type:text;not null;default:''" json:"notes"`
	Source                  ReservationSource          `gorm:"column:source;type:reservation_source;not null;default:'manual'" json:"source"`
	CreatedBy               *uint64                    `gorm:"column:created_by;type:bigint" json:"created_by,omitempty"`
	IsStaffDelegated        bool                       `gorm:"column:is_staff_delegated;type:boolean;not null;default:false" json:"is_staff_delegated"`
	CustomerFields          json.RawMessage            `gorm:"column:customer_fields;type:jsonb;not null;default:'{}'" json:"customer_fields"`
	ReservationRoute        *string                    `gorm:"column:reservation_route;type:varchar(20)" json:"reservation_route,omitempty"`
	ActualReservationAt     *time.Time                 `gorm:"column:actual_reservation_at;type:timestamptz" json:"actual_reservation_at,omitempty"`
	CreatedAt               time.Time                  `gorm:"column:created_at;type:timestamptz;not null;default:now();autoCreateTime" json:"created_at"`
	UpdatedAt               time.Time                  `gorm:"column:updated_at;type:timestamptz;not null;default:now();autoUpdateTime" json:"updated_at"`
	DeletedAt               gorm.DeletedAt             `gorm:"column:deleted_at;type:timestamptz" json:"-"`
	LineCustomerID          *uint64                    `gorm:"column:line_customer_id;type:bigint" json:"line_customer_id,omitempty"`
	CheckedInAt             *time.Time                 `gorm:"column:checked_in_at;type:timestamptz" json:"checked_in_at,omitempty"`
	IntendedDate            *time.Time                 `gorm:"column:intended_date;type:date" json:"intended_date,omitempty"`
	SourceMedicalRecordID   *uint64                    `gorm:"column:source_medical_record_id;type:bigint" json:"source_medical_record_id,omitempty"`
	RevisitIntakeGeneration int64                      `gorm:"column:revisit_intake_generation;type:bigint;not null;default:1" json:"-"`
	Version                 int64                      `gorm:"column:version;type:bigint;not null;default:1" json:"-"`
	ReservationNameSnapshot *string                    `gorm:"column:reservation_name_snapshot;type:text" json:"-"`
	ReservationNameRevision *int64                     `gorm:"column:reservation_name_revision;type:bigint" json:"-"`
	ReservationNameSource   *string                    `gorm:"column:reservation_name_source;type:text" json:"-"`
	StayCaseID              *uint64                    `gorm:"column:stay_case_id;type:bigint" json:"-"`
	Owner                   *Owner                     `gorm:"foreignKey:ClinicID,OwnerID;references:ClinicID,ID"              json:"owner,omitempty"`
	Pet                     *Pet                       `gorm:"foreignKey:ClinicID,PetID;references:ClinicID,ID"                json:"pet,omitempty"`
	ReservationType         *ReservationType           `gorm:"foreignKey:ReservationTypeID,ClinicID;references:ID,ClinicID"    json:"reservation_type,omitempty"`
	Doctor                  *Staff                     `gorm:"foreignKey:DoctorID;references:ID"             json:"doctor,omitempty"`
	CreatedByStaff          *Staff                     `gorm:"foreignKey:CreatedBy;references:ID"            json:"created_by_staff,omitempty" tygo:"-"`
	LineCustomer            *LineCustomer              `gorm:"foreignKey:LineCustomerID,ClinicID;references:ID,ClinicID"       json:"line_customer,omitempty"`
	TrimmingDetail          *AppointmentTrimmingDetail `gorm:"foreignKey:AppointmentID,ClinicID;references:ID,ClinicID"        json:"trimming_detail,omitempty"`
	SourceMedicalRecord     *MedicalRecord             `gorm:"foreignKey:SourceMedicalRecordID,ClinicID;references:ID,ClinicID"      json:"source_medical_record,omitempty"`
}

func (Reservation) TableName() string { return "appointments" }
