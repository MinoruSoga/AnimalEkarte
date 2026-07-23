package handler

import (
	"net/url"

	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
)

type listClinicHolidaysQuery = clinicdomain.ListClinicHolidaysQuery

func newListClinicHolidaysQuery(values url.Values) listClinicHolidaysQuery {
	return clinicdomain.NewListClinicHolidaysQuery(values)
}
