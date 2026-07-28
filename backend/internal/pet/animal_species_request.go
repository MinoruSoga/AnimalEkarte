package pet

// createAnimalSpeciesRequest is the presence-aware create body for animal species.
// IsActive is *bool so JSON binding can distinguish omitted / false / true.
type createAnimalSpeciesRequest struct {
	Name      string `json:"name"      binding:"required"`
	IsActive  *bool  `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

// toServiceInput maps the request to the pet use-case input.
// Omitted is_active (nil) resolves to true; explicit false/true are preserved.
func (r createAnimalSpeciesRequest) toServiceInput() *CreateAnimalSpeciesInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreateAnimalSpeciesInput{
		Name:      r.Name,
		IsActive:  isActive,
		SortOrder: r.SortOrder,
	}
}

type updateAnimalSpeciesRequest struct {
	Name      *string `json:"name"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}

func (r updateAnimalSpeciesRequest) toServiceInput() *UpdateAnimalSpeciesInput {
	return &UpdateAnimalSpeciesInput{
		Name:      r.Name,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}
