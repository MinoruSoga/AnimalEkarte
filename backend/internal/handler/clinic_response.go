package handler

import (
	"time"

	"github.com/animal-ekarte/backend/internal/model"
)

type clinicResponse struct {
	ID                 uint64    `json:"id"`
	CompanyID          uint64    `json:"company_id"`
	Name               string    `json:"name"`
	PostalCode         string    `json:"postal_code"`
	Address            string    `json:"address"`
	PhoneNumber        string    `json:"phone_number"`
	FaxNumber          string    `json:"fax_number"`
	RegistrationNumber string    `json:"registration_number"`
	DirectorName       string    `json:"director_name"`
	Email              string    `json:"email"`
	Website            string    `json:"website"`
	LogoURL            string    `json:"logo_url"`
	IsActive           bool      `json:"is_active"`
	StandardTaxRate    float64   `json:"standard_tax_rate"`
	ReducedTaxRate     float64   `json:"reduced_tax_rate"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func toClinicResponse(c *model.Clinic) clinicResponse {
	return clinicResponse{
		ID:                 c.ID,
		CompanyID:          c.CompanyID,
		Name:               c.Name,
		PostalCode:         c.PostalCode,
		Address:            c.Address,
		PhoneNumber:        c.PhoneNumber,
		FaxNumber:          c.FaxNumber,
		RegistrationNumber: c.RegistrationNumber,
		DirectorName:       c.DirectorName,
		Email:              c.Email,
		Website:            c.Website,
		LogoURL:            c.LogoURL,
		IsActive:           c.IsActive,
		StandardTaxRate:    c.StandardTaxRate,
		ReducedTaxRate:     c.ReducedTaxRate,
		CreatedAt:          localTime(c.CreatedAt),
		UpdatedAt:          localTime(c.UpdatedAt),
	}
}
