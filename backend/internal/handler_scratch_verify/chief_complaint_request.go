package handler

import "github.com/animal-ekarte/backend/internal/service"

type createChiefComplaintRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

func (r createChiefComplaintRequest) toServiceInput() *service.CreateChiefComplaintTypeInput {
	return &service.CreateChiefComplaintTypeInput{
		Name:        r.Name,
		Description: r.Description,
		IsActive:    r.IsActive,
		SortOrder:   r.SortOrder,
	}
}

type updateChiefComplaintRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateChiefComplaintRequest) toServiceInput() *service.UpdateChiefComplaintTypeInput {
	return &service.UpdateChiefComplaintTypeInput{
		Name:        r.Name,
		Description: r.Description,
		SortOrder:   r.SortOrder,
		IsActive:    r.IsActive,
	}
}
