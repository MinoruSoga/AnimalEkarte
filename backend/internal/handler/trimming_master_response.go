package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingCourseResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	TargetSize  string    `json:"target_size,omitempty"`
	Duration    *int      `json:"duration,omitempty"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTrimmingCourseResponse(c *model.TrimmingCourse) trimmingCourseResponse {
	targetSize := ""
	if c.TargetSize != nil {
		targetSize = string(*c.TargetSize)
	}
	return trimmingCourseResponse{
		ID:          c.ID,
		ClinicID:    c.ClinicID,
		Name:        c.Name,
		Price:       c.Price,
		IsActive:    c.IsActive,
		Description: c.Description,
		TargetSize:  targetSize,
		Duration:    c.Duration,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toTrimmingCourseResponseList(courses []model.TrimmingCourse) []trimmingCourseResponse {
	list := make([]trimmingCourseResponse, 0, len(courses))
	for i := range courses {
		list = append(list, toTrimmingCourseResponse(&courses[i]))
	}
	return list
}

type trimmingOptionResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *float64  `json:"price,omitempty"`
	IsActive    bool      `json:"is_active"`
	Description string    `json:"description"`
	Duration    *int      `json:"duration,omitempty"`
	Combinable  bool      `json:"combinable"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTrimmingOptionResponse(o *model.TrimmingOption) trimmingOptionResponse {
	return trimmingOptionResponse{
		ID:          o.ID,
		ClinicID:    o.ClinicID,
		Name:        o.Name,
		Price:       o.Price,
		IsActive:    o.IsActive,
		Description: o.Description,
		Duration:    o.Duration,
		Combinable:  o.Combinable,
		SortOrder:   o.SortOrder,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

func toTrimmingOptionResponseList(options []model.TrimmingOption) []trimmingOptionResponse {
	list := make([]trimmingOptionResponse, 0, len(options))
	for i := range options {
		list = append(list, toTrimmingOptionResponse(&options[i]))
	}
	return list
}
