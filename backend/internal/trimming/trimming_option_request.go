package trimming

type createTrimmingOptionRequest struct {
	Name         string `json:"name"        binding:"required"`
	Price        *int64 `json:"price"`
	// IsActive / IsCombinable are *bool so omitted resolves to true; explicit false is preserved.
	IsActive     *bool  `json:"is_active"`
	Description  string `json:"description"`
	Duration     *int   `json:"duration"`
	IsCombinable *bool  `json:"is_combinable"`
	SortOrder    int    `json:"sort_order"`
}

func (r createTrimmingOptionRequest) toServiceInput() *CreateTrimmingOptionInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	isCombinable := true
	if r.IsCombinable != nil {
		isCombinable = *r.IsCombinable
	}
	return &CreateTrimmingOptionInput{
		Name:         r.Name,
		Price:        r.Price,
		IsActive:     isActive,
		Description:  r.Description,
		Duration:     r.Duration,
		IsCombinable: isCombinable,
		SortOrder:    r.SortOrder,
	}
}

type updateTrimmingOptionRequest struct {
	Name         *string `json:"name"`
	Price        *int64  `json:"price"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	Duration     *int    `json:"duration"`
	IsCombinable *bool   `json:"is_combinable"`
	SortOrder    *int    `json:"sort_order"`
}

func (r updateTrimmingOptionRequest) toServiceInput() *UpdateTrimmingOptionInput {
	return &UpdateTrimmingOptionInput{
		Name:         r.Name,
		Price:        r.Price,
		IsActive:     r.IsActive,
		Description:  r.Description,
		Duration:     r.Duration,
		IsCombinable: r.IsCombinable,
		SortOrder:    r.SortOrder,
	}
}
