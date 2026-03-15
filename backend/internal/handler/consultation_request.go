package handler

type createConsultationRequest struct {
	Name          string   `json:"name"           binding:"required"`
	Price         *float64 `json:"price"`
	IsActive      bool     `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	ParentID      *uint64  `json:"parent_id"`
	SortOrder     int      `json:"sort_order"`
}

type updateConsultationRequest struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	TimeCondition string   `json:"time_condition"`
	Duration      *int     `json:"duration"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     int      `json:"sort_order"`
}

// reorderConsultationRequest は診察料金並び替えのバインド struct
type reorderConsultationRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
