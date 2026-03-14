package handler

// createHospitalizationPlanRequest は入院プラン作成リクエスト。
type createHospitalizationPlanRequest struct {
	Name        string   `json:"name"         binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
}

// updateHospitalizationPlanRequest は入院プラン更新リクエスト。
type updateHospitalizationPlanRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	BodySize    string   `json:"body_size"`
	BillingUnit string   `json:"billing_unit"`
	SortOrder   int      `json:"sort_order"`
}
