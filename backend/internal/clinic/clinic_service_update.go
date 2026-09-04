package clinic

import (
	"fmt"

	"github.com/lib/pq"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func applyClinicProfileFields(fields map[string]any, input *UpdateClinicInput) {
	if input.Name != nil {
		fields[colClinicName] = *input.Name
	}
	if input.PostalCode != nil {
		fields[colClinicPostalCode] = *input.PostalCode
	}
	if input.Address != nil {
		fields[colClinicAddress] = *input.Address
	}
	if input.PhoneNumber != nil {
		fields[colClinicPhoneNumber] = *input.PhoneNumber
	}
	if input.FaxNumber != nil {
		fields[colClinicFaxNumber] = *input.FaxNumber
	}
	if input.RegistrationNumber != nil {
		fields[colClinicRegistrationNumber] = *input.RegistrationNumber
	}
	if input.DirectorName != nil {
		fields[colClinicDirectorName] = *input.DirectorName
	}
	if input.Email != nil {
		fields[colClinicEmail] = *input.Email
	}
	if input.Website != nil {
		fields[colClinicWebsite] = *input.Website
	}
	if input.LogoURL != nil {
		fields[colClinicLogoURL] = *input.LogoURL
	}
	if input.IsActive != nil {
		fields[colClinicIsActive] = *input.IsActive
	}
}

func applyClinicTaxFields(fields map[string]any, input *UpdateClinicInput) error {
	if input.StandardTaxRate != nil {
		r := *input.StandardTaxRate
		if r < 0 || r > 1 {
			return apperrors.WrapInvalidInput("standard_tax_rate must be between 0 and 1")
		}
		fields[colClinicStandardTaxRate] = r
	}
	if input.ReducedTaxRate != nil {
		r := *input.ReducedTaxRate
		if r < 0 || r > 1 {
			return apperrors.WrapInvalidInput("reduced_tax_rate must be between 0 and 1")
		}
		fields[colClinicReducedTaxRate] = r
	}
	return nil
}

func applyClinicAccountingDocumentFlags(fields map[string]any, input *UpdateClinicInput) {
	if input.AccountingDocumentShowLogo != nil {
		fields[colClinicAccountingDocumentShowLogo] = *input.AccountingDocumentShowLogo
	}
	if input.AccountingDocumentShowRegistrationWarning != nil {
		fields[colClinicAccountingDocumentShowRegistrationWarning] = *input.AccountingDocumentShowRegistrationWarning
	}
	if input.AccountingDocumentShowItemCategory != nil {
		fields[colClinicAccountingDocumentShowItemCategory] = *input.AccountingDocumentShowItemCategory
	}
	if input.AccountingDocumentFooterNote != nil {
		fields[colClinicAccountingDocumentFooterNote] = *input.AccountingDocumentFooterNote
	}
	if input.AccountingDocumentShowClinicHeader != nil {
		fields[colClinicAccountingDocumentShowClinicHeader] = *input.AccountingDocumentShowClinicHeader
	}
	if input.AccountingDocumentShowOwnerPetInfo != nil {
		fields[colClinicAccountingDocumentShowOwnerPetInfo] = *input.AccountingDocumentShowOwnerPetInfo
	}
	if input.AccountingDocumentShowItemsTable != nil {
		fields[colClinicAccountingDocumentShowItemsTable] = *input.AccountingDocumentShowItemsTable
	}
	if input.AccountingDocumentShowPaymentSummary != nil {
		fields[colClinicAccountingDocumentShowPaymentSummary] = *input.AccountingDocumentShowPaymentSummary
	}
}

func applyClinicAccountingDocumentSectionOrder(fields map[string]any, input *UpdateClinicInput) error {
	if input.AccountingDocumentSectionOrder == nil {
		return nil
	}
	order := *input.AccountingDocumentSectionOrder
	allowedKeys := map[string]struct{}{
		"clinic_header": {}, "owner_pet_info": {}, "items_table": {},
		"payment_summary": {}, "footer_note": {},
	}
	seen := make(map[string]struct{}, len(order))
	for _, key := range order {
		if _, ok := allowedKeys[key]; !ok {
			return apperrors.WrapInvalidInput(fmt.Sprintf("unknown section key: %q", key))
		}
		if _, dup := seen[key]; dup {
			return apperrors.WrapInvalidInput(fmt.Sprintf("duplicate section key: %q", key))
		}
		seen[key] = struct{}{}
	}
	fields[colClinicAccountingDocumentSectionOrder] = pq.StringArray(order)
	return nil
}
