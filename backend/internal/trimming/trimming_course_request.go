package trimming

type createTrimmingCourseRequest struct {
	Name  string `json:"name"        binding:"required"`
	Price *int64 `json:"price"`
	// IsActive is *bool so JSON binding can distinguish omitted / false / true.
	// Omitted (nil) resolves to true in toServiceInput.
	IsActive     *bool   `json:"is_active"`
	Description  string  `json:"description"`
	TargetSize   string  `json:"target_size" binding:"omitempty,oneof=small medium large cat"`
	CourseTypeID *uint64 `json:"course_type_id"`
	Duration     *int    `json:"duration"`
	SortOrder    int     `json:"sort_order"`
	// ShortName / ImageObjectKey are optional; empty means not registered (NULL).
	ShortName      string `json:"short_name"`
	ImageObjectKey string `json:"image_object_key"`
	// AllowedWeekdays uses 0=Sunday..6=Saturday. Omitted or empty means every weekday.
	AllowedWeekdays []int `json:"allowed_weekdays"`
}

func (r *createTrimmingCourseRequest) toServiceInput() *CreateTrimmingCourseInput {
	isActive := true
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return &CreateTrimmingCourseInput{
		Name:            r.Name,
		TargetSize:      r.TargetSize,
		CourseTypeID:    r.CourseTypeID,
		Price:           r.Price,
		Duration:        r.Duration,
		IsActive:        isActive,
		Description:     r.Description,
		SortOrder:       r.SortOrder,
		ShortName:       r.ShortName,
		ImageObjectKey:  r.ImageObjectKey,
		AllowedWeekdays: r.AllowedWeekdays,
	}
}

type updateTrimmingCourseRequest struct {
	Name         *string `json:"name"`
	Price        *int64  `json:"price"`
	IsActive     *bool   `json:"is_active"`
	Description  *string `json:"description"`
	TargetSize   *string `json:"target_size" binding:"omitempty,oneof=small medium large cat"`
	CourseTypeID *uint64 `json:"course_type_id"`
	Duration     *int    `json:"duration"`
	SortOrder    *int    `json:"sort_order"`
	// Omitted keeps the stored value; "" clears ShortName / ImageObjectKey.
	ShortName      *string `json:"short_name"`
	ImageObjectKey *string `json:"image_object_key"`
	// Omitted keeps the stored value; [] clears to every weekday.
	AllowedWeekdays *[]int `json:"allowed_weekdays"`
}

func (r updateTrimmingCourseRequest) toServiceInput() *UpdateTrimmingCourseInput {
	return &UpdateTrimmingCourseInput{
		Name:            r.Name,
		Price:           r.Price,
		IsActive:        r.IsActive,
		Description:     r.Description,
		TargetSize:      r.TargetSize,
		CourseTypeID:    r.CourseTypeID,
		Duration:        r.Duration,
		SortOrder:       r.SortOrder,
		ShortName:       r.ShortName,
		ImageObjectKey:  r.ImageObjectKey,
		AllowedWeekdays: r.AllowedWeekdays,
	}
}
