import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Staff as ModelStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Strict role union (models.ts uses string, but we keep this for form safety)
// ─────────────────────────────────────────────────

export type StaffRoleValue =
  | "veterinarian"
  | "nurse"
  | "trimmer"
  | "reception"
  | "manager";

// ─────────────────────────────────────────────────
// Request types - api.yaml StaffRegistrationRequest 準拠
// ─────────────────────────────────────────────────

export interface CreateStaffRequest {
  name: string;
  staff_role: StaffRoleValue;
  email: string;
  password: string;
  license_number?: string;
  job_title_id?: string | null;
  sort_order?: number;
}

export interface UpdateStaffRequest {
  name?: string;
  staff_role?: StaffRoleValue;
  license_number?: string;
  is_active?: boolean;
  job_title_id?: string | null;
  sort_order?: number;
}

// ─────────────────────────────────────────────────
// Role label mapping
// ─────────────────────────────────────────────────

export const STAFF_ROLE_LABELS: Record<StaffRoleValue, string> = {
  veterinarian: "獣医師",
  nurse: "看護師",
  trimmer: "トリマー",
  reception: "受付",
  manager: "運営管理者",
};

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
    staffRole: data.staff_role as StaffRoleValue,
    jobTitleId: data.job_title_id ? String(data.job_title_id) : null,
    licenseNumber: data.license_number,
    sortOrder: data.sort_order,
    email,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
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
  const { data } = await axios.post<ModelStaff>("/v1/masters/staffs", req);
  return transformStaff(data);
}

export async function updateStaff(
  id: string,
  req: UpdateStaffRequest,
): Promise<Staff> {
  const { data } = await axios.patch<ModelStaff>(
    `/v1/masters/staffs/${id}`,
    req,
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
  });
}

export function useDeleteStaff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: STAFFS_QUERY_KEY });
    },
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
  });
}

// ─────────────────────────────────────────────────
// Clinics list (for staff assignment UI)
// ─────────────────────────────────────────────────

export interface ClinicSummary {
  id: string;
  name: string;
}

export function useGetClinicsList() {
  return useQuery({
    queryKey: ["clinics-list"],
    queryFn: async (): Promise<ClinicSummary[]> => {
      const { data } = await axios.get<Array<{ id: number; name: string }>>(
        "/v1/clinics",
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
  });
}
