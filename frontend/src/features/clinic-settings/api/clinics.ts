import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformClinic } from "./transforms";

// Types
import type { Clinic as BackendClinic } from "@/types/generated/models";
import type { TransformedClinic } from "./transforms";

// Re-export for consumers
export type { TransformedClinic as Clinic };

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateClinicRequest {
  name: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
}

export interface UpdateClinicRequest {
  name?: string;
  postal_code?: string;
  address?: string;
  phone_number?: string;
  fax_number?: string;
  registration_number?: string;
  director_name?: string;
  email?: string;
  website?: string;
  is_active?: boolean;
  standard_tax_rate?: number;
  reduced_tax_rate?: number;
  accounting_document_show_logo?: boolean;
  accounting_document_show_registration_warning?: boolean;
  accounting_document_show_item_category?: boolean;
  accounting_document_footer_note?: string;
  // #190: セクション表示/非表示トグルと表示順
  accounting_document_show_clinic_header?: boolean;
  accounting_document_show_owner_pet_info?: boolean;
  accounting_document_show_items_table?: boolean;
  accounting_document_show_payment_summary?: boolean;
  accounting_document_section_order?: string[];
}

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function listClinics(): Promise<TransformedClinic[]> {
  // BUG-378 / BUG-038: 医院マスタでは割当外も含む全件が必要のため scope=all。
  // 認可は hospital-settings.view（system_admin 不要）。失敗時は throw して silent empty を避ける。
  const { data } = await axios.get<BackendClinic[]>("/v1/clinics", {
    params: { scope: "all" },
  });
  return (data ?? []).map(transformClinic);
}

async function createClinic(
  req: CreateClinicRequest,
): Promise<TransformedClinic> {
  const { data } = await axios.post<BackendClinic>("/v1/clinics", req);
  return transformClinic(data);
}

async function updateClinic(
  id: number,
  req: UpdateClinicRequest,
): Promise<TransformedClinic> {
  const { data } = await axios.patch<BackendClinic>(`/v1/clinics/${id}`, req);
  return transformClinic(data);
}

async function deleteClinic(id: number): Promise<void> {
  await axios.delete(`/v1/clinics/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetClinics() {
  return useQuery({
    queryKey: queryKeys.clinics.all(),
    queryFn: listClinics,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clinics.all() });
    },
    onError: (error) => handleApiError(error, "クリニック作成"),
  });
}

export function useUpdateClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: UpdateClinicRequest }) =>
      updateClinic(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clinics.all() });
    },
    onError: (error) => handleApiError(error, "クリニック更新"),
  });
}

export function useDeleteClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.clinics.all() });
    },
    onError: (error) => handleApiError(error, "クリニック削除"),
  });
}
