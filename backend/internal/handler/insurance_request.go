package handler

import "github.com/animal-ekarte/backend/internal/service"

type createInsuranceRequest struct {
	Name         string `json:"name"          binding:"required"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
	CoverageRate *int   `json:"coverage_rate" binding:"omitempty,min=0,max=100"`
	ContactPhone string `json:"contact_phone"`
	SortOrder    int    `json:"sort_order"`
}

func (r createInsuranceRequest) toServiceInput() *service.CreateInsuranceInput {
	return &service.CreateInsuranceInput{
		Name:         r.Name,
		IsActive:     r.IsActive,
		Description:  r.Description,
		CoverageRate: r.CoverageRate,
		ContactPhone: r.ContactPhone,
		SortOrder:    r.SortOrder,
	}
}

type updateInsuranceRequest struct {
	Name         *string `json:"name"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	CoverageRate *int    `json:"coverage_rate" binding:"omitempty,min=0,max=100"`
	ContactPhone *string `json:"contact_phone"`
	SortOrder    *int    `json:"sort_order"`
}

func (r updateInsuranceRequest) toServiceInput() *service.UpdateInsuranceInput {
	return &service.UpdateInsuranceInput{
		Name:         r.Name,
		IsActive:     r.IsActive,
		Description:  r.Description,
		CoverageRate: r.CoverageRate,
		ContactPhone: r.ContactPhone,
		SortOrder:    r.SortOrder,
	}
}
