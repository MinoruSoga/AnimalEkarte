package handler

type createTrimmingCourseRequest struct {
	Name        string `json:"name"        binding:"required"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	TargetSize  string `json:"target_size"`
	Duration    *int   `json:"duration"`
	SortOrder   int    `json:"sort_order"`
}

type updateTrimmingCourseRequest struct {
	Name        *string `json:"name"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	TargetSize  *string `json:"target_size"`
	Duration    *int    `json:"duration"`
	SortOrder   *int    `json:"sort_order"`
}

type createTrimmingOptionRequest struct {
	Name        string `json:"name"        binding:"required"`
	Price       *int64 `json:"price"`
	IsActive    bool   `json:"is_active"`
	Description string `json:"description"`
	Duration    *int   `json:"duration"`
	Combinable  bool   `json:"combinable"`
	SortOrder   int    `json:"sort_order"`
}

type updateTrimmingOptionRequest struct {
	Name        *string `json:"name"`
	Price       *int64  `json:"price"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
	Duration    *int    `json:"duration"`
	Combinable  *bool   `json:"combinable"`
	SortOrder   *int    `json:"sort_order"`
}

// reorderTrimmingCourseRequest はトリミングコース並び替えリクエスト。
type reorderTrimmingCourseRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}

// reorderTrimmingOptionRequest はトリミングオプション並び替えリクエスト。
type reorderTrimmingOptionRequest struct {
	IDs []uint64 `json:"ids" binding:"required,min=1"`
}
