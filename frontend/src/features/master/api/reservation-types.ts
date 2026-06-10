import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ReservationType as ModelReservationType } from "@/types/generated/models";
import type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
  ReorderReservationTypeRequest,
} from "@/types/reservation-type";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformReservationType(data: ModelReservationType) {
  const groupId = data.group_id ?? data.group?.id;

  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    groupId: groupId ? String(groupId) : undefined,
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
    category: data.category ?? "general",
  };
}

// Frontend domain type (ReturnType から自動導出)
export type ReservationType = ReturnType<typeof transformReservationType>;

// Re-export request types
export type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
  ReorderReservationTypeRequest,
} from "@/types/reservation-type";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const SERVICE_TYPES_QUERY_KEY = ["masters", "reservation-types"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listReservationTypes(): Promise<ReservationType[]> {
  const { data } = await axios.get<ModelReservationType[]>("/v1/masters/reservation-types");
  return data.map(transformReservationType);
}

export async function createReservationType(
  req: CreateReservationTypeRequest,
): Promise<ReservationType> {
  const { data } = await axios.post<ModelReservationType>("/v1/masters/reservation-types", req);
  return transformReservationType(data);
}

export async function updateReservationType(
  id: string,
  req: UpdateReservationTypeRequest,
): Promise<ReservationType> {
  const { data } = await axios.patch<ModelReservationType>(
    `/v1/masters/reservation-types/${id}`,
    req,
  );
  return transformReservationType(data);
}

export async function deleteReservationType(id: string): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${id}`);
}

export async function reorderReservationTypes(
  req: ReorderReservationTypeRequest,
): Promise<void> {
  await axios.patch("/v1/masters/reservation-types/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationTypes() {
  return useQuery({
    queryKey: SERVICE_TYPES_QUERY_KEY,
    queryFn: listReservationTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateReservationType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createReservationType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateReservationType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateReservationTypeRequest }) =>
      updateReservationType(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteReservationType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteReservationType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderReservationTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderReservationTypes,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
