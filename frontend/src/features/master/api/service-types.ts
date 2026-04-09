import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ServiceType as ModelServiceType } from "@/types/generated/models";
import type {
  CreateServiceTypeRequest,
  UpdateServiceTypeRequest,
  ReorderServiceTypeRequest,
} from "@/types/service-type";

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformServiceType(data: ModelServiceType) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
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
export type ServiceType = ReturnType<typeof transformServiceType>;

// Re-export request types
export type {
  CreateServiceTypeRequest,
  UpdateServiceTypeRequest,
  ReorderServiceTypeRequest,
} from "@/types/service-type";

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

export const SERVICE_TYPES_QUERY_KEY = ["masters", "service-types"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listServiceTypes(): Promise<ServiceType[]> {
  const { data } = await axios.get<ModelServiceType[]>("/v1/masters/service-types");
  return data.map(transformServiceType);
}

export async function createServiceType(
  req: CreateServiceTypeRequest,
): Promise<ServiceType> {
  const { data } = await axios.post<ModelServiceType>("/v1/masters/service-types", req);
  return transformServiceType(data);
}

export async function updateServiceType(
  id: string,
  req: UpdateServiceTypeRequest,
): Promise<ServiceType> {
  const { data } = await axios.patch<ModelServiceType>(
    `/v1/masters/service-types/${id}`,
    req,
  );
  return transformServiceType(data);
}

export async function deleteServiceType(id: string): Promise<void> {
  await axios.delete(`/v1/masters/service-types/${id}`);
}

export async function reorderServiceTypes(
  req: ReorderServiceTypeRequest,
): Promise<void> {
  await axios.patch("/v1/masters/service-types/reorder", req);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetServiceTypes() {
  return useQuery({
    queryKey: SERVICE_TYPES_QUERY_KEY,
    queryFn: listServiceTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateServiceType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createServiceType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useUpdateServiceType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateServiceTypeRequest }) =>
      updateServiceType(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useDeleteServiceType() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteServiceType,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}

export function useReorderServiceTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderServiceTypes,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: SERVICE_TYPES_QUERY_KEY });
    },
  });
}
