package medicalrecord

type createConsultationRequest struct {
	Name          string   `json:"name"           binding:"required,min=1,max=255"`
	Price         *int64   `json:"price"`
	IsActive      bool     `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"  binding:"omitempty,oneof=anytime morning afternoon evening"`
	Duration      *int     `json:"duration"`
	ParentID      *uint64  `json:"parent_id"`
	SortOrder     int      `json:"sort_order"`
	TaxType       string   `json:"tax_type"        binding:"omitempty,oneof=included excluded exempt"`
	TaxRate       *float64 `json:"tax_rate"        binding:"omitempty,min=0,max=1"`
}

func (r *createConsultationRequest) toServiceInput() *CreateConsultationInput {
	return &CreateConsultationInput{
		Name:          r.Name,
		Price:         r.Price,
		IsActive:      r.IsActive,
		Description:   r.Description,
		TimeCondition: r.TimeCondition,
		Duration:      r.Duration,
		ParentID:      r.ParentID,
		SortOrder:     r.SortOrder,
		TaxType:       nilIfEmpty(r.TaxType),
		TaxRate:       r.TaxRate,
	}
}

type updateConsultationRequest struct {
	Name          *string  `json:"name"`
	Price         *int64   `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   *string  `json:"description"`
	TimeCondition *string  `json:"time_condition"  binding:"omitempty,oneof=anytime morning afternoon evening"`
	Duration      *int     `json:"duration"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     *int     `json:"sort_order"`
	TaxType       *string  `json:"tax_type"        binding:"omitempty,oneof=included excluded exempt"`
	TaxRate       *float64 `json:"tax_rate"        binding:"omitempty,min=0,max=1"`
}

func (r *updateConsultationRequest) toServiceInput() *UpdateConsultationInput {
	return &UpdateConsultationInput{
		Name:          r.Name,
		Price:         r.Price,
		IsActive:      r.IsActive,
		Description:   r.Description,
		TimeCondition: r.TimeCondition,
		Duration:      r.Duration,
		ParentID:      r.ParentID,
		ClearParentID: r.ClearParentID,
		SortOrder:     r.SortOrder,
		TaxType:       r.TaxType,
		TaxRate:       r.TaxRate,
	}
}
