package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type clinicHolidayResponse struct {
	ID        uint64 `json:"id"`
	ClinicID  uint64 `json:"clinic_id"`
	Date      string `json:"date"` // YYYY-MM-DD
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toClinicHolidayResponse(h *model.ClinicHoliday) clinicHolidayResponse {
	return clinicHolidayResponse{
		ID:        h.ID,
		ClinicID:  h.ClinicID,
		Date:      h.Date.In(time.Local).Format(time.DateOnly),
		Reason:    h.Reason,
		CreatedAt: h.CreatedAt.In(time.Local).Format(time.RFC3339),
		UpdatedAt: h.UpdatedAt.In(time.Local).Format(time.RFC3339),
	}
}
