package handler

type createCheckupTypeRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Price       *int64  `json:"price"`
	IsActive    bool    `json:"is_active"`
	Description string  `json:"description"`
	Interval    string  `json:"interval"`
	TargetAge   string  `json:"target_age"`
	ParentID    *uint64 `json:"parent_id"`
	SortOrder   int     `json:"sort_order"`
}

type updateCheckupTypeRequest struct {
	Name          *string `json:"name"`
	Price         *int64  `json:"price"`
	IsActive      *bool   `json:"is_active"`
	Description   *string `json:"description"`
	Interval      *string `json:"interval"`
	TargetAge     *string `json:"target_age"`
	ParentID      *uint64 `json:"parent_id"`
	ClearParentID bool    `json:"clear_parent_id"`
	SortOrder     *int    `json:"sort_order"`
}
