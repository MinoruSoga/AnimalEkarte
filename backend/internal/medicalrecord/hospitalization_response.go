package medicalrecord

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// HospitalizationResponse is the hospitalization list/detail/create/update HTTP wire DTO
// (domain-owned). TASK-444-S2: tygo source for frontend hospitalization-responses.ts.
// Nested treatment_plans / care_plan_items / daily_records are NOT on this wire — they are
// loaded via dedicated nested endpoints (e.g. GET /hospitalizations/:id/treatment-plans).
type HospitalizationResponse struct {
	ID                   uint64    `json:"id"`
	ClinicID             uint64    `json:"clinic_id"`
	OwnerID              uint64    `json:"owner_id"`
	PetID                uint64    `json:"pet_id"`
	HospitalizationType  string    `json:"hospitalization_type"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	Status               string    `json:"status"`
	CageID               *uint64   `json:"cage_id,omitempty"`
	DoctorID             *uint64   `json:"doctor_id,omitempty"`
	InsuranceCompanyName *string   `json:"insurance_company_name,omitempty"`
	InsuranceNumber      *string   `json:"insurance_number,omitempty"`
	Memo                 string    `json:"memo"`
	OwnerRequest         string    `json:"owner_request"`
	StaffNotes           string    `json:"staff_notes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	// リレーション: 一覧/詳細で飼主名/ペット名/種別/担当医を表示するため。Preload 時のみ埋まる。
	Owner  *OwnerSummaryResponse `json:"owner,omitempty"`
	Pet    *PetSummaryResponse   `json:"pet,omitempty"`
	Doctor *StaffSummaryResponse `json:"doctor,omitempty"`
}

func toHospitalizationResponse(h *model.Hospitalization) HospitalizationResponse {
	return HospitalizationResponse{
		ID:                   h.ID,
		ClinicID:             h.ClinicID,
		OwnerID:              h.OwnerID,
		PetID:                h.PetID,
		HospitalizationType:  string(h.HospitalizationType),
		StartDate:            httpapi.LocalTime(h.StartDate),
		EndDate:              httpapi.LocalTime(h.EndDate),
		Status:               string(h.Status),
		CageID:               h.CageID,
		DoctorID:             h.DoctorID,
		InsuranceCompanyName: h.InsuranceCompanyName,
		InsuranceNumber:      h.InsuranceNumber,
		Memo:                 h.Memo,
		OwnerRequest:         h.OwnerRequest,
		StaffNotes:           h.StaffNotes,
		CreatedAt:            httpapi.LocalTime(h.CreatedAt),
		UpdatedAt:            httpapi.LocalTime(h.UpdatedAt),
		Owner:                toOwnerSummary(h.Owner),
		Pet:                  toPetSummary(h.Pet),
		Doctor:               toStaffSummary(h.Doctor),
	}
}

type dischargeWithBillingResponse struct {
	HospitalizationID uint64  `json:"hospitalization_id"`
	AccountingID      *uint64 `json:"accounting_id,omitempty"`
	Status            string  `json:"status"`
}

func toDischargeWithBillingResponse(r *DischargeWithBillingResult) dischargeWithBillingResponse {
	return dischargeWithBillingResponse{
		HospitalizationID: r.HospitalizationID,
		AccountingID:      r.AccountingID,
		Status:            r.Status,
	}
}
