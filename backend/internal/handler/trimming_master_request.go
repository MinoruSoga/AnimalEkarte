package handler

type createTrimmingCourseRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	TargetSize  string   `json:"target_size"`
	Duration    *int     `json:"duration"`
	SortOrder   int      `json:"sort_order"`
}

type updateTrimmingCourseRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	TargetSize  string   `json:"target_size"`
	Duration    *int     `json:"duration"`
	SortOrder   int      `json:"sort_order"`
}

type createTrimmingOptionRequest struct {
	Name        string   `json:"name"        binding:"required"`
	Price       *float64 `json:"price"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Duration    *int     `json:"duration"`
	Combinable  bool     `json:"combinable"`
	SortOrder   int      `json:"sort_order"`
}

type updateTrimmingOptionRequest struct {
	Name        string   `json:"name"`
	Price       *float64 `json:"price"`
	IsActive    *bool    `json:"is_active"`
	Description string   `json:"description"`
	Duration    *int     `json:"duration"`
	Combinable  *bool    `json:"combinable"`
	SortOrder   int      `json:"sort_order"`
}
