package handler

type updateCompanyRequest struct {
	Name               *string `json:"name"`
	PostalCode         *string `json:"postal_code"`
	Address            *string `json:"address"`
	PhoneNumber        *string `json:"phone_number"`
	FaxNumber          *string `json:"fax_number"`
	Email              *string `json:"email"`
	Website            *string `json:"website"`
	DirectorName       *string `json:"director_name"`
	RegistrationNumber *string `json:"registration_number"`
	LogoURL            *string `json:"logo_url"`
}
