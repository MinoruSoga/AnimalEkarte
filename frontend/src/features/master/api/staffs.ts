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

function transformStaff(data: ModelStaff) {
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id ?? 0),
    name: data.name,
    isActive: data.is_active,
    staffRole: data.staff_role as StaffRoleValue,
    jobTitleId: data.job_title_id ? String(data.job_title_id) : null,
    licenseNumber: data.license_number,
    sortOrder: data.sort_order,
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
