package pet

type createAnimalSpeciesRequest struct {
	Name      string `json:"name"      binding:"required"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

func (r createAnimalSpeciesRequest) toServiceInput() *CreateAnimalSpeciesInput {
	return &CreateAnimalSpeciesInput{
		Name:      r.Name,
		IsActive:  r.IsActive,
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
