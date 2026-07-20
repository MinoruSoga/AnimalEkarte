package medicalrecord

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
	TaxRate     *float64 `json:"tax_rate" binding:"omitempty,min=0,max=1"`
}

func (r *createProcedureRequest) toServiceInput() *CreateProcedureInput {
	return &CreateProcedureInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		Duration:    r.Duration,
		Anesthesia:  r.Anesthesia,
		ParentID:    r.ParentID,
		SortOrder:   r.SortOrder,
		TaxType:     r.TaxType,
		TaxRate:     r.TaxRate,
	}
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
	TaxRate       *float64 `json:"tax_rate" binding:"omitempty,min=0,max=1"`
}

func (r *updateProcedureRequest) toServiceInput() *UpdateProcedureInput {
	return &UpdateProcedureInput{
		Name:          r.Name,
		Price:         r.Price,
		IsActive:      r.IsActive,
		Description:   r.Description,
		Duration:      r.Duration,
		Anesthesia:    r.Anesthesia,
		ParentID:      r.ParentID,
		ClearParentID: r.ClearParentID,
		SortOrder:     r.SortOrder,
		TaxType:       r.TaxType,
		TaxRate:       r.TaxRate,
	}
}
