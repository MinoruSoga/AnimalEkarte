package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type clinicalPlanResponse struct {
	ID               string                 `json:"id"`
	MedicalRecordID  string                 `json:"medical_record_id"`
	PhysicalExam     string                 `json:"physical_exam"`
	DiagnosisTypeID  *string                `json:"diagnosis_type_id,omitempty"`
	DiagnosisNameID  *string                `json:"diagnosis_name_id,omitempty"`
	Diagnosis2TypeID *string                `json:"diagnosis_2_type_id,omitempty"`
	Diagnosis2NameID *string                `json:"diagnosis_2_name_id,omitempty"`
	DiagnosisDetails string                 `json:"diagnosis_details"`
	TreatmentPolicy  string                 `json:"treatment_policy"`
	Version          int                    `json:"version"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DiagnosisType    *DiagnosisTypeResponse `json:"diagnosis_type,omitempty"`
	DiagnosisName    *DiagnosisNameResponse `json:"diagnosis_name,omitempty"`
	Diagnosis2Type   *DiagnosisTypeResponse `json:"diagnosis_2_type,omitempty"`
	Diagnosis2Name   *DiagnosisNameResponse `json:"diagnosis_2_name,omitempty"`
}

func toClinicalPlanResponse(p *model.ClinicalPlan) clinicalPlanResponse {
	resp := clinicalPlanResponse{
		ID:               strconv.FormatUint(p.ID, 10),
		MedicalRecordID:  strconv.FormatUint(p.MedicalRecordID, 10),
		PhysicalExam:     p.PhysicalExam,
		DiagnosisDetails: p.DiagnosisDetails,
		TreatmentPolicy:  p.TreatmentPolicy,
		Version:          p.Version,
		CreatedAt:        httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(p.UpdatedAt),
	}
	if p.DiagnosisTypeID != nil {
		s := strconv.FormatUint(*p.DiagnosisTypeID, 10)
		resp.DiagnosisTypeID = &s
	}
	if p.DiagnosisNameID != nil {
		s := strconv.FormatUint(*p.DiagnosisNameID, 10)
		resp.DiagnosisNameID = &s
	}
	if p.Diagnosis2TypeID != nil {
		s := strconv.FormatUint(*p.Diagnosis2TypeID, 10)
		resp.Diagnosis2TypeID = &s
	}
	if p.Diagnosis2NameID != nil {
		s := strconv.FormatUint(*p.Diagnosis2NameID, 10)
		resp.Diagnosis2NameID = &s
	}
	if p.DiagnosisType != nil {
		r := ToDiagnosisTypeResponse(p.DiagnosisType)
		resp.DiagnosisType = &r
	}
	if p.DiagnosisName != nil {
		r := ToDiagnosisNameResponse(p.DiagnosisName)
		resp.DiagnosisName = &r
	}
	if p.Diagnosis2Type != nil {
		r := ToDiagnosisTypeResponse(p.Diagnosis2Type)
		resp.Diagnosis2Type = &r
	}
	if p.Diagnosis2Name != nil {
		r := ToDiagnosisNameResponse(p.Diagnosis2Name)
		resp.Diagnosis2Name = &r
	}
	return resp
}
