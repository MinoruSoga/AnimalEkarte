package handler

import (
	"net/url"

	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/service"
)

type listClinicQuery = clinicdomain.ListClinicQuery

func newListClinicQuery(values url.Values) listClinicQuery {
	return clinicdomain.NewListClinicQuery(values)
}

type createClinicRequest clinicdomain.CreateClinicRequest

func (r *createClinicRequest) toServiceInput() *service.CreateClinicInput {
	return (*clinicdomain.CreateClinicRequest)(r).ToServiceInput()
}

type updateClinicRequest clinicdomain.UpdateClinicRequest

func (r *updateClinicRequest) toServiceInput() *service.UpdateClinicInput {
	return (*clinicdomain.UpdateClinicRequest)(r).ToServiceInput()
}
