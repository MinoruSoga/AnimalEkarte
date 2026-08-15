package clinic

import (
	"time"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type CompanyResponse struct {
	ID                        uint64    `json:"id"`
	Name                      string    `json:"name"`
	PostalCode                string    `json:"postal_code"`
	Address                   string    `json:"address"`
	PhoneNumber               string    `json:"phone_number"`
	FaxNumber                 string    `json:"fax_number"`
	Email                     string    `json:"email"`
	Website                   string    `json:"website"`
	DirectorName              string    `json:"director_name"`
	RegistrationNumber        string    `json:"registration_number"`
	InvoiceRegistrationNumber string    `json:"invoice_registration_number"`
	LogoURL                   string    `json:"logo_url"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

func ToCompanyResponse(c *model.Company) CompanyResponse {
	return CompanyResponse{
		ID:                        c.ID,
		Name:                      c.Name,
		PostalCode:                c.PostalCode,
		Address:                   c.Address,
		PhoneNumber:               c.PhoneNumber,
		FaxNumber:                 c.FaxNumber,
		Email:                     c.Email,
		Website:                   c.Website,
		DirectorName:              c.DirectorName,
		RegistrationNumber:        c.RegistrationNumber,
		InvoiceRegistrationNumber: c.InvoiceRegistrationNumber,
		LogoURL:                   c.LogoURL,
		CreatedAt:                 httpapi.LocalTime(c.CreatedAt),
		UpdatedAt:                 httpapi.LocalTime(c.UpdatedAt),
	}
}
