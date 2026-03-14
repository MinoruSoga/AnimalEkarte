package handler

type createVaccineRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}

type updateVaccineRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	SortOrder   int      `json:"sort_order"`
}
