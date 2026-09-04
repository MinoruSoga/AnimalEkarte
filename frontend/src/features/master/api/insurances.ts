import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Insurance as ModelInsurance } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type InsuranceRequestBase = Omit<ModelInsurance, "id" | "clinic_id" | "created_at" | "updated_at">;

export type CreateInsuranceRequest = Pick<InsuranceRequestBase, "name"> &
  Partial<Omit<InsuranceRequestBase, "name">>;

export type UpdateInsuranceRequest = Partial<InsuranceRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformInsurance(data: ModelInsurance) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    isActive: data.is_active,
    description: data.description,
    coverageRate: data.coverage_rate,
    contactPhone: data.contact_phone,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type Insurance = ReturnType<typeof transformInsurance>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getAllInsurances = async (): Promise<Insurance[]> => {
  const { data } = await axios.get<ModelInsurance[]>("/v1/masters/insurances");
  return data.map(transformInsurance);
};

const createInsurance = async (req: CreateInsuranceRequest): Promise<Insurance> => {
  const { data } = await axios.post<ModelInsurance>("/v1/masters/insurances", req);
  return transformInsurance(data);
};

const updateInsurance = async (id: string, req: UpdateInsuranceRequest): Promise<Insurance> => {
  const { data } = await axios.patch<ModelInsurance>(`/v1/masters/insurances/${id}`, req);
  return transformInsurance(data);
};

const deleteInsurance = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/insurances/${id}`);
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetAllInsurances = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("insurances"),
    queryFn: getAllInsurances,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const useCreateInsurance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createInsurance,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("insurances") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateInsurance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateInsuranceRequest }) =>
      updateInsurance(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("insurances") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteInsurance = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteInsurance,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("insurances") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};
