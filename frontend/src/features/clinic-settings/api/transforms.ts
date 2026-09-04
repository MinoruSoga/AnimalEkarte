import type { Clinic } from "@/types/generated/models";
import { DEFAULT_STANDARD_TAX_RATE, DEFAULT_REDUCED_TAX_RATE } from "@/constants/tax";

export function transformClinic(data: Clinic) {
  return {
    id: data.id,
    name: data.name,
    postalCode: data.postal_code,
    address: data.address,
    phoneNumber: data.phone_number,
    faxNumber: data.fax_number,
    registrationNumber: data.registration_number,
    directorName: data.director_name,
    email: data.email,
    website: data.website,
    logoUrl: data.logo_url,
    isActive: data.is_active,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    standardTaxRate: data.standard_tax_rate ?? DEFAULT_STANDARD_TAX_RATE,
    reducedTaxRate: data.reduced_tax_rate ?? DEFAULT_REDUCED_TAX_RATE,
    accountingDocumentShowLogo: data.accounting_document_show_logo ?? false,
    accountingDocumentShowRegistrationWarning:
      data.accounting_document_show_registration_warning ?? true,
    accountingDocumentShowItemCategory: data.accounting_document_show_item_category ?? true,
    accountingDocumentFooterNote: data.accounting_document_footer_note ?? "",
    // #190: セクション表示/非表示トグルと表示順
    accountingDocumentShowClinicHeader: data.accounting_document_show_clinic_header ?? true,
    accountingDocumentShowOwnerPetInfo: data.accounting_document_show_owner_pet_info ?? true,
    accountingDocumentShowItemsTable: data.accounting_document_show_items_table ?? true,
    accountingDocumentShowPaymentSummary: data.accounting_document_show_payment_summary ?? true,
    accountingDocumentSectionOrder: data.accounting_document_section_order ?? [],
  };
}

export type TransformedClinic = ReturnType<typeof transformClinic>;
