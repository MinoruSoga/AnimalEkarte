package handler

type createDiagnosisCategoryRequest struct {
	Name        string `json:"name"        binding:"required"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type updateDiagnosisCategoryRequest struct {
	Name        string `json:"name"`
	IsActive    *bool  `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type createDiagnosisNameRequest struct {
	Name                string `json:"name"                  binding:"required"`
	DiagnosisCategoryID uint64 `json:"diagnosis_category_id" binding:"required"`
	IsActive            bool   `json:"is_active"`
	Description         string `json:"description"`
	SortOrder           int    `json:"sort_order"`
}

type updateDiagnosisNameRequest struct {
	Name                string `json:"name"`
	DiagnosisCategoryID uint64 `json:"diagnosis_category_id"`
	IsActive            *bool  `json:"is_active"`
	Description         string `json:"description"`
	SortOrder           int    `json:"sort_order"`
}
