package handler

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/service"
)

type updateClinicSettingsRequest clinicdomain.UpdateClinicSettingsRequest

func (r updateClinicSettingsRequest) toServiceInput() service.UpdateClinicSettingsInput {
	return clinicdomain.UpdateClinicSettingsRequest(r).ToServiceInput()
}

type createSpecialPeriodRequest clinicdomain.CreateSpecialPeriodRequest

func (r *createSpecialPeriodRequest) toServiceInput() *service.CreateSpecialPeriodInput {
	return (*clinicdomain.CreateSpecialPeriodRequest)(r).ToServiceInput()
}

type updateSpecialPeriodRequest clinicdomain.UpdateSpecialPeriodRequest

func (r updateSpecialPeriodRequest) toServiceInput() service.UpdateSpecialPeriodInput {
	return clinicdomain.UpdateSpecialPeriodRequest(r).ToServiceInput()
}
