import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Company as ModelCompany } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformCompany(data: ModelCompany) {
  return {
    id: data.id,
    name: data.name,
    postalCode: data.postal_code,
    address: data.address,
    phoneNumber: data.phone_number,
    faxNumber: data.fax_number,
    email: data.email,
    website: data.website,
    directorName: data.director_name,
    registrationNumber: data.registration_number,
    invoiceRegistrationNumber: data.invoice_registration_number,
    logoUrl: data.logo_url,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

type Company = ReturnType<typeof transformCompany>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const COMPANY_QUERY_KEY = ["masters", "company"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getCompany(): Promise<Company> {
  const { data } = await axios.get<ModelCompany>("/v1/company");
  return transformCompany(data);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetCompany() {
  return useQuery({
    queryKey: COMPANY_QUERY_KEY,
    queryFn: getCompany,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
