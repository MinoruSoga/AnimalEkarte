package trimming

type createTrimmingCourseTypeRequest struct {
	Name      string `json:"name"       binding:"required"`
	SortOrder int    `json:"sort_order"`
}

func (r createTrimmingCourseTypeRequest) toServiceInput() *CreateTrimmingCourseTypeInput {
	return &CreateTrimmingCourseTypeInput{
		Name:      r.Name,
		SortOrder: r.SortOrder,
	}
}

type updateTrimmingCourseTypeRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

func (r updateTrimmingCourseTypeRequest) toServiceInput() *UpdateTrimmingCourseTypeInput {
	return &UpdateTrimmingCourseTypeInput{
		Name:      r.Name,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
	}
}
