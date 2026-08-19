import { CircleDot } from "lucide-react";

import type { ActiveFilter, FilterProperty } from "@/components/shared/PropertyFilter/types";
import { DEFAULT_STANDARD_TAX_RATE, DEFAULT_REDUCED_TAX_RATE } from "@/constants/tax";
import { normalizedIncludes } from "@/lib/normalize-kana";

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
  // #190: セクション表示/非表示トグルと表示順
  accounting_document_show_clinic_header: boolean;
  accounting_document_show_owner_pet_info: boolean;
  accounting_document_show_items_table: boolean;
  accounting_document_show_payment_summary: boolean;
  accounting_document_section_order: string[];
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
  standard_tax_rate: DEFAULT_STANDARD_TAX_RATE,
  reduced_tax_rate: DEFAULT_REDUCED_TAX_RATE,
  accounting_document_show_logo: false,
  accounting_document_show_registration_warning: true,
  accounting_document_show_item_category: true,
  accounting_document_footer_note: "",
  accounting_document_show_clinic_header: true,
  accounting_document_show_owner_pet_info: true,
  accounting_document_show_items_table: true,
  accounting_document_show_payment_summary: true,
  accounting_document_section_order: [],
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
    standard_tax_rate: item.standardTaxRate ?? DEFAULT_STANDARD_TAX_RATE,
    reduced_tax_rate: item.reducedTaxRate ?? DEFAULT_REDUCED_TAX_RATE,
    accounting_document_show_logo: item.accountingDocumentShowLogo,
    accounting_document_show_registration_warning: item.accountingDocumentShowRegistrationWarning,
    accounting_document_show_item_category: item.accountingDocumentShowItemCategory,
    accounting_document_footer_note: item.accountingDocumentFooterNote,
    accounting_document_show_clinic_header: item.accountingDocumentShowClinicHeader,
    accounting_document_show_owner_pet_info: item.accountingDocumentShowOwnerPetInfo,
    accounting_document_show_items_table: item.accountingDocumentShowItemsTable,
    accounting_document_show_payment_summary: item.accountingDocumentShowPaymentSummary,
    accounting_document_section_order: item.accountingDocumentSectionOrder,
  };
}

export const CLINIC_STATUS_FILTER: FilterProperty = {
  key: "status",
  label: "ステータス",
  type: "select",
  icon: CircleDot,
  options: [
    { value: "active", label: "有効" },
    { value: "inactive", label: "無効" },
  ],
};

export function filterClinics(
  clinics: Clinic[],
  searchTerm: string,
  filters: ActiveFilter[],
): Clinic[] {
  const term = searchTerm.trim();
  return clinics.filter((clinic) => {
    if (
      term
      && !normalizedIncludes(clinic.name, term)
      && !normalizedIncludes(clinic.phoneNumber, term)
      && !normalizedIncludes(clinic.email, term)
    ) {
      return false;
    }
    for (const filter of filters) {
      if (filter.key !== "status" || typeof filter.value !== "string") {
        continue;
      }
      const wantActive = filter.value === "active";
      if (filter.condition === "is" && clinic.isActive !== wantActive) {
        return false;
      }
      if (filter.condition === "is_not" && clinic.isActive === wantActive) {
        return false;
      }
    }
    return true;
  });
}
