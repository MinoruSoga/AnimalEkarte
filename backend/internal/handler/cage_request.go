// Package handler provides HTTP handler implementations for Cage entity.
package handler

type createCageRequest struct {
	Name        string `json:"name"        binding:"required"`
	CageType    string `json:"cage_type"   binding:"required,oneof=icu dog cat general"`
	CageSize    string `json:"cage_size"   binding:"required,oneof=small medium large"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type updateCageRequest struct {
	Name        *string `json:"name"`
	CageType    *string `json:"cage_type"`
	CageSize    *string `json:"cage_size"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}
