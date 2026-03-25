package handler

// createHospitalizationPlanRequest は入院プラン作成リクエスト。
type createHospitalizationPlanRequest struct {
	Name        string   `json:"name"         binding:"required"`
	Price       *int64   `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
	TaxType     string   `json:"tax_type"`
	TaxRate     *float64 `json:"tax_rate"`
}

// updateHospitalizationPlanRequest は入院プラン更新リクエスト。
type updateHospitalizationPlanRequest struct {
	Name        string   `json:"name"`
	Price       *int64   `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
	TaxType     string   `json:"tax_type"`
	TaxRate     *float64 `json:"tax_rate"`
}

// reorderHospitalizationPlanRequest は入院プラン並び替えリクエスト。
type reorderHospitalizationPlanRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
