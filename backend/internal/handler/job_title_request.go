package handler

type createJobTitleRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=255"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
}

type updateJobTitleRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}

// reorderJobTitleRequest は役職並び替えリクエスト。
type reorderJobTitleRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
