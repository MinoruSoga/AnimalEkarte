import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationCategoryGroup as ModelReservationCategoryGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformReservationCategoryGroup(data: ModelReservationCategoryGroup) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    color: data.color,
    sortOrder: data.sort_order,
    isActive: data.is_active,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type ReservationCategoryGroup = ReturnType<typeof transformReservationCategoryGroup>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const RESERVATION_CATEGORY_GROUPS_QUERY_KEY = ["masters", "reservation-category-groups"] as const;

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

export type CreateReservationCategoryGroupRequest = Pick<ModelReservationCategoryGroup, "name"> & {
  color?: string;
  sort_order?: number;
  is_active?: boolean;
};

export type UpdateReservationCategoryGroupRequest = Partial<
  Pick<ModelReservationCategoryGroup, "name" | "color" | "sort_order" | "is_active">
>;

export type ReorderReservationCategoryGroupRequest = {
  ids: number[];
};

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listReservationCategoryGroups(): Promise<ReservationCategoryGroup[]> {
  const { data } = await axios.get<ModelReservationCategoryGroup[]>(
    "/v1/masters/reservation-category-groups",
  );
  return data.map(transformReservationCategoryGroup);
}

export async function createReservationCategoryGroup(
  req: CreateReservationCategoryGroupRequest,
): Promise<ReservationCategoryGroup> {
  const { data } = await axios.post<ModelReservationCategoryGroup>(
    "/v1/masters/reservation-category-groups",
    req,
  );
  return transformReservationCategoryGroup(data);
}

export async function updateReservationCategoryGroup(
  id: string,
  req: UpdateReservationCategoryGroupRequest,
): Promise<ReservationCategoryGroup> {
  const { data } = await axios.patch<ModelReservationCategoryGroup>(
    `/v1/masters/reservation-category-groups/${id}`,
    req,
  );
  return transformReservationCategoryGroup(data);
}

export async function deleteReservationCategoryGroup(id: string): Promise<void> {
  await axios.delete(`/v1/masters/reservation-category-groups/${id}`);
}

export async function reorderReservationCategoryGroups(
  req: ReorderReservationCategoryGroupRequest,
): Promise<void> {
  await axios.patch("/v1/masters/reservation-category-groups/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationCategoryGroups() {
  return useQuery({
    queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY,
    queryFn: listReservationCategoryGroups,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateReservationCategoryGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createReservationCategoryGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
  });
}

export function useUpdateReservationCategoryGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationCategoryGroupRequest }) =>
      updateReservationCategoryGroup(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
  });
}

export function useDeleteReservationCategoryGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteReservationCategoryGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
  });
}

export function useReorderReservationCategoryGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderReservationCategoryGroups,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
  });
}
