package handler

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/service"
)

type listVaccinesQuery struct {
	Species string
}

func newListVaccinesQuery(values url.Values) listVaccinesQuery {
	return listVaccinesQuery{Species: values.Get("species")}
}

func (q listVaccinesQuery) toServiceFilter() *string {
	return optionalStringQueryFilter(q.Species)
}

type createVaccineRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Price       *int64  `json:"price"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Species     string  `json:"species"     binding:"omitempty,oneof=dog cat both"`
	Interval    string  `json:"interval"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
}

func (r createVaccineRequest) toServiceInput() *service.CreateVaccineInput {
	return &service.CreateVaccineInput{
		Name:        r.Name,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		Species:     nilIfEmpty(r.Species),
		Interval:    r.Interval,
		ParentID:    r.ParentID,
		SortOrder:   r.SortOrder,
	}
}

type updateVaccineRequest struct {
	Name          *string `json:"name"`
	Price         *int64  `json:"price"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	Species       *string `json:"species"      binding:"omitempty,oneof=dog cat both"`
	Interval      *string `json:"interval"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	SortOrder     *int    `json:"sort_order"`
}

func (r updateVaccineRequest) toServiceInput() *service.UpdateVaccineInput {
	return &service.UpdateVaccineInput{
		Name:          r.Name,
		Price:         r.Price,
		IsActive:      r.IsActive,
		Description:   r.Description,
		Species:       r.Species,
		Interval:      r.Interval,
		ParentID:      r.ParentID,
		ClearParentID: r.ClearParentID,
		SortOrder:     r.SortOrder,
	}
}
