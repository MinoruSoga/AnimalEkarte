import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

// ─────────────────────────────────────────────────
// Backend types (snake_case) - api.yaml Staff schema 準拠
// ─────────────────────────────────────────────────

export type StaffRoleValue =
  | "veterinarian"
  | "nurse"
  | "trimmer"
  | "reception"
  | "manager";

export interface BackendStaff {
  id: string;
  clinic_id: string;
  code: string;
  name: string;
  is_active: boolean;
  staff_role: StaffRoleValue;
  job_title_id: string | null;
  license_number: string;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// ─────────────────────────────────────────────────
// Request types - api.yaml StaffRegistrationRequest 準拠
// ─────────────────────────────────────────────────

export interface CreateStaffRequest {
  name: string;
  staff_role: StaffRoleValue;
  email: string;
  password: string;
  code?: string;
  license_number?: string;
  job_title_id?: string | null;
  sort_order?: number;
}

export interface UpdateStaffRequest {
  name?: string;
  staff_role?: StaffRoleValue;
  code?: string;
  license_number?: string;
  is_active?: boolean;
  job_title_id?: string | null;
  sort_order?: number;
}

// ─────────────────────────────────────────────────
// Frontend display type
// ─────────────────────────────────────────────────

export interface Staff {
  id: string;
  clinicId: string;
  code: string;
  name: string;
  isActive: boolean;
  staffRole: StaffRoleValue;
  jobTitleId: string | null;
  licenseNumber: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
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

function transformStaff(data: BackendStaff): Staff {
  return {
    id: data.id,
    clinicId: data.clinic_id,
    code: data.code,
    name: data.name,
    isActive: data.is_active,
    staffRole: data.staff_role,
    jobTitleId: data.job_title_id,
    licenseNumber: data.license_number,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

// ─────────────────────────────────────────────────
// Query keys
// ─────────────────────────────────────────────────

const STAFFS_QUERY_KEY = ["masters", "staffs"] as const;

// ─────────────────────────────────────────────────
// API functions
// ─────────────────────────────────────────────────

export async function listStaffs(): Promise<Staff[]> {
  const { data } = await axios.get<BackendStaff[]>("/v1/masters/staffs");
  return data.map(transformStaff);
}

export async function createStaff(req: CreateStaffRequest): Promise<Staff> {
  const { data } = await axios.post<BackendStaff>("/v1/masters/staffs", req);
  return transformStaff(data);
}

export async function updateStaff(
  id: string,
  req: UpdateStaffRequest,
): Promise<Staff> {
  const { data } = await axios.patch<BackendStaff>(
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

export function useListStaffs() {
  return useQuery({
    queryKey: STAFFS_QUERY_KEY,
    queryFn: listStaffs,
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
