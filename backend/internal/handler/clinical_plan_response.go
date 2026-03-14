package handler

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type clinicalPlanResponse struct {
	ID                  string                      `json:"id"`
	MedicalRecordID     string                      `json:"medical_record_id"`
	PhysicalExam        string                      `json:"physical_exam"`
	DiagnosisCategoryID *string                     `json:"diagnosis_category_id,omitempty"`
	DiagnosisNameID     *string                     `json:"diagnosis_name_id,omitempty"`
	DiagnosisDetails    string                      `json:"diagnosis_details"`
	TreatmentPolicy     string                      `json:"treatment_policy"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
	DiagnosisCategory   *diagnosisCategoryResponse  `json:"diagnosis_category,omitempty"`
	DiagnosisName       *diagnosisNameResponse      `json:"diagnosis_name,omitempty"`
}

func toClinicalPlanResponse(p *model.ClinicalPlan) clinicalPlanResponse {
	resp := clinicalPlanResponse{
		ID:               strconv.FormatUint(p.ID, 10),
		MedicalRecordID:  strconv.FormatUint(p.MedicalRecordID, 10),
		PhysicalExam:     p.PhysicalExam,
		DiagnosisDetails: p.DiagnosisDetails,
		TreatmentPolicy:  p.TreatmentPolicy,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
	if p.DiagnosisCategoryID != nil {
		s := strconv.FormatUint(*p.DiagnosisCategoryID, 10)
		resp.DiagnosisCategoryID = &s
	}
	if p.DiagnosisNameID != nil {
		s := strconv.FormatUint(*p.DiagnosisNameID, 10)
		resp.DiagnosisNameID = &s
	}
	if p.DiagnosisCategory != nil {
		r := toDiagnosisCategoryResponse(p.DiagnosisCategory)
		resp.DiagnosisCategory = &r
	}
	if p.DiagnosisName != nil {
		r := toDiagnosisNameResponse(p.DiagnosisName)
		resp.DiagnosisName = &r
	}
	return resp
}
