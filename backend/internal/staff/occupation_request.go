package staff

type createOccupationRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"max=2000"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (r createOccupationRequest) toServiceInput() *CreateOccupationInput {
	return &CreateOccupationInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
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
