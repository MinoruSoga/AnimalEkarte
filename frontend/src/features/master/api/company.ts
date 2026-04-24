import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Company as ModelCompany } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

export type UpdateCompanyRequest = Partial<
  Omit<ModelCompany, "id" | "created_at" | "updated_at">
>;

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

export type Company = ReturnType<typeof transformCompany>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const COMPANY_QUERY_KEY = ["masters", "company"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function getCompany(): Promise<Company> {
  const { data } = await axios.get<ModelCompany>("/v1/company");
  return transformCompany(data);
}

export async function updateCompany(req: UpdateCompanyRequest): Promise<Company> {
  const { data } = await axios.patch<ModelCompany>("/v1/company", req);
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

export function useUpdateCompany() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: updateCompany,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: COMPANY_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}
