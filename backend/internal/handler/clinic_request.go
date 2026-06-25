package handler

import (
	"net/url"

	"github.com/animal-ekarte/backend/internal/service"
)

type listClinicQuery struct {
	Scope string
}

func newListClinicQuery(values url.Values) listClinicQuery {
	return listClinicQuery{Scope: values.Get("scope")}
}

// createClinicRequest はクリニック作成リクエスト。
type createClinicRequest struct {
	Name               string `json:"name"                binding:"required"`
	PostalCode         string `json:"postal_code"`
	Address            string `json:"address"`
	PhoneNumber        string `json:"phone_number"`
	FaxNumber          string `json:"fax_number"`
	RegistrationNumber string `json:"registration_number"`
	DirectorName       string `json:"director_name"`
	Email              string `json:"email"`
	Website            string `json:"website"`
}

func (r *createClinicRequest) toServiceInput() *service.CreateClinicInput {
	return &service.CreateClinicInput{
		Name:               r.Name,
		PostalCode:         r.PostalCode,
		Address:            r.Address,
		PhoneNumber:        r.PhoneNumber,
		FaxNumber:          r.FaxNumber,
		RegistrationNumber: r.RegistrationNumber,
		DirectorName:       r.DirectorName,
		Email:              r.Email,
		Website:            r.Website,
	}
}

// updateClinicRequest はクリニック更新リクエスト。
// PATCH セマンティクス: 未送信フィールドは nil → 既存値を保持する。
type updateClinicRequest struct {
	Name                                      *string  `json:"name"`
	PostalCode                                *string  `json:"postal_code"`
	Address                                   *string  `json:"address"`
	PhoneNumber                               *string  `json:"phone_number"`
	FaxNumber                                 *string  `json:"fax_number"`
	RegistrationNumber                        *string  `json:"registration_number"`
	DirectorName                              *string  `json:"director_name"`
	Email                                     *string  `json:"email"`
	Website                                   *string  `json:"website"`
	LogoURL                                   *string  `json:"logo_url"`
	IsActive                                  *bool    `json:"is_active"`
	StandardTaxRate                           *float64 `json:"standard_tax_rate"`
	ReducedTaxRate                            *float64 `json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                *bool    `json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning *bool    `json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        *bool    `json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              *string  `json:"accounting_document_footer_note"`
}

func (r *updateClinicRequest) toServiceInput() *service.UpdateClinicInput {
	return &service.UpdateClinicInput{
		Name:                       r.Name,
		PostalCode:                 r.PostalCode,
		Address:                    r.Address,
		PhoneNumber:                r.PhoneNumber,
		FaxNumber:                  r.FaxNumber,
		RegistrationNumber:         r.RegistrationNumber,
		DirectorName:               r.DirectorName,
		Email:                      r.Email,
		Website:                    r.Website,
		LogoURL:                    r.LogoURL,
		IsActive:                   r.IsActive,
		StandardTaxRate:            r.StandardTaxRate,
		ReducedTaxRate:             r.ReducedTaxRate,
		AccountingDocumentShowLogo: r.AccountingDocumentShowLogo,
		AccountingDocumentShowRegistrationWarning: r.AccountingDocumentShowRegistrationWarning,
		AccountingDocumentShowItemCategory:        r.AccountingDocumentShowItemCategory,
		AccountingDocumentFooterNote:              r.AccountingDocumentFooterNote,
	}
}
