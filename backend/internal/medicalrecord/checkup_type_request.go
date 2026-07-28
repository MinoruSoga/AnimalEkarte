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
}

func (r *createCheckupTypeRequest) toServiceInput() *CreateCheckupTypeInput {
	return &CreateCheckupTypeInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		Interval:    r.Interval,
		TargetAge:   r.TargetAge,
		ParentID:    r.ParentID,
		SortOrder:   r.SortOrder,
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
}

func (r updateCheckupTypeRequest) toServiceInput() *UpdateCheckupTypeInput {
	return &UpdateCheckupTypeInput{
		Name:          r.Name,
		Price:         r.Price,
		IsActive:      r.IsActive,
		Description:   r.Description,
		Interval:      r.Interval,
		TargetAge:     r.TargetAge,
		ParentID:      r.ParentID,
		ClearParentID: r.ClearParentID,
		SortOrder:     r.SortOrder,
	}
}
