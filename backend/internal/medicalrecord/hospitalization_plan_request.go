package medicalrecord

// createHospitalizationPlanRequest は入院プラン作成リクエスト。
type createHospitalizationPlanRequest struct {
	Name        string   `json:"name"         binding:"required"`
	Price       *int64   `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"    binding:"omitempty,oneof=small medium large"`
	BillingUnit string   `json:"billing_unit" binding:"omitempty,oneof=per_day per_night"`
	SortOrder   int      `json:"sort_order"`
	TaxType     string   `json:"tax_type"     binding:"omitempty,oneof=included excluded exempt"`
	TaxRate     *float64 `json:"tax_rate"`
}

func (r *createHospitalizationPlanRequest) toServiceInput() *CreateHospitalizationPlanInput {
	return &CreateHospitalizationPlanInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		TaxType:     r.TaxType,
		TaxRate:     r.TaxRate,
		BodySize:    r.BodySize,
		BillingUnit: r.BillingUnit,
	}
}

// updateHospitalizationPlanRequest は入院プラン更新リクエスト。
type updateHospitalizationPlanRequest struct {
	Name        *string  `json:"name"`
	Price       *int64   `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description *string  `json:"description"`
	BodySize    *string  `json:"body_size"`
	BillingUnit *string  `json:"billing_unit"`
	SortOrder   *int     `json:"sort_order"`
	TaxType     *string  `json:"tax_type"`
	TaxRate     *float64 `json:"tax_rate"`
}

func (r updateHospitalizationPlanRequest) toServiceInput() *UpdateHospitalizationPlanInput {
	return &UpdateHospitalizationPlanInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		BodySize:    r.BodySize,
		BillingUnit: r.BillingUnit,
		SortOrder:   r.SortOrder,
		TaxType:     r.TaxType,
		TaxRate:     r.TaxRate,
	}
}
