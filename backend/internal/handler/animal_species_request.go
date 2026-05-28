package handler

import "github.com/animal-ekarte/backend/internal/service"

type createAnimalSpeciesRequest struct {
	Name      string `json:"name"      binding:"required"`
	IsActive  bool   `json:"is_active"`
	SortOrder int    `json:"sort_order"`
}

func (r createAnimalSpeciesRequest) toServiceInput() *service.CreateAnimalSpeciesInput {
	return &service.CreateAnimalSpeciesInput{
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

func (r updateAnimalSpeciesRequest) toServiceInput() *service.UpdateAnimalSpeciesInput {
	return &service.UpdateAnimalSpeciesInput{
		Name:      r.Name,
		IsActive:  r.IsActive,
		SortOrder: r.SortOrder,
	}
}
