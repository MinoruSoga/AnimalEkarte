package reservation

import (
	"encoding/json"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// intendedDateString は tentative 行の intended_date を JST 暦日（YYYY-MM-DD）で返す。
func intendedDateString(d *time.Time) *string {
	if d == nil {
		return nil
	}
	s := d.In(config.JST).Format(time.DateOnly)
	return &s
}

type reservationResponse struct {
	ID                  uint64          `json:"id"`
	ClinicID            uint64          `json:"clinic_id"`
	StartTime           time.Time       `json:"start_time"`
	EndTime             time.Time       `json:"end_time"`
	OwnerID             *uint64         `json:"owner_id,omitempty"`
	PetID               *uint64         `json:"pet_id,omitempty"`
	VisitType           string          `json:"visit_type"`
	ReservationTypeID   uint64          `json:"reservation_type_id"`
	DoctorID            *uint64         `json:"doctor_id,omitempty"`
	IsDesignated        *bool           `json:"is_designated"`
	Status              string          `json:"status"`
	Notes               string          `json:"notes"`
	CheckedInAt         *time.Time      `json:"checked_in_at,omitempty"`
	Source              string          `json:"source"`
	CreatedBy           *uint64         `json:"created_by,omitempty"`
	LineCustomerID      *uint64         `json:"line_customer_id,omitempty"`
	IsStaffDelegated    bool            `json:"is_staff_delegated"`
	CustomerFields      json.RawMessage `json:"customer_fields"`
	ReservationRoute    *string         `json:"reservation_route,omitempty"`
	ActualReservationAt *time.Time      `json:"actual_reservation_at,omitempty"`
	// EMR-308: tentative 次回予定の暦日（YYYY-MM-DD）と登録元カルテ provenance。
	IntendedDate          *string `json:"intended_date,omitempty"`
	SourceMedicalRecordID *uint64 `json:"source_medical_record_id,omitempty"`
	// EMR-308 A3: intended_date が設定されていたかの明示フラグ。cancel/update 応答で
	// 常に true/false を返し、フロントエンドが「次回予定の再設定」を任意提示できる
	// ようにする（cancelled になっても intended_date は残り、自動復活はしない）。
	HadIntendedDate bool                     `json:"had_intended_date"`
	Owner           *ownerSummaryResponse    `json:"owner,omitempty"`
	Pet             *petSummaryResponse      `json:"pet,omitempty"`
	ReservationType *ReservationTypeResponse `json:"reservation_type,omitempty"`
	Doctor          *staffSummaryResponse    `json:"doctor,omitempty"`
	CreatedByStaff  *staffSummaryResponse    `json:"created_by_staff,omitempty"`
	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
}

func toReservationResponse(r *model.Reservation) reservationResponse {
	var reservationType *ReservationTypeResponse
	if r.ReservationType != nil {
		rt := ToReservationTypeResponse(r.ReservationType)
		reservationType = &rt
	}
	ownerID, owner := r.OwnerID, r.Owner
	// Tentative intent follows the current pet owner; real bookings keep history.
	if sharedkernel.PtrEquals(r.Status, model.ReservationStatusTentative) && r.Pet != nil && r.Pet.ClinicID == r.ClinicID {
		ownerID = &r.Pet.OwnerID
		owner = r.Pet.Owner
		if owner != nil && (owner.ID != r.Pet.OwnerID || owner.ClinicID != r.ClinicID) {
			owner = nil
		}
	}

	return reservationResponse{
		ID:                    r.ID,
		ClinicID:              r.ClinicID,
		StartTime:             httpapi.LocalTime(r.StartTime),
		EndTime:               httpapi.LocalTime(r.EndTime),
		OwnerID:               ownerID,
		PetID:                 r.PetID,
		VisitType:             string(r.VisitType),
		ReservationTypeID:     r.ReservationTypeID,
		DoctorID:              r.DoctorID,
		IsDesignated:          r.IsDesignated,
		Status:                sharedkernel.TextOrEmpty(r.Status),
		Notes:                 r.Notes,
		CheckedInAt:           httpapi.LocalTimePtr(r.CheckedInAt),
		Source:                string(r.Source),
		CreatedBy:             r.CreatedBy,
		LineCustomerID:        r.LineCustomerID,
		IsStaffDelegated:      r.IsStaffDelegated,
		CustomerFields:        r.CustomerFields,
		ReservationRoute:      r.ReservationRoute,
		ActualReservationAt:   httpapi.LocalTimePtr(r.ActualReservationAt),
		IntendedDate:          intendedDateString(r.IntendedDate),
		SourceMedicalRecordID: r.SourceMedicalRecordID,
		HadIntendedDate:       r.IntendedDate != nil,
		Owner:                 toOwnerSummary(owner),
		Pet:                   toPetSummary(r.Pet),
		ReservationType:       reservationType,
		Doctor:                toStaffSummary(r.Doctor),
		CreatedByStaff:        toStaffSummary(r.CreatedByStaff),
		CreatedAt:             httpapi.LocalTime(r.CreatedAt),
		UpdatedAt:             httpapi.LocalTime(r.UpdatedAt),
	}
}
