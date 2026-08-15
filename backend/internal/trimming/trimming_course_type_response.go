package trimming

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type trimmingCourseTypeResponse struct {
	ID        uint64    `json:"id"`
	ClinicID  uint64    `json:"clinic_id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toTrimmingCourseTypeResponse(m *model.TrimmingCourseType) trimmingCourseTypeResponse {
	return trimmingCourseTypeResponse{
		ID:        m.ID,
		ClinicID:  m.ClinicID,
		Name:      m.Name,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: localTime(m.CreatedAt),
		UpdatedAt: localTime(m.UpdatedAt),
	}
}
