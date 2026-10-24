package medicalrecord

import (
	"strconv"
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
)

// clinicalPlanDiagnosisResponse は診断行（契約 §0.3 の closed shape + id / version）。
// マスタ参照とスナップショットは保存値のまま（現在のマスタから補完しない）。
type clinicalPlanDiagnosisResponse struct {
	ID                    string  `json:"id"`
	Version               int64   `json:"version"`
	DiagnosisSlot         int16   `json:"diagnosis_slot"`
	DiagnosisTypeID       *string `json:"diagnosis_type_id"`
	DiagnosisNameID       *string `json:"diagnosis_name_id"`
	DiagnosisNameSnapshot *string `json:"diagnosis_name_snapshot"`
	CategoryNameSnapshot  *string `json:"category_name_snapshot"`
	SortOrder             int     `json:"sort_order"`
}

type clinicalPlanResponse struct {
	ID               string                          `json:"id"`
	MedicalRecordID  string                          `json:"medical_record_id"`
	PhysicalExam     string                          `json:"physical_exam"`
	Diagnoses        []clinicalPlanDiagnosisResponse `json:"diagnoses"`
	DiagnosisDetails string                          `json:"diagnosis_details"`
	TreatmentPolicy  string                          `json:"treatment_policy"`
	Version          int                             `json:"version"`
	CreatedAt        time.Time                       `json:"created_at"`
	UpdatedAt        time.Time                       `json:"updated_at"`
}

func optionalIDString(id *uint64) *string {
	if id == nil {
		return nil
	}
	s := strconv.FormatUint(*id, 10)
	return &s
}

func toClinicalPlanResponse(v *ClinicalPlanView) clinicalPlanResponse {
	p := v.Plan
	diagnoses := make([]clinicalPlanDiagnosisResponse, 0, len(v.Diagnoses))
	for _, d := range v.Diagnoses {
		diagnoses = append(diagnoses, clinicalPlanDiagnosisResponse{
			ID:                    strconv.FormatUint(d.ID, 10),
			Version:               d.Version,
			DiagnosisSlot:         d.DiagnosisSlot,
			DiagnosisTypeID:       optionalIDString(d.DiagnosisTypeID),
			DiagnosisNameID:       optionalIDString(d.DiagnosisNameID),
			DiagnosisNameSnapshot: d.DiagnosisNameSnapshot,
			CategoryNameSnapshot:  d.CategoryNameSnapshot,
			SortOrder:             d.SortOrder,
		})
	}
	return clinicalPlanResponse{
		ID:               strconv.FormatUint(p.ID, 10),
		MedicalRecordID:  strconv.FormatUint(p.MedicalRecordID, 10),
		PhysicalExam:     p.PhysicalExam,
		Diagnoses:        diagnoses,
		DiagnosisDetails: p.DiagnosisDetails,
		TreatmentPolicy:  p.TreatmentPolicy,
		Version:          p.Version,
		CreatedAt:        httpapi.LocalTime(p.CreatedAt),
		UpdatedAt:        httpapi.LocalTime(p.UpdatedAt),
	}
}
