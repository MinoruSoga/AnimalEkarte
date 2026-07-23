package handler

import (
	"github.com/lib/pq"

	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
)

type clinicResponse = clinicdomain.ClinicResponse

func sectionOrderToSlice(order pq.StringArray) []string {
	return clinicdomain.SectionOrderToSlice(order)
}

func toClinicResponse(value *model.Clinic) clinicResponse {
	return clinicdomain.ToClinicResponse(value)
}
