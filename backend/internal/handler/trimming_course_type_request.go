package handler

import "github.com/animal-ekarte/backend/internal/service"

type createTrimmingCourseTypeRequest struct {
	Name      string `json:"name"       binding:"required"`
	SortOrder int    `json:"sort_order"`
}

func (r createTrimmingCourseTypeRequest) toServiceInput() *service.CreateTrimmingCourseTypeInput {
	return &service.CreateTrimmingCourseTypeInput{
		Name:      r.Name,
		SortOrder: r.SortOrder,
	}
}

type updateTrimmingCourseTypeRequest struct {
	Name      *string `json:"name"`
	SortOrder *int    `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

func (r updateTrimmingCourseTypeRequest) toServiceInput() *service.UpdateTrimmingCourseTypeInput {
	return &service.UpdateTrimmingCourseTypeInput{
		Name:      r.Name,
		SortOrder: r.SortOrder,
		IsActive:  r.IsActive,
	}
}
