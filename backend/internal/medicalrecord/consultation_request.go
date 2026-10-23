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

	// EMR-454: 品目別保険設定。省略時は新規登録の初期値=保険対象外（true）を保存する。
	// 明示 null は 400（null を省略と同一視しない）。
	IsNonInsurance boolCreateField `json:"is_non_insurance"`

	// EMR-447 A01: 全体割引除外。省略時は新規登録の初期値=false を service が明示保存。
	// 明示 null / 非 boolean は 400。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
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

		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toCreateInput(),
	}
}

type updateConsultationRequest struct {
	Version       int64    `json:"version" binding:"required,gt=0"`
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

	// EMR-454: 省略=変更なし、null=400、true/false=明示更新。
	IsNonInsurance boolUpdateField `json:"is_non_insurance"`

	// EMR-447 A01: 省略=変更なし（NULL/false/true 保持）、null/非 boolean=400、true/false=明示 SET。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
}

func (r *updateConsultationRequest) toServiceInput() *UpdateConsultationInput {
	return &UpdateConsultationInput{
		Version:       r.Version,
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

		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toUpdateInput(),
	}
}
