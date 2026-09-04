package clinic

type UpdateCompanyRequest struct {
	Name                      *string `json:"name" binding:"omitempty,max=255"`
	PostalCode                *string `json:"postal_code" binding:"omitempty,jp_postal"`
	Address                   *string `json:"address" binding:"omitempty,max=500"`
	PhoneNumber               *string `json:"phone_number" binding:"omitempty,jp_phone"`
	FaxNumber                 *string `json:"fax_number" binding:"omitempty,jp_phone"`
	Email                     *string `json:"email" binding:"omitempty,jp_email"`
	Website                   *string `json:"website" binding:"omitempty,max=500"`
	DirectorName              *string `json:"director_name" binding:"omitempty,max=255"`
	RegistrationNumber        *string `json:"registration_number" binding:"omitempty,max=100"`
	InvoiceRegistrationNumber *string `json:"invoice_registration_number" binding:"omitempty,max=100"`
	LogoURL                   *string `json:"logo_url" binding:"omitempty,max=500"`
}

func (r *UpdateCompanyRequest) ToServiceInput() *UpdateCompanyInput {
	return &UpdateCompanyInput{
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
