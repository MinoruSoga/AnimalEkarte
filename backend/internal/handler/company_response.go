package handler

import (
	clinicdomain "github.com/animal-ekarte/backend/internal/clinic"
	"github.com/animal-ekarte/backend/internal/model"
)

type companyResponse = clinicdomain.CompanyResponse

func toCompanyResponse(company *model.Company) companyResponse {
	return clinicdomain.ToCompanyResponse(company)
}
