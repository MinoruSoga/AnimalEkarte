import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import type { Staff as ModelStaff } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Request types (derived from models.ts, extended for api.yaml StaffRegistrationRequest)
// ─────────────────────────────────────────────────

// ModelStaff fields usable in Create/Update (excludes system + relations + occupation_id which is overridden below as string|null)
type StaffModelFields = Omit<
  ModelStaff,
  | "id"
  | "clinic_id"
  | "account_id"
  | "account"
  | "occupation"
  | "clinic_assignments"
  | "occupation_id"
  | "created_at"
  | "updated_at"
>;

export type CreateStaffRequest = Pick<StaffModelFields, "name"> &
  Partial<Omit<StaffModelFields, "name">> & {
    email: string;
    password: string;
    occupation_id?: string | null;
  };

export type UpdateStaffRequest = Partial<StaffModelFields> & {
  password?: string;
  occupation_id?: string | null;
};

// ─────────────────────────────────────────────────
// Transform
// ─────────────────────────────────────────────────

/** Exported for unit tests — missing/unknown staff_type must not become "doctor". */
export function transformStaff(data: ModelStaff & { email?: string }) {
  // 複数所属は clinic_assignments を正とし、なければ staffs.clinic_id を主所属として使う。
  const mainClinicAssignment = data.clinic_assignments?.find((a) => a.is_main);
  const clinicId = mainClinicAssignment?.clinic_id ?? data.clinic_id ?? null;
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
    // LINE予約用フィールド — staff_type 欠損/非文字列は doctor に昇格させない
    staffType: typeof data.staff_type === "string" ? data.staff_type : "",
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

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

async function listStaffs(): Promise<Staff[]> {
  const { data } = await axios.get<ModelStaff[]>("/v1/masters/staffs");
  return data.map(transformStaff);
}

async function createStaff(req: CreateStaffRequest): Promise<Staff> {
  const payload = {
    ...req,
    occupation_id: req.occupation_id ? Number(req.occupation_id) : undefined,
  };
  const { data } = await axios.post<ModelStaff>("/v1/masters/staffs", payload);
  return transformStaff(data);
}

async function updateStaff(
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

async function deleteStaff(id: string): Promise<void> {
  await axios.delete(`/v1/masters/staffs/${id}`);
}

// ─────────────────────────────────────────────────
// TanStack Query hooks
// ─────────────────────────────────────────────────

export function useGetStaffs() {
  return useQuery({
    queryKey: queryKeys.masters.category("staffs"),
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
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("staffs") });
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
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("staffs") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
}

export function useDeleteStaff() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("staffs") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
}


// ─────────────────────────────────────────────────
// Re-exports: 呼び出し元の "../api/staffs" 単一 import 経路を維持するため、
// 分割後の権限グループ / 所属医院 / 予約タイプ制限 API をここから再公開する。
// ─────────────────────────────────────────────────

export {
  useGetAllStaffPermissionGroupMap,
  useGetStaffPermissionGroups,
  useUpdateStaffPermissionGroups,
} from "./staff-permission-groups";

export {
  useGetClinicsList,
  useGetStaffClinics,
  useUpdateStaffClinics,
} from "./staff-clinics";
export type { ClinicSummary } from "./staff-clinics";

export {
  useGetStaffCapableReservationTypes,
  useUpdateStaffCapableReservationTypes,
} from "./staff-reservation-types";
