import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
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

interface UpdateCompanyRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  email?: string;
  website?: string;
  director_name?: string;
  registration_number?: string;
  invoice_registration_number?: string;
}

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function getCompany(): Promise<Company> {
  const { data } = await axios.get<ModelCompany>("/v1/company");
  return transformCompany(data);
}

async function updateCompany(req: UpdateCompanyRequest): Promise<Company> {
  const { data } = await axios.patch<ModelCompany>("/v1/company", req);
  return transformCompany(data);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

/**
 * FE-RC-015: 法人情報（会社マスタ）は master/accounting/clinic-settings の
 * 複数 feature から参照される。単一 feature 専有ではないため src/hooks へ配置
 * （features/master/api/company.ts から移設・2026-09-03）。
 */
export function useGetCompany() {
  return useQuery({
    queryKey: queryKeys.masters.category("company"),
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
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("company") });
    },
    onError: (error) => handleApiError(error, "法人情報更新"),
  });
}
