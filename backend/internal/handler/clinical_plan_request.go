package handler

type updateClinicalPlanRequest struct {
	PhysicalExam         *string `json:"physical_exam"`
	DiagnosisTypeID      *uint64 `json:"diagnosis_type_id"`
	DiagnosisNameID      *uint64 `json:"diagnosis_name_id"`
	DiagnosisDetails     *string `json:"diagnosis_details"`
	TreatmentPolicy      *string `json:"treatment_policy"`
	Diagnosis2CategoryID *uint64 `json:"diagnosis_2_category_id"`
	Diagnosis2NameID     *uint64 `json:"diagnosis_2_name_id"`
}
