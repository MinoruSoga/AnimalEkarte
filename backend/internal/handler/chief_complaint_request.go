package handler

type createChiefComplaintRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type updateChiefComplaintRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}
