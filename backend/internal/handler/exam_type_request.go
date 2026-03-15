package handler

type createExaminationTypeRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	ParentID    *uint64  `json:"parent_id"`
	SortOrder   int      `json:"sort_order"`
}

type updateExaminationTypeRequest struct {
	Name          string   `json:"name"`
	Price         *float64 `json:"price"`
	IsActive      *bool    `json:"is_active"`
	Description   string   `json:"description"`
	ParentID      *uint64  `json:"parent_id"`
	ClearParentID bool     `json:"clear_parent_id"`
	SortOrder     int      `json:"sort_order"`
}

// reorderExaminationTypeRequest は検査種別並び替えのバインド struct
type reorderExaminationTypeRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

type examTypeItemInput struct {
	Name            string `json:"name"             binding:"required"`
	InspectionValue string `json:"inspection_value"`
	NormalValue     string `json:"normal_value"`
	SortOrder       int    `json:"sort_order"`
}

type replaceExamTypeItemsRequest struct {
	Items []examTypeItemInput `json:"items" binding:"required"`
}
