import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { StaffMember } from "@/types";

// Backend types (snake_case)
interface BackendStaffMember {
  id: string;
  code: string;
  name: string;
  name_kana?: string;
  role: string;
  license_number?: string;
  email?: string;
  phone?: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateStaffMemberRequest {
  code: string;
  name: string;
  name_kana?: string;
  role: string;
  license_number?: string;
  email?: string;
  phone?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateStaffMemberRequest {
  code?: string;
  name?: string;
  name_kana?: string;
  role?: string;
  license_number?: string;
  email?: string;
  phone?: string;
  is_active?: boolean;
  sort_order?: number;
}

function transformStaffMember(data: BackendStaffMember): StaffMember {
  return {
    id: data.id,
    code: data.code,
    name: data.name,
    nameKana: data.name_kana,
    role: data.role as StaffMember["role"],
    licenseNumber: data.license_number,
    email: data.email,
    phone: data.phone,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export const getAllStaffMembers = async (): Promise<StaffMember[]> => {
  const { data } = await axios.get<BackendStaffMember[]>("/v1/staff-members");
  return data.map(transformStaffMember);
};

export const useGetAllStaffMembers = () => {
  return useQuery({
    queryKey: ["staffMembers"],
    queryFn: getAllStaffMembers,
  });
};

export const getStaffMemberById = async (id: string): Promise<StaffMember> => {
  const { data } = await axios.get<BackendStaffMember>(`/v1/staff-members/${id}`);
  return transformStaffMember(data);
};

export const useGetStaffMemberById = (id: string) => {
  return useQuery({
    queryKey: ["staffMembers", id],
    queryFn: () => getStaffMemberById(id),
    enabled: !!id,
  });
};

export const createStaffMember = async (req: CreateStaffMemberRequest): Promise<StaffMember> => {
  const { data } = await axios.post<BackendStaffMember>("/v1/staff-members", req);
  return transformStaffMember(data);
};

export const useCreateStaffMember = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createStaffMember,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffMembers"] });
    },
  });
};

export const updateStaffMember = async (id: string, req: UpdateStaffMemberRequest): Promise<StaffMember> => {
  const { data } = await axios.put<BackendStaffMember>(`/v1/staff-members/${id}`, req);
  return transformStaffMember(data);
};

export const useUpdateStaffMember = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateStaffMemberRequest }) =>
      updateStaffMember(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffMembers"] });
    },
  });
};

export const deleteStaffMember = async (id: string): Promise<void> => {
  await axios.delete(`/v1/staff-members/${id}`);
};

export const useDeleteStaffMember = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteStaffMember,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffMembers"] });
    },
  });
};
