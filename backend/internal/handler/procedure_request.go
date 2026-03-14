package handler

type createProcedureRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Duration    *int     `json:"duration"`
	Anesthesia  string   `json:"anesthesia"`
	ParentID    *uint64  `json:"parent_id"`
	SortOrder   int      `json:"sort_order"`
}

type updateProcedureRequest struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	Duration      *int     `json:"duration"`
	Anesthesia    string   `json:"anesthesia"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     int      `json:"sort_order"`
}
