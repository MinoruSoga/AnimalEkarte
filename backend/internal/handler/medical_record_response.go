package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type medicalRecordResponse struct {
	ID            uint64                  `json:"id"`
	ClinicID      uint64                  `json:"clinic_id"`
	RecordNo      string                  `json:"record_no"`
	Date          time.Time               `json:"date"`
	OwnerID       *uint64                 `json:"owner_id,omitempty"`
	PetID         *uint64                 `json:"pet_id,omitempty"`
	DoctorID      *uint64                 `json:"doctor_id,omitempty"`
	AppointmentID *uint64                 `json:"appointment_id,omitempty"`
	Status        string                  `json:"status"`
	AccountingID  *uint64                 `json:"accounting_id,omitempty"`
	VisitCount    int64                   `json:"visit_count"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	Owner         *ownerSummaryResponse   `json:"owner,omitempty"`
	Pet           *petSummaryResponse     `json:"pet,omitempty"`
	Doctor        *staffSummaryResponse   `json:"doctor,omitempty"`
	Inquiry       *inquirySummaryResponse `json:"inquiry,omitempty"`
}

// inquirySummaryResponse はカルテ一覧で使用する問診の要約型
type inquirySummaryResponse struct {
	ID             uint64 `json:"id"`
	ChiefComplaint string `json:"chief_complaint"`
}

func toMedicalRecordResponse(r *model.MedicalRecord) medicalRecordResponse {
	return toMedicalRecordResponseWithVisitCount(r, 0)
}

func toMedicalRecordResponseWithVisitCount(r *model.MedicalRecord, visitCount int64) medicalRecordResponse {
	resp := medicalRecordResponse{
		ID:            r.ID,
		ClinicID:      r.ClinicID,
		RecordNo:      r.RecordNo,
		Date:          r.Date,
		OwnerID:       r.OwnerID,
		PetID:         r.PetID,
		DoctorID:      r.DoctorID,
		AppointmentID: r.AppointmentID,
		Status:        string(r.Status),
		VisitCount:    visitCount,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		Owner:         toOwnerSummary(r.Owner),
		Pet:           toPetSummary(r.Pet),
		Doctor:        toStaffSummary(r.Doctor),
	}
	if r.Billing != nil {
		resp.AccountingID = &r.Billing.ID
	}
	if r.Inquiry != nil {
		resp.Inquiry = &inquirySummaryResponse{
			ID:             r.Inquiry.ID,
			ChiefComplaint: r.Inquiry.ChiefComplaint,
		}
	}
	return resp
}
