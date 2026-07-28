package staff

type createOccupationRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=2000"`
	// IsActive is *bool so JSON binding can distinguish omitted / false / true.
	// Omitted (nil) resolves to true in toServiceInput.
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (r createOccupationRequest) toServiceInput() *CreateOccupationInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreateOccupationInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    isActive,
	}
}

type updateOccupationRequest struct {
	Name        *string `json:"name"        binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty,max=2000"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateOccupationRequest) toServiceInput() *UpdateOccupationInput {
	return &UpdateOccupationInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}
