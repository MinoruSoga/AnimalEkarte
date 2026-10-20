package medicalrecord

type createCheckupTypeRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Price       *int64  `json:"price"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Interval    string  `json:"interval"`
	TargetAge   string  `json:"target_age"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
	// EMR-456: 未指定/null は新規初期値（保険対象外=true）を service が明示保存する。
	IsNonInsurance *bool `json:"is_non_insurance"`
}

func (r *createCheckupTypeRequest) toServiceInput() *CreateCheckupTypeInput {
	return &CreateCheckupTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Interval:       r.Interval,
		TargetAge:      r.TargetAge,
		ParentID:       r.ParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
	}
}

type updateCheckupTypeRequest struct {
	Name          *string `json:"name"`
	Price         *int64  `json:"price"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	Interval      *string `json:"interval"`
	TargetAge     *string `json:"target_age"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	SortOrder     *int    `json:"sort_order"`
	// EMR-456: 未指定/null は変更なし（NULL=未保存へ戻す経路は持たない）。
	IsNonInsurance *bool `json:"is_non_insurance"`
	// EMR-456: 楽観ロックの期待版。未指定時は service が読取時点の版で照合する。
	Version *int64 `json:"version" binding:"omitempty,min=1"`
}

func (r updateCheckupTypeRequest) toServiceInput() *UpdateCheckupTypeInput {
	return &UpdateCheckupTypeInput{
		Name:           r.Name,
		Price:          r.Price,
		IsActive:       r.IsActive,
		Description:    r.Description,
		Interval:       r.Interval,
		TargetAge:      r.TargetAge,
		ParentID:       r.ParentID,
		ClearParentID:  r.ClearParentID,
		SortOrder:      r.SortOrder,
		IsNonInsurance: r.IsNonInsurance,
		Version:        r.Version,
	}
}
