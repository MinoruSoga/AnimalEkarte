import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Occupation as ModelOccupation } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type OccupationRequestBase = Omit<
  ModelOccupation,
  "id" | "clinic_id" | "created_at" | "updated_at"
>;

export type CreateOccupationRequest = Pick<OccupationRequestBase, "name"> &
  Partial<Omit<OccupationRequestBase, "name">>;

export type UpdateOccupationRequest = Partial<OccupationRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformOccupation(data: ModelOccupation) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    description: data.description,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type Occupation = ReturnType<typeof transformOccupation>;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

const getAllOccupations = async (): Promise<Occupation[]> => {
  const { data } = await axios.get<ModelOccupation[]>("/v1/masters/occupations");
  return data.map(transformOccupation);
};

const createOccupation = async (req: CreateOccupationRequest): Promise<Occupation> => {
  const { data } = await axios.post<ModelOccupation>("/v1/masters/occupations", req);
  return transformOccupation(data);
};

const updateOccupation = async (id: string, req: UpdateOccupationRequest): Promise<Occupation> => {
  const { data } = await axios.patch<ModelOccupation>(`/v1/masters/occupations/${id}`, req);
  return transformOccupation(data);
};

const deleteOccupation = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/occupations/${id}`);
};

// ─────────────────────────────────────────────────
// Query hooks
// ─────────────────────────────────────────────────

export const useGetAllOccupations = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("occupations"),
    queryFn: getAllOccupations,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const useCreateOccupation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createOccupation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("occupations") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const useUpdateOccupation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateOccupationRequest }) =>
      updateOccupation(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("occupations") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const useDeleteOccupation = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteOccupation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("occupations") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};
