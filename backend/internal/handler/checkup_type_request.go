package handler

type createCheckupTypeRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	ParentID    *uint64  `json:"parent_id"`
	SortOrder   int      `json:"sort_order"`
}

type updateCheckupTypeRequest struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     int      `json:"sort_order"`
}

// reorderCheckupTypeRequest は健康診断種別並び替えのバインド struct
type reorderCheckupTypeRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
