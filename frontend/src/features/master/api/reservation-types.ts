import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { ReservationType as ModelReservationType } from "@/types/generated/models";
import type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
  ReorderReservationTypeRequest,
} from "@/types/reservation-type";

// ─────────────────────────────────────────────────
// Raw type (make codegen 実行後は ModelReservationType に parent_id/children が追加されれば削除可)
// ─────────────────────────────────────────────────

interface ReservationTypeRaw extends Omit<ModelReservationType, "parent" | "children"> {
  parent_id?: number;
  parent?: { id: number; name: string };
  children?: ReservationTypeRaw[];
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformReservationType(
  data: ReservationTypeRaw,
  parentData?: { id: string; name: string },
) {
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
    // 階層ツリー用フィールド
    parentId: parentData?.id,
    parentName: parentData?.name,
    isLeaf: (data.children?.length ?? 0) === 0,
    depth: parentData ? 1 : 0,
    childIds: data.children?.map((c) => String(c.id)) ?? [],
  };
}

// Frontend domain type (ReturnType から自動導出)
export type ReservationType = ReturnType<typeof transformReservationType>;

// Re-export request types
export type {
  CreateReservationTypeRequest,
  UpdateReservationTypeRequest,
} from "@/types/reservation-type";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listReservationTypes(): Promise<ReservationType[]> {
  const { data } = await axios.get<ReservationTypeRaw[]>("/v1/masters/reservation-types");
  // BE は root-only tree response を返す（parent_id IS NULL のルートのみ、children 埋め込み）
  const result: ReservationType[] = [];
  for (const root of data) {
    result.push(transformReservationType(root, undefined));
    for (const child of root.children ?? []) {
      result.push(transformReservationType(child, { id: String(root.id), name: root.name }));
    }
  }
  return result;
}

async function createReservationType(req: CreateReservationTypeRequest): Promise<ReservationType> {
  const { data } = await axios.post<ModelReservationType>("/v1/masters/reservation-types", req);
  return transformReservationType(data);
}

async function updateReservationType(
  id: string,
  req: UpdateReservationTypeRequest,
): Promise<ReservationType> {
  const { data } = await axios.patch<ModelReservationType>(
    `/v1/masters/reservation-types/${id}`,
    req,
  );
  return transformReservationType(data);
}

async function deleteReservationType(id: string): Promise<void> {
  await axios.delete(`/v1/masters/reservation-types/${id}`);
}

async function reorderReservationTypes(req: ReorderReservationTypeRequest): Promise<void> {
  await axios.patch("/v1/masters/reservation-types/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetReservationTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("reservation-types"),
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
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("reservation-types") });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("reservation-types") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteReservationType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteReservationType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("reservation-types") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

export function useReorderReservationTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderReservationTypes,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("reservation-types") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
}
