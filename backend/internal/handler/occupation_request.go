package handler

import "github.com/animal-ekarte/backend/internal/service"

type createOccupationRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (r createOccupationRequest) toServiceInput() *service.CreateOccupationInput {
	return &service.CreateOccupationInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}

type updateOccupationRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateOccupationRequest) toServiceInput() *service.UpdateOccupationInput {
	return &service.UpdateOccupationInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}
