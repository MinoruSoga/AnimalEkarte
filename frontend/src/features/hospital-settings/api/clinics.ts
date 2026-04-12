import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
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
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const CLINICS_QUERY_KEY = ["clinics"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listClinics(): Promise<TransformedClinic[]> {
  const { data } = await axios.get<BackendClinic[]>("/v1/clinics");
  return data.map(transformClinic);
}

export async function createClinic(
  req: CreateClinicRequest,
): Promise<TransformedClinic> {
  const { data } = await axios.post<BackendClinic>("/v1/clinics", req);
  return transformClinic(data);
}

export async function updateClinic(
  id: number,
  req: UpdateClinicRequest,
): Promise<TransformedClinic> {
  const { data } = await axios.patch<BackendClinic>(`/v1/clinics/${id}`, req);
  return transformClinic(data);
}

export async function deleteClinic(id: number): Promise<void> {
  await axios.delete(`/v1/clinics/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetClinics() {
  return useQuery({
    queryKey: CLINICS_QUERY_KEY,
    queryFn: listClinics,
  });
}

export function useCreateClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
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
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "クリニック更新"),
  });
}

export function useDeleteClinic() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteClinic,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CLINICS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "クリニック削除"),
  });
}
