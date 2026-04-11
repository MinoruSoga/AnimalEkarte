import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationCategory as ModelReservationCategory } from "@/types/generated/models";
import type {
  CreateReservationCategoryRequest,
  UpdateReservationCategoryRequest,
  ReorderReservationCategoryRequest,
} from "@/types/reservation-category";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformReservationCategory(data: ModelReservationCategory) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    groupId: data.group_id ? String(data.group_id) : undefined,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    // LINE予約用フィールド
    reservationDisplayName: data.reservation_display_name ?? "",
    durationMinutes: data.duration_minutes ?? 15,
    shortName: data.short_name ?? "",
    showShortName: data.show_short_name ?? false,
    reservationVisible: data.reservation_visible ?? true,
    reservationComment: data.reservation_comment ?? "",
    reservationImageUrl: data.reservation_image_url ?? "",
    reservationDayOption: data.reservation_day_option ?? "none",
    isInternal: data.is_internal ?? false,
  };
}

// Frontend domain type (ReturnType から自動導出)
export type ReservationCategory = ReturnType<typeof transformReservationCategory>;

// Re-export request types
export type {
  CreateReservationCategoryRequest,
  UpdateReservationCategoryRequest,
  ReorderReservationCategoryRequest,
} from "@/types/reservation-category";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const SERVICE_TYPES_QUERY_KEY = ["masters", "reservation-types"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listReservationCategories(): Promise<ReservationCategory[]> {
  const { data } = await axios.get<ModelReservationCategory[]>("/v1/masters/reservation-types");
  return data.map(transformReservationCategory);
}

export async function createReservationCategory(
  req: CreateReservationCategoryRequest,
): Promise<ReservationCategory> {
  const { data } = await axios.post<ModelReservationCategory>("/v1/masters/reservation-types", req);
  return transformReservationCategory(data);
}

export async function updateReservationCategory(
  id: string,
  req: UpdateReservationCategoryRequest,
): Promise<ReservationCategory> {
  const { data } = await axios.patch<ModelReservationCategory>(
    `/v1/masters/reservation-types/${id}`,
    req,
  );
  return transformReservationCategory(data);
}

export async function deleteReservationCategory(id: string): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${id}`);
}

export async function reorderReservationCategories(
  req: ReorderReservationCategoryRequest,
): Promise<void> {
  await axios.patch("/v1/masters/reservation-types/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationCategories() {
  return useQuery({
    queryKey: SERVICE_TYPES_QUERY_KEY,
    queryFn: listReservationCategories,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateReservationCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createReservationCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useUpdateReservationCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationCategoryRequest }) =>
      updateReservationCategory(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useDeleteReservationCategory() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteReservationCategory,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useReorderReservationCategories() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderReservationCategories,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}
