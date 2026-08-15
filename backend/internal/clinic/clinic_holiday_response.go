package clinic

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicHolidayResponse struct {
	ID        uint64 `json:"id"`
	ClinicID  uint64 `json:"clinic_id"`
	Date      string `json:"date"` // YYYY-MM-DD
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func ToClinicHolidayResponse(h *model.ClinicHoliday) ClinicHolidayResponse {
	return ClinicHolidayResponse{
		ID:        h.ID,
		ClinicID:  h.ClinicID,
		Date:      h.Date.In(time.Local).Format(time.DateOnly),
		Reason:    h.Reason,
		CreatedAt: httpapi.LocalTimeRFC3339(h.CreatedAt),
		UpdatedAt: httpapi.LocalTimeRFC3339(h.UpdatedAt),
	}
}
