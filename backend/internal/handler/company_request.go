package handler

import "github.com/animal-ekarte/backend/internal/service"

type updateCompanyRequest struct {
	Name                      *string `json:"name"`
	PostalCode                *string `json:"postal_code"`
	Address                   *string `json:"address"`
	PhoneNumber               *string `json:"phone_number"`
	FaxNumber                 *string `json:"fax_number"`
	Email                     *string `json:"email"`
	Website                   *string `json:"website"`
	DirectorName              *string `json:"director_name"`
	RegistrationNumber        *string `json:"registration_number"`
	InvoiceRegistrationNumber *string `json:"invoice_registration_number"`
	LogoURL                   *string `json:"logo_url"`
}

func (r updateCompanyRequest) toServiceInput() *service.UpdateCompanyInput {
	return &service.UpdateCompanyInput{
		Name:                      r.Name,
		PostalCode:                r.PostalCode,
		Address:                   r.Address,
		PhoneNumber:               r.PhoneNumber,
		FaxNumber:                 r.FaxNumber,
		Email:                     r.Email,
		Website:                   r.Website,
		DirectorName:              r.DirectorName,
		RegistrationNumber:        r.RegistrationNumber,
		InvoiceRegistrationNumber: r.InvoiceRegistrationNumber,
		LogoURL:                   r.LogoURL,
	}
}
