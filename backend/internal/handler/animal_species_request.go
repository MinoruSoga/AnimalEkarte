package handler

type createAnimalSpeciesRequest struct {
	Name      string `json:"name"      binding:"required"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

type updateAnimalSpeciesRequest struct {
	Name      *string `json:"name"`
	IsActive  *bool   `json:"is_active"`
	SortOrder *int    `json:"sort_order"`
}
