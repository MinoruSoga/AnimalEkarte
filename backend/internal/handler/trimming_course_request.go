package handler

type createTrimmingCourseRequest struct {
	Name        string `json:"name"        binding:"required"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	TargetSize  string `json:"target_size" binding:"omitempty,oneof=small medium large cat"`
	Duration    *int   `json:"duration"`
	SortOrder   int    `json:"sort_order"`
}

type updateTrimmingCourseRequest struct {
	Name        *string `json:"name"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	TargetSize  *string `json:"target_size" binding:"omitempty,oneof=small medium large cat"`
	Duration    *int    `json:"duration"`
	SortOrder   *int    `json:"sort_order"`
}
