package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
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
	ID                uint64            `gorm:"primaryKey;autoIncrement"                       json:"id"`
	ClinicID          uint64            `gorm:"not null"                                       json:"clinic_id"`
	StartTime         time.Time         `gorm:"not null"                                       json:"start_time"`
	EndTime           time.Time         `gorm:"not null"                                       json:"end_time"`
	OwnerID           *uint64           `                                                      json:"owner_id,omitempty"`
	PetID             *uint64           `                                                      json:"pet_id,omitempty"`
	VisitType         VisitType         `gorm:"type:visit_type;not null;default:'revisit'"     json:"visit_type"`
	ReservationTypeID uint64            `gorm:"not null"                                       json:"reservation_type_id"`
	DoctorID          *uint64           `                                                      json:"doctor_id,omitempty"`
	IsDesignated      bool              `gorm:"default:false"                                  json:"is_designated"`
	Status            ReservationStatus `gorm:"type:reservation_status;default:'pending'"      json:"status"`
	Notes             string            `gorm:"default:''"                                     json:"notes"`
	// 受付ヘッダー テレメトリ（change-ui.md Phase 2）: checked_in へ遷移した時刻。待ち時間算出専用。
	CheckedInAt *time.Time     `gorm:"column:checked_in_at" json:"checked_in_at,omitempty"`
	DeletedAt   gorm.DeletedAt `                                                      json:"-"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"                                 json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"                                 json:"updated_at"`

	// FEAT-381-2 Commit 3
	ReservationRoute    *string    `gorm:"column:reservation_route;type:varchar(20)"  json:"reservation_route,omitempty"`
	ActualReservationAt *time.Time `gorm:"column:actual_reservation_at"               json:"actual_reservation_at,omitempty"`

	// LINE予約用フィールド
	Source           ReservationSource `gorm:"type:reservation_source;not null;default:'manual'" json:"source"`
	CreatedBy        *uint64           `gorm:""                                                  json:"created_by,omitempty"`
	LineCustomerID   *uint64           `                                                         json:"line_customer_id,omitempty"`
	IsStaffDelegated bool              `gorm:"not null;default:false"                            json:"is_staff_delegated"`
	CustomerFields   json.RawMessage   `gorm:"type:jsonb;not null;default:'{}'"                  json:"customer_fields"`

	// EMR-308: tentative appointment（次回予定）schema foundation
	// intended_date = 次回予定の対象日（日付のみ確定）; source_medical_record_id = 起点となったカルテの provenance
	IntendedDate          *time.Time `gorm:"column:intended_date;type:date"              json:"intended_date,omitempty"`
	SourceMedicalRecordID *uint64    `gorm:"column:source_medical_record_id"           json:"source_medical_record_id,omitempty"`

	// Relations
	Owner           *Owner                     `gorm:"foreignKey:OwnerID"              json:"owner,omitempty"`
	Pet             *Pet                       `gorm:"foreignKey:PetID"                json:"pet,omitempty"`
	ReservationType *ReservationType           `gorm:"foreignKey:ReservationTypeID"    json:"reservation_type,omitempty"`
	Doctor          *Staff                     `gorm:"foreignKey:DoctorID"             json:"doctor,omitempty"`
	CreatedByStaff  *Staff                     `gorm:"foreignKey:CreatedBy"            json:"created_by_staff,omitempty" tygo:"-"`
	LineCustomer    *LineCustomer              `gorm:"foreignKey:LineCustomerID"       json:"line_customer,omitempty"`
	TrimmingDetail  *AppointmentTrimmingDetail `gorm:"foreignKey:AppointmentID"        json:"trimming_detail,omitempty"`

	SourceMedicalRecord *MedicalRecord `gorm:"foreignKey:SourceMedicalRecordID"      json:"source_medical_record,omitempty"`
}

func (Reservation) TableName() string { return "appointments" }
