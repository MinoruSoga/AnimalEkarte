import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ChiefComplaintType as ModelChiefComplaintType } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

type ChiefComplaintTypeRequestBase = Omit<
  ModelChiefComplaintType,
  "id" | "clinic_id" | "created_at" | "updated_at"
>;

export type CreateChiefComplaintTypeRequest = Pick<
  ChiefComplaintTypeRequestBase,
  "name" | "is_active"
> &
  Partial<Omit<ChiefComplaintTypeRequestBase, "name" | "is_active">>;

export type UpdateChiefComplaintTypeRequest = Partial<ChiefComplaintTypeRequestBase>;

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformChiefComplaintType(data: ModelChiefComplaintType) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    description: data.description,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type ChiefComplaintType = ReturnType<typeof transformChiefComplaintType>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function listChiefComplaintTypes(): Promise<ChiefComplaintType[]> {
  const { data } = await axios.get<ModelChiefComplaintType[]>(
    "/v1/masters/chief-complaint-types"
  );
  return data.map(transformChiefComplaintType);
}

async function createChiefComplaintType(
  req: CreateChiefComplaintTypeRequest
): Promise<ChiefComplaintType> {
  const { data } = await axios.post<ModelChiefComplaintType>(
    "/v1/masters/chief-complaint-types",
    req
  );
  return transformChiefComplaintType(data);
}

async function updateChiefComplaintType(
  id: string,
  req: UpdateChiefComplaintTypeRequest
): Promise<ChiefComplaintType> {
  const { data } = await axios.patch<ModelChiefComplaintType>(
    `/v1/masters/chief-complaint-types/${id}`,
    req
  );
  return transformChiefComplaintType(data);
}

async function deleteChiefComplaintType(id: string): Promise<void> {
  await axios.delete(`/v1/masters/chief-complaint-types/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetChiefComplaintTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("chief-complaint-types"),
    queryFn: listChiefComplaintTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateChiefComplaintType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createChiefComplaintType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("chief-complaint-types") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateChiefComplaintType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateChiefComplaintTypeRequest }) =>
      updateChiefComplaintType(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("chief-complaint-types") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteChiefComplaintType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteChiefComplaintType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("chief-complaint-types") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

