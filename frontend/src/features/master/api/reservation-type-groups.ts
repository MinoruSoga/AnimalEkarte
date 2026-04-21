import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationTypeGroup as ModelReservationTypeGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformReservationTypeGroup(data: ModelReservationTypeGroup) {
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

export type ReservationTypeGroup = ReturnType<typeof transformReservationTypeGroup>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const RESERVATION_CATEGORY_GROUPS_QUERY_KEY = ["masters", "reservation-type-groups"] as const;

// ─────────────────────────────────────────────────
// Request types (derived from models.ts)
// ─────────────────────────────────────────────────

export type CreateReservationTypeGroupRequest = Pick<ModelReservationTypeGroup, "name"> & {
  color?: string;
  sort_order?: number;
  is_active?: boolean;
};

export type UpdateReservationTypeGroupRequest = Partial<
  Pick<ModelReservationTypeGroup, "name" | "color" | "sort_order" | "is_active">
>;

export type ReorderReservationTypeGroupRequest = {
  ids: number[];
};

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listReservationTypeGroups(): Promise<ReservationTypeGroup[]> {
  const { data } = await axios.get<ModelReservationTypeGroup[]>(
    "/v1/masters/reservation-type-groups",
  );
  return data.map(transformReservationTypeGroup);
}

export async function createReservationTypeGroup(
  req: CreateReservationTypeGroupRequest,
): Promise<ReservationTypeGroup> {
  const { data } = await axios.post<ModelReservationTypeGroup>(
    "/v1/masters/reservation-type-groups",
    req,
  );
  return transformReservationTypeGroup(data);
}

export async function updateReservationTypeGroup(
  id: string,
  req: UpdateReservationTypeGroupRequest,
): Promise<ReservationTypeGroup> {
  const { data } = await axios.patch<ModelReservationTypeGroup>(
    `/v1/masters/reservation-type-groups/${id}`,
    req,
  );
  return transformReservationTypeGroup(data);
}

export async function deleteReservationTypeGroup(id: string): Promise<void> {
  await axios.delete(`/v1/masters/reservation-type-groups/${id}`);
}

export async function reorderReservationTypeGroups(
  req: ReorderReservationTypeGroupRequest,
): Promise<void> {
  await axios.patch("/v1/masters/reservation-type-groups/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationTypeGroups() {
  return useQuery({
    queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY,
    queryFn: listReservationTypeGroups,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateReservationTypeGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createReservationTypeGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateReservationTypeGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationTypeGroupRequest }) =>
      updateReservationTypeGroup(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteReservationTypeGroup() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteReservationTypeGroup,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderReservationTypeGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (ids: number[]) => reorderReservationTypeGroups({ ids }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: RESERVATION_CATEGORY_GROUPS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
