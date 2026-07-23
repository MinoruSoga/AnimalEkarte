package handler

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
)

type clinicHolidayResponse = clinicdomain.ClinicHolidayResponse

func toClinicHolidayResponse(holiday *model.ClinicHoliday) clinicHolidayResponse {
	return clinicdomain.ToClinicHolidayResponse(holiday)
}
