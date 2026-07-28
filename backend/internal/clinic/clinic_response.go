package clinic

import (
	"time"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

type ClinicResponse struct {
	ID                                        uint64  `json:"id"`
	CompanyID                                 uint64  `json:"company_id"`
	Name                                      string  `json:"name"`
	PostalCode                                string  `json:"postal_code"`
	Address                                   string  `json:"address"`
	PhoneNumber                               string  `json:"phone_number"`
	FaxNumber                                 string  `json:"fax_number"`
	RegistrationNumber                        string  `json:"registration_number"`
	DirectorName                              string  `json:"director_name"`
	Email                                     string  `json:"email"`
	Website                                   string  `json:"website"`
	LogoURL                                   string  `json:"logo_url"`
	IsActive                                  bool    `json:"is_active"`
	StandardTaxRate                           float64 `json:"standard_tax_rate"`
	ReducedTaxRate                            float64 `json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                bool    `json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning bool    `json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        bool    `json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              string  `json:"accounting_document_footer_note"`
	// #190: セクション表示/非表示トグルと表示順
	AccountingDocumentShowClinicHeader   bool      `json:"accounting_document_show_clinic_header"`
	AccountingDocumentShowOwnerPetInfo   bool      `json:"accounting_document_show_owner_pet_info"`
	AccountingDocumentShowItemsTable     bool      `json:"accounting_document_show_items_table"`
	AccountingDocumentShowPaymentSummary bool      `json:"accounting_document_show_payment_summary"`
	AccountingDocumentSectionOrder       []string  `json:"accounting_document_section_order"`
	CreatedAt                            time.Time `json:"created_at"`
	UpdatedAt                            time.Time `json:"updated_at"`
}

// sectionOrderToSlice は pq.StringArray を nil セーフな []string に変換する。
// nil または空スライスはどちらも空の []string を返し、JSON では [] になる。
func SectionOrderToSlice(s pq.StringArray) []string {
	if len(s) == 0 {
		return []string{}
	}
	return []string(s)
}

func ToClinicResponse(c *model.Clinic) ClinicResponse {
	return ClinicResponse{
		ID:                         c.ID,
		CompanyID:                  c.CompanyID,
		Name:                       c.Name,
		PostalCode:                 c.PostalCode,
		Address:                    c.Address,
		PhoneNumber:                c.PhoneNumber,
		FaxNumber:                  c.FaxNumber,
		RegistrationNumber:         c.RegistrationNumber,
		DirectorName:               c.DirectorName,
		Email:                      c.Email,
		Website:                    c.Website,
		LogoURL:                    c.LogoURL,
		IsActive:                   c.IsActive,
		StandardTaxRate:            c.StandardTaxRate,
		ReducedTaxRate:             c.ReducedTaxRate,
		AccountingDocumentShowLogo: c.AccountingDocumentShowLogo,
		AccountingDocumentShowRegistrationWarning: c.AccountingDocumentShowRegistrationWarning,
		AccountingDocumentShowItemCategory:        c.AccountingDocumentShowItemCategory,
		AccountingDocumentFooterNote:              c.AccountingDocumentFooterNote,
		AccountingDocumentShowClinicHeader:        c.AccountingDocumentShowClinicHeader,
		AccountingDocumentShowOwnerPetInfo:        c.AccountingDocumentShowOwnerPetInfo,
		AccountingDocumentShowItemsTable:          c.AccountingDocumentShowItemsTable,
		AccountingDocumentShowPaymentSummary:      c.AccountingDocumentShowPaymentSummary,
		AccountingDocumentSectionOrder:            SectionOrderToSlice(c.AccountingDocumentSectionOrder),
		CreatedAt:                                 httpapi.LocalTime(c.CreatedAt),
		UpdatedAt:                                 httpapi.LocalTime(c.UpdatedAt),
	}
}
