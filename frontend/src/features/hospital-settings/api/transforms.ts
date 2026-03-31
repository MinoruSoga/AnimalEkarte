import type { Clinic } from "@/types/generated/models";

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
    standardTaxRate: data.standard_tax_rate ?? 0.1,
    reducedTaxRate: data.reduced_tax_rate ?? 0.08,
  };
}

export type TransformedClinic = ReturnType<typeof transformClinic>;
