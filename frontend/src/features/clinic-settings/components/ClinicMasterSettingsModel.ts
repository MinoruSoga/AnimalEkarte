import type { Clinic } from "../api/clinics";

export interface ClinicFormData {
  name: string;
  postal_code: string;
  address: string;
  phone_number: string;
  fax_number: string;
  registration_number: string;
  director_name: string;
  email: string;
  website: string;
  is_active: boolean;
  standard_tax_rate: number;
  reduced_tax_rate: number;
  accounting_document_show_logo: boolean;
  accounting_document_show_registration_warning: boolean;
  accounting_document_show_item_category: boolean;
  accounting_document_footer_note: string;
}

export const DEFAULT_CLINIC_FORM_DATA: ClinicFormData = {
  name: "",
  postal_code: "",
  address: "",
  phone_number: "",
  fax_number: "",
  registration_number: "",
  director_name: "",
  email: "",
  website: "",
  is_active: true,
  standard_tax_rate: 0.1,
  reduced_tax_rate: 0.08,
  accounting_document_show_logo: false,
  accounting_document_show_registration_warning: true,
  accounting_document_show_item_category: true,
  accounting_document_footer_note: "",
};

export function clinicToFormData(item: Clinic): ClinicFormData {
  return {
    name: item.name,
    postal_code: item.postalCode,
    address: item.address,
    phone_number: item.phoneNumber,
    fax_number: item.faxNumber,
    registration_number: item.registrationNumber,
    director_name: item.directorName,
    email: item.email,
    website: item.website,
    is_active: item.isActive,
    standard_tax_rate: item.standardTaxRate ?? 0.1,
    reduced_tax_rate: item.reducedTaxRate ?? 0.08,
    accounting_document_show_logo: item.accountingDocumentShowLogo,
    accounting_document_show_registration_warning: item.accountingDocumentShowRegistrationWarning,
    accounting_document_show_item_category: item.accountingDocumentShowItemCategory,
    accounting_document_footer_note: item.accountingDocumentFooterNote,
  };
}
