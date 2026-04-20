package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingCourseResponse struct {
	ID          uint64    `json:"id"`
	ClinicID    uint64    `json:"clinic_id"`
	Name        string    `json:"name"`
	Price       *int64    `json:"price,omitempty"`
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
