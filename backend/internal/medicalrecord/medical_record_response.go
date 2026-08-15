package medicalrecord

import (
	"github.com/animal-ekarte/backend/internal/httpapi"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordResponse is the medical-record list/detail/create/update HTTP wire DTO
// (domain-owned). TASK-444-S2: tygo source for frontend medicalrecord-responses.ts.
// Version is on the wire so FE OCC can read the current CAS token (not models.ts).
type MedicalRecordResponse struct {
	ID                       uint64                  `json:"id"`
	ClinicID                 uint64                  `json:"clinic_id"`
	RecordNo                 string                  `json:"record_no"`
	Date                     time.Time               `json:"date"`
	OwnerID                  *uint64                 `json:"owner_id,omitempty"`
	PetID                    *uint64                 `json:"pet_id,omitempty"`
	DoctorID                 *uint64                 `json:"doctor_id,omitempty"`
	AppointmentID            *uint64                 `json:"appointment_id,omitempty"`
	Status                   string                  `json:"status"`
	Version                  int                     `json:"version"`
	AccountingID             *uint64                 `json:"accounting_id,omitempty"`
	VisitCount               int64                   `json:"visit_count"`
	NextVisitRecommendedDate *string                 `json:"next_visit_recommended_date,omitempty"`
	RecommendationReason     *string                 `json:"recommendation_reason,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
	Owner                    *OwnerSummaryResponse   `json:"owner,omitempty"`
	Pet                      *PetSummaryResponse     `json:"pet,omitempty"`
	Doctor                   *StaffSummaryResponse   `json:"doctor,omitempty"`
	Inquiry                  *InquirySummaryResponse `json:"inquiry,omitempty"`
}

// InquirySummaryResponse is the inquiry embed on medical-record list/detail wire.
// Notes carries 問診タブ「治療方針」(inquiry.notes). Omitting it forced FE reload to DEFAULT
// "# 治療方針" after save/finalize (BUG-034) even though notes persisted in DB.
type InquirySummaryResponse struct {
	ID             uint64 `json:"id"`
	ChiefComplaint string `json:"chief_complaint"`
	Notes          string `json:"notes"`
}

func toMedicalRecordResponseWithVisitCount(r *model.MedicalRecord, visitCount int64) MedicalRecordResponse {
	resp := MedicalRecordResponse{
		ID:            r.ID,
		ClinicID:      r.ClinicID,
		RecordNo:      r.RecordNo,
		Date:          httpapi.LocalTime(r.Date),
		OwnerID:       r.OwnerID,
		PetID:         r.PetID,
		DoctorID:      r.DoctorID,
		AppointmentID: r.AppointmentID,
		Status:        string(r.Status),
		Version:       r.Version,
		VisitCount:    visitCount,
		CreatedAt:     httpapi.LocalTime(r.CreatedAt),
		UpdatedAt:     httpapi.LocalTime(r.UpdatedAt),
		Owner:         toOwnerSummary(r.Owner),
		Pet:           toPetSummary(r.Pet),
		Doctor:        toStaffSummary(r.Doctor),
	}
	if r.NextVisitRecommendedDate != nil {
		s := r.NextVisitRecommendedDate.In(time.Local).Format(time.DateOnly)
		resp.NextVisitRecommendedDate = &s
	}
	resp.RecommendationReason = r.RecommendationReason
	if r.Billing != nil {
		resp.AccountingID = &r.Billing.ID
	}
	if r.Inquiry != nil {
		resp.Inquiry = &InquirySummaryResponse{
			ID:             r.Inquiry.ID,
			ChiefComplaint: r.Inquiry.ChiefComplaint,
			Notes:          r.Inquiry.Notes,
		}
	}
	return resp
}

func toMedicalRecordResponse(r *model.MedicalRecord) MedicalRecordResponse {
	return toMedicalRecordResponseWithVisitCount(r, 0)
}
