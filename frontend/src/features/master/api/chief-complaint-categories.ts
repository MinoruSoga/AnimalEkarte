import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ChiefComplaintCategory as ModelChiefComplaintCategory } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types
// ─────────────────────────────────────────────────

export interface CreateChiefComplaintCategoryRequest {
  name: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateChiefComplaintCategoryRequest {
  name?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformChiefComplaintCategory(data: ModelChiefComplaintCategory) {
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

export type ChiefComplaintCategory = ReturnType<typeof transformChiefComplaintCategory>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const QUERY_KEY = ["masters", "chief-complaint-categories"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listChiefComplaintCategories(): Promise<ChiefComplaintCategory[]> {
  const { data } = await axios.get<ModelChiefComplaintCategory[]>(
    "/v1/masters/chief-complaint-categories"
  );
  return data.map(transformChiefComplaintCategory);
}

export async function createChiefComplaintCategory(
  req: CreateChiefComplaintCategoryRequest
): Promise<ChiefComplaintCategory> {
  const { data } = await axios.post<ModelChiefComplaintCategory>(
    "/v1/masters/chief-complaint-categories",
    req
  );
  return transformChiefComplaintCategory(data);
}

export async function updateChiefComplaintCategory(
  id: string,
  req: UpdateChiefComplaintCategoryRequest
): Promise<ChiefComplaintCategory> {
  const { data } = await axios.patch<ModelChiefComplaintCategory>(
    `/v1/masters/chief-complaint-categories/${id}`,
    req
  );
  return transformChiefComplaintCategory(data);
}

export async function deleteChiefComplaintCategory(id: string): Promise<void> {
  await axios.delete(`/v1/masters/chief-complaint-categories/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetChiefComplaintCategories() {
  return useQuery({
    queryKey: QUERY_KEY,
    queryFn: listChiefComplaintCategories,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateChiefComplaintCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createChiefComplaintCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}

export function useUpdateChiefComplaintCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateChiefComplaintCategoryRequest }) =>
      updateChiefComplaintCategory(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}

export function useDeleteChiefComplaintCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteChiefComplaintCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: QUERY_KEY });
    },
  });
}
