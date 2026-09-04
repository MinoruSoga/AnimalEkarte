package clinic

import (
	"net/url"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// POC-17: same contact formats as owner/validators_contact (email/phone/postal).
var (
	clinicEmailPattern      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	clinicPhonePattern      = regexp.MustCompile(`^0\d{1,4}-?\d{1,4}-?\d{4}$`)
	clinicPostalCodePattern = regexp.MustCompile(`^\d{3}-?\d{4}$`)
)

func init() {
	engine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	_ = engine.RegisterValidation("jp_email", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || clinicEmailPattern.MatchString(s)
	})
	_ = engine.RegisterValidation("jp_phone", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || clinicPhonePattern.MatchString(s)
	})
	_ = engine.RegisterValidation("jp_postal", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return s == "" || clinicPostalCodePattern.MatchString(s)
	})
}

type ListClinicQuery struct {
	Scope string
}

func NewListClinicQuery(values url.Values) ListClinicQuery {
	return ListClinicQuery{Scope: values.Get("scope")}
}

// createClinicRequest はクリニック作成リクエスト。
type CreateClinicRequest struct {
	Name               string `json:"name"                binding:"required,max=255"`
	PostalCode         string `json:"postal_code"         binding:"omitempty,jp_postal"`
	Address            string `json:"address"             binding:"omitempty,max=500"`
	PhoneNumber        string `json:"phone_number"        binding:"omitempty,jp_phone"`
	FaxNumber          string `json:"fax_number"          binding:"omitempty,jp_phone"`
	RegistrationNumber string `json:"registration_number" binding:"omitempty,max=100"`
	DirectorName       string `json:"director_name"       binding:"omitempty,max=255"`
	Email              string `json:"email"               binding:"omitempty,jp_email"`
	Website            string `json:"website"             binding:"omitempty,max=500"`
}

func (r *CreateClinicRequest) ToServiceInput() *CreateClinicInput {
	return &CreateClinicInput{
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
type UpdateClinicRequest struct {
	Name                                      *string  `json:"name" binding:"omitempty,max=255"`
	PostalCode                                *string  `json:"postal_code" binding:"omitempty,jp_postal"`
	Address                                   *string  `json:"address" binding:"omitempty,max=500"`
	PhoneNumber                               *string  `json:"phone_number" binding:"omitempty,jp_phone"`
	FaxNumber                                 *string  `json:"fax_number" binding:"omitempty,jp_phone"`
	RegistrationNumber                        *string  `json:"registration_number" binding:"omitempty,max=100"`
	DirectorName                              *string  `json:"director_name" binding:"omitempty,max=255"`
	Email                                     *string  `json:"email" binding:"omitempty,jp_email"`
	Website                                   *string  `json:"website" binding:"omitempty,max=500"`
	LogoURL                                   *string  `json:"logo_url"`
	IsActive                                  *bool    `json:"is_active"`
	StandardTaxRate                           *float64 `json:"standard_tax_rate"`
	ReducedTaxRate                            *float64 `json:"reduced_tax_rate"`
	AccountingDocumentShowLogo                *bool    `json:"accounting_document_show_logo"`
	AccountingDocumentShowRegistrationWarning *bool    `json:"accounting_document_show_registration_warning"`
	AccountingDocumentShowItemCategory        *bool    `json:"accounting_document_show_item_category"`
	AccountingDocumentFooterNote              *string  `json:"accounting_document_footer_note" binding:"omitempty,max=1000"`
	// #190: セクション表示/非表示トグルと表示順 (migration 010)
	AccountingDocumentShowClinicHeader   *bool     `json:"accounting_document_show_clinic_header"`
	AccountingDocumentShowOwnerPetInfo   *bool     `json:"accounting_document_show_owner_pet_info"`
	AccountingDocumentShowItemsTable     *bool     `json:"accounting_document_show_items_table"`
	AccountingDocumentShowPaymentSummary *bool     `json:"accounting_document_show_payment_summary"`
	AccountingDocumentSectionOrder       *[]string `json:"accounting_document_section_order"`
}

func (r *UpdateClinicRequest) ToServiceInput() *UpdateClinicInput {
	return &UpdateClinicInput{
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
		AccountingDocumentShowClinicHeader:        r.AccountingDocumentShowClinicHeader,
		AccountingDocumentShowOwnerPetInfo:        r.AccountingDocumentShowOwnerPetInfo,
		AccountingDocumentShowItemsTable:          r.AccountingDocumentShowItemsTable,
		AccountingDocumentShowPaymentSummary:      r.AccountingDocumentShowPaymentSummary,
		AccountingDocumentSectionOrder:            r.AccountingDocumentSectionOrder,
	}
}
