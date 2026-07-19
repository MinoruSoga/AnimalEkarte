package handler

import "github.com/animal-ekarte/backend/internal/service"

// updateClinicalPlanRequest は診療計画 PATCH のバインド struct。
// Diagnosis2TypeID / Diagnosis2NameID は nullableUint64RequestField:
// JSON未送信=未更新, null=NULLクリア, 数値=セット。
type updateClinicalPlanRequest struct {
	PhysicalExam     *string                    `json:"physical_exam"`
	DiagnosisTypeID  *uint64                    `json:"diagnosis_type_id"`
	DiagnosisNameID  *uint64                    `json:"diagnosis_name_id"`
	DiagnosisDetails *string                    `json:"diagnosis_details"`
	TreatmentPolicy  *string                    `json:"treatment_policy"`
	Diagnosis2TypeID nullableUint64RequestField `json:"diagnosis_2_type_id"`
	Diagnosis2NameID nullableUint64RequestField `json:"diagnosis_2_name_id"`
	Version          *int                       `json:"version"` // 楽観的ロック用
}

func (r updateClinicalPlanRequest) toServiceInput() *service.UpdateClinicalPlanInput {
	return &service.UpdateClinicalPlanInput{
		PhysicalExam:     r.PhysicalExam,
		DiagnosisTypeID:  r.DiagnosisTypeID,
		DiagnosisNameID:  r.DiagnosisNameID,
		DiagnosisDetails: r.DiagnosisDetails,
		TreatmentPolicy:  r.TreatmentPolicy,
		Diagnosis2TypeID: r.Diagnosis2TypeID.toServiceInput(),
		Diagnosis2NameID: r.Diagnosis2NameID.toServiceInput(),
		Version:          r.Version,
	}
}
