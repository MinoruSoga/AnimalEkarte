package handler

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/service"
)

type updateCompanyRequest clinicdomain.UpdateCompanyRequest

func (r *updateCompanyRequest) toServiceInput() *service.UpdateCompanyInput {
	return (*clinicdomain.UpdateCompanyRequest)(r).ToServiceInput()
}
