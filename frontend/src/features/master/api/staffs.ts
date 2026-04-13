import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Staff as ModelStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types - api.yaml StaffRegistrationRequest 準拠
// ─────────────────────────────────────────────────

export interface CreateStaffRequest {
  name: string;
  email: string;
  password: string;
  license_number?: string;
  occupation_id?: string | null;
  sort_order?: number;
  // LINE予約用
  staff_type?: string;
  reservation_display_name?: string;
  reservation_visible?: boolean;
  reservation_comment?: string;
  reservation_image_url?: string;
}

export interface UpdateStaffRequest {
  name?: string;
  license_number?: string;
  is_active?: boolean;
  occupation_id?: string | null;
  sort_order?: number;
  password?: string;
  // LINE予約用
  staff_type?: string;
  reservation_display_name?: string;
  reservation_visible?: boolean;
  reservation_comment?: string;
  reservation_image_url?: string;
}

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

function transformStaff(data: ModelStaff & { email?: string }) {
  // Staff 型には clinic_id は存在しない（clinic_assignments で管理される）
  // メインクリニック ID を clinic_assignments から取得
  const mainClinicAssignment = data.clinic_assignments?.find((a) => a.is_main);
  const clinicId = mainClinicAssignment?.clinic_id ?? null;
  // API が直接 email フィールドを返す（Account Preload された場合）
  const email = data.email ?? data.account?.email ?? "";

  return {
    id: String(data.id ?? 0),
    clinicId: clinicId ? String(clinicId) : null,
    name: data.name,
    isActive: data.is_active,
    occupationId: data.occupation_id ? String(data.occupation_id) : null,
    occupationName: data.occupation?.name ?? null,
    licenseNumber: data.license_number,
    sortOrder: data.sort_order,
    email,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    // LINE予約用フィールド
    staffType: data.staff_type ?? "doctor",
    reservationDisplayName: data.reservation_display_name ?? "",
    reservationVisible: data.reservation_visible ?? true,
    reservationComment: data.reservation_comment ?? "",
    reservationImageUrl: data.reservation_image_url ?? "",
  };
}

export type Staff = ReturnType<typeof transformStaff>;

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const STAFFS_QUERY_KEY = ["masters", "staffs"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listStaffs(): Promise<Staff[]> {
  const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
  return data.map(transformStaff);
}

export async function createStaff(req: CreateStaffRequest): Promise<Staff> {
  const payload = {
    ...req,
    occupation_id: req.occupation_id ? Number(req.occupation_id) : undefined,
  };
  const { data } = await axios.post<ModelStaff>("/v1/masters/staffs", payload);
  return transformStaff(data);
}

export async function updateStaff(
  id: string,
  req: UpdateStaffRequest,
): Promise<Staff> {
  const payload = {
    ...req,
    occupation_id: req.occupation_id ? Number(req.occupation_id) : undefined,
  };
  const { data } = await axios.patch<ModelStaff>(
    `/v1/masters/staffs/${id}`,
    payload,
  );
  return transformStaff(data);
}

export async function deleteStaff(id: string): Promise<void> {
  await axios.delete(`/v1/masters/staffs/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetStaffs() {
  return useQuery({
    queryKey: STAFFS_QUERY_KEY,
    queryFn: listStaffs,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useCreateStaff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFFS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
}

export function useUpdateStaff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateStaffRequest }) =>
      updateStaff(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFFS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteStaff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFFS_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}

// ─────────────────────────────────────────────────
// Staff Permission Groups API
// ─────────────────────────────────────────────────

const STAFF_PERM_GROUPS_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "permission-groups"] as const;

/**
 * 全スタッフの権限グループIDマップを一括取得する。
 * staffId → groupId[] の Map を返す。
 */
export function useGetAllStaffPermissionGroupMap(staffIds: string[]) {
  return useQuery({
    queryKey: [...STAFFS_QUERY_KEY, "all-permission-group-map", ...staffIds],
    queryFn: async (): Promise<Map<string, string[]>> => {
      const map = new Map<string, string[]>();
      await Promise.all(
        staffIds.map(async (id) => {
          try {
            const { data } = await axios.get<{ group_ids: number[] }>(
              `/v1/masters/staffs/${id}/permission-groups`,
            );
            map.set(id, (data.group_ids ?? []).map(String));
          } catch {
            // バッチ取得: 個別スタッフの失敗（404含む）はスキップして継続
            map.set(id, []);
          }
        }),
      );
      return map;
    },
    enabled: staffIds.length > 0,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useGetStaffPermissionGroups(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_PERM_GROUPS_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ group_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/permission-groups`,
      );
      return (data.group_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useSetStaffPermissionGroups() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      groupIds,
    }: {
      staffId: string;
      groupIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/permission-groups`, {
        group_ids: groupIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_PERM_GROUPS_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}

// ─────────────────────────────────────────────────
// Clinics list (for staff assignment UI)
// ─────────────────────────────────────────────────

export interface ClinicSummary {
  id: string;
  name: string;
}

export function useGetClinicsList(scope?: "all") {
  return useQuery({
    queryKey: ["clinics-list", scope ?? "assigned"],
    queryFn: async (): Promise<ClinicSummary[]> => {
      const params = scope ? { scope } : undefined;
      const { data } = await axios.get<Array<{ id: number; name: string }>>(
        "/v1/clinics",
        { params },
      );
      return data.map((c) => ({ id: String(c.id), name: c.name }));
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

// ─────────────────────────────────────────────────
// Staff Clinic Assignments API
// ─────────────────────────────────────────────────

const STAFF_CLINICS_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "clinics"] as const;

export function useGetStaffClinics(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_CLINICS_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ clinic_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/clinics`,
      );
      return (data.clinic_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useSetStaffClinics() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      clinicIds,
    }: {
      staffId: string;
      clinicIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/clinics`, {
        clinic_ids: clinicIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_CLINICS_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}

// ─────────────────────────────────────────────────
// Staff Excluded Service Types API
// ─────────────────────────────────────────────────

const STAFF_EXCLUDED_ST_KEY = (staffId: string) =>
  [...STAFFS_QUERY_KEY, staffId, "excluded-reservation-types"] as const;

export function useGetStaffExcludedReservationTypes(staffId: string | null) {
  return useQuery({
    queryKey: STAFF_EXCLUDED_ST_KEY(staffId ?? ""),
    queryFn: async (): Promise<string[]> => {
      const { data } = await axios.get<{ reservation_type_ids: number[] }>(
        `/v1/masters/staffs/${staffId}/excluded-reservation-types`,
      );
      return (data.reservation_type_ids ?? []).map(String);
    },
    enabled: staffId !== null,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

export function useSetStaffExcludedReservationTypes() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      staffId,
      reservationTypeIds,
    }: {
      staffId: string;
      reservationTypeIds: string[];
    }) => {
      await axios.put(`/v1/masters/staffs/${staffId}/excluded-reservation-types`, {
        reservation_type_ids: reservationTypeIds.map((id) => parseInt(id, 10)),
      });
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: STAFF_EXCLUDED_ST_KEY(variables.staffId),
      });
    },
    onError: (error) => handleApiError(error, "設定"),
  });
}
