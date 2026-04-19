package handler

type createProcedureRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *int64   `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Duration    *int     `json:"duration"`
	Anesthesia  string   `json:"anesthesia"  binding:"required,oneof=none local sedation general"`
	ParentID    *uint64  `json:"parent_id"`
	SortOrder   int      `json:"sort_order"`
	TaxType     string   `json:"tax_type"    binding:"required,oneof=included excluded exempt"`
	TaxRate     *float64 `json:"tax_rate"`
}

type updateProcedureRequest struct {
	Name          *string  `json:"name"`
	Price         *int64   `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   *string  `json:"description"`
	Duration      *int     `json:"duration"`
	Anesthesia    *string  `json:"anesthesia"   binding:"omitempty,oneof=none local sedation general"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     *int     `json:"sort_order"`
	TaxType       *string  `json:"tax_type"     binding:"omitempty,oneof=included excluded exempt"`
	TaxRate       *float64 `json:"tax_rate"`
}
