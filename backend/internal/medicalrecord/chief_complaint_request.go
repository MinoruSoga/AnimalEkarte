package medicalrecord

// createChiefComplaintRequest is the presence-aware create body for chief complaint types.
// IsActive is *bool so JSON binding can distinguish omitted / false / true.
type createChiefComplaintRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

// toServiceInput maps the request to the medicalrecord use-case input.
// Omitted is_active (nil) resolves to true; explicit false/true are preserved.
func (r createChiefComplaintRequest) toServiceInput() *CreateChiefComplaintTypeInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreateChiefComplaintTypeInput{
		Name:        r.Name,
		Description: r.Description,
		IsActive:    isActive,
		SortOrder:   r.SortOrder,
	}
}

type updateChiefComplaintRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateChiefComplaintRequest) toServiceInput() *UpdateChiefComplaintTypeInput {
	return &UpdateChiefComplaintTypeInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}
