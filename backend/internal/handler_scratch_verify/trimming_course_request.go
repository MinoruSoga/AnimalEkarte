package handler

import "github.com/animal-ekarte/backend/internal/service"

type createTrimmingCourseRequest struct {
	Name         string  `json:"name"        binding:"required"`
	Price        *int64  `json:"price"`
	IsActive     bool    `json:"is_active"`
	Description  string  `json:"description"`
	TargetSize   string  `json:"target_size" binding:"omitempty,oneof=small medium large cat"`
	CourseTypeID *uint64 `json:"course_type_id"`
	Duration     *int    `json:"duration"`
	SortOrder    int     `json:"sort_order"`
}

func (r *createTrimmingCourseRequest) toServiceInput() *service.CreateTrimmingCourseInput {
	return &service.CreateTrimmingCourseInput{
		Name:         r.Name,
		TargetSize:   r.TargetSize,
		CourseTypeID: r.CourseTypeID,
		Price:        r.Price,
		Duration:     r.Duration,
		IsActive:     r.IsActive,
		Description:  r.Description,
		SortOrder:    r.SortOrder,
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
}

func (r updateTrimmingCourseRequest) toServiceInput() *service.UpdateTrimmingCourseInput {
	return &service.UpdateTrimmingCourseInput{
		Name:         r.Name,
		Price:        r.Price,
		IsActive:     r.IsActive,
		Description:  r.Description,
		TargetSize:   r.TargetSize,
		CourseTypeID: r.CourseTypeID,
		Duration:     r.Duration,
		SortOrder:    r.SortOrder,
	}
}
