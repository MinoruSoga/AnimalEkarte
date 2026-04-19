package handler

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

// reorderVaccineRequest はワクチン並び替えのバインド struct
type reorderVaccineRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
