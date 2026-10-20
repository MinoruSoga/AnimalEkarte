package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingCourseResponse struct {
	ID           uint64    `json:"id"`
	ClinicID     uint64    `json:"clinic_id"`
	Name         string    `json:"name"`
	Price        *int64    `json:"price,omitempty"`
	IsActive     bool      `json:"is_active"`
	Description  string    `json:"description"`
	TargetSize   *string   `json:"target_size,omitempty"`
	CourseTypeID *uint64   `json:"course_type_id,omitempty"`
	Duration     *int      `json:"duration,omitempty"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// EMR-457: null は未登録。allowed_weekdays の null は全曜日（0=日〜6=土）。
	ShortName       *string `json:"short_name"`
	ImageObjectKey  *string `json:"image_object_key"`
	ImageURL        *string `json:"image_url"`
	AllowedWeekdays []int   `json:"allowed_weekdays"`
}

func toTrimmingCourseResponse(c *model.TrimmingCourse) trimmingCourseResponse {
	var targetSize *string
	if c.TargetSize != nil {
		s := string(*c.TargetSize)
		targetSize = &s
	}
	sortOrder := 0
	if c.SortOrder != nil {
		sortOrder = *c.SortOrder
	}
	var weekdays []int
	if c.AllowedWeekdays != nil {
		weekdays = make([]int, len(*c.AllowedWeekdays))
		for i, day := range *c.AllowedWeekdays {
			weekdays[i] = int(day)
		}
	}
	return trimmingCourseResponse{
		ID:           c.ID,
		ClinicID:     c.ClinicID,
		Name:         c.Name,
		Price:        c.Price,
		IsActive:     c.IsActive,
		Description:  c.Description,
		TargetSize:   targetSize,
		CourseTypeID: c.CourseTypeID,
		Duration:     c.Duration,
		SortOrder:    sortOrder,
		CreatedAt:    localTime(c.CreatedAt),
		UpdatedAt:    localTime(c.UpdatedAt),

		ShortName:       c.ShortName,
		ImageObjectKey:  c.ImageObjectKey,
		AllowedWeekdays: weekdays,
	}
}
