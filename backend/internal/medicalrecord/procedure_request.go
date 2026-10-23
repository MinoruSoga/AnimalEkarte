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

	// EMR-454: 品目別保険設定。省略時は新規登録の初期値=保険対象外（true）を保存する。
	// 明示 null は 400（null を省略と同一視しない）。
	IsNonInsurance boolCreateField `json:"is_non_insurance"`

	// EMR-447 A01: 全体割引除外。省略時は新規登録の初期値=false を service が明示保存。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
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

		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toCreateInput(),
	}
}

type updateProcedureRequest struct {
	Version       int64    `json:"version" binding:"required,gt=0"`
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

	// EMR-454: 省略=変更なし、null=400、true/false=明示更新。
	IsNonInsurance boolUpdateField `json:"is_non_insurance"`

	// EMR-447 A01: 省略=変更なし（NULL/false/true 保持）、null/非 boolean=400、true/false=明示 SET。
	IsBillDiscountExcluded billDiscountExcludedField `json:"is_bill_discount_excluded"`
}

func (r *updateProcedureRequest) toServiceInput() *UpdateProcedureInput {
	return &UpdateProcedureInput{
		Version:       r.Version,
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

		IsNonInsurance: r.IsNonInsurance.toServiceInput(),

		IsBillDiscountExcluded: r.IsBillDiscountExcluded.toUpdateInput(),
	}
}
