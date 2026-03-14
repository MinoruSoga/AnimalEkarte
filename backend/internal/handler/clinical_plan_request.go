package handler

type updateClinicalPlanRequest struct {
	PhysicalExam        *string `json:"physical_exam"`
	DiagnosisCategoryID *uint64 `json:"diagnosis_category_id"`
	DiagnosisNameID     *uint64 `json:"diagnosis_name_id"`
	DiagnosisDetails    *string `json:"diagnosis_details"`
	TreatmentPolicy     *string `json:"treatment_policy"`
}
