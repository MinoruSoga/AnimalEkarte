package handler

type createDiagnosisTypeRequest struct {
	Name        string `json:"name"        binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type updateDiagnosisTypeRequest struct {
	Name        *string `json:"name"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

type createDiagnosisNameRequest struct {
	Name            string `json:"name"                  binding:"required"`
	DiagnosisTypeID uint64 `json:"diagnosis_type_id" binding:"required"`
	IsActive        bool   `json:"is_active"`
	Description     string `json:"description"`
	SortOrder       int    `json:"sort_order"`
}

type updateDiagnosisNameRequest struct {
	Name            *string `json:"name"`
	DiagnosisTypeID *uint64 `json:"diagnosis_type_id"`
	IsActive        *bool   `json:"is_active"`
	Description     *string `json:"description"`
	SortOrder       *int    `json:"sort_order"`
}
