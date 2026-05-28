// Package handler provides HTTP handler implementations for Cage entity.
package handler

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/service"
)

type listCageQuery struct {
	CageType string
}

func newListCageQuery(values url.Values) listCageQuery {
	return listCageQuery{CageType: values.Get("cage_type")}
}

func (q listCageQuery) toServiceFilter() *string {
	return optionalStringQueryFilter(q.CageType)
}

type createCageRequest struct {
	Name        string `json:"name"        binding:"required"`
	CageType    string `json:"cage_type"   binding:"required,oneof=icu dog cat general"`
	CageSize    string `json:"cage_size"   binding:"required,oneof=small medium large"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func (r createCageRequest) toServiceInput() *service.CreateCageInput {
	return &service.CreateCageInput{
		Name:        r.Name,
		CageType:    r.CageType,
		CageSize:    r.CageSize,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}

type updateCageRequest struct {
	Name        *string `json:"name"`
	CageType    *string `json:"cage_type"    binding:"omitempty,oneof=icu dog cat general"`
	CageSize    *string `json:"cage_size"    binding:"omitempty,oneof=small medium large"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

func (r updateCageRequest) toServiceInput() *service.UpdateCageInput {
	return &service.UpdateCageInput{
		Name:        r.Name,
		CageType:    r.CageType,
		CageSize:    r.CageSize,
		Price:       r.Price,
		IsActive:    r.IsActive,
		Description: r.Description,
		SortOrder:   r.SortOrder,
	}
}
