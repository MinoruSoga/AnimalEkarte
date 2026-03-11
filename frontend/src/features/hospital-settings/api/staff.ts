import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { BackendStaff, CreateStaffRequest, UpdateStaffRequest } from "./types";

// staff は master_items の category="staff" として管理される

export const getStaffs = async (): Promise<BackendStaff[]> => {
  const { data } = await axios.get<BackendStaff[]>("/v1/master-items/category/staff");
  return data;
};

export const useGetStaffs = () => {
  return useQuery({
    queryKey: ["staffs"],
    queryFn: getStaffs,
  });
};

export const getStaff = async (id: string): Promise<BackendStaff> => {
  const { data } = await axios.get<BackendStaff>(`/v1/master-items/${id}`);
  return data;
};

export const useGetStaff = (id: string) => {
  return useQuery({
    queryKey: ["staff", id],
    queryFn: () => getStaff(id),
    enabled: !!id,
  });
};

// clinicId によるフィルタはフロントエンド側で行う（/v1/master-items/category/staff は全スタッフを返す）
export const getStaffsByClinicId = async (_clinicId: string): Promise<BackendStaff[]> => {
  const { data } = await axios.get<BackendStaff[]>("/v1/master-items/category/staff");
  return data;
};

export const useGetStaffsByClinicId = (clinicId: string) => {
  return useQuery({
    queryKey: ["staffs", "clinic", clinicId],
    queryFn: () => getStaffsByClinicId(clinicId),
    enabled: !!clinicId,
  });
};

export const createStaff = async (req: CreateStaffRequest): Promise<BackendStaff> => {
  const { data } = await axios.post<BackendStaff>("/v1/master-items", {
    ...req,
    category: "staff",
  });
  return data;
};

export const useCreateStaff = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffs"] });
    },
  });
};

export const updateStaff = async (id: string, req: UpdateStaffRequest): Promise<BackendStaff> => {
  const { data } = await axios.put<BackendStaff>(`/v1/master-items/${id}`, req);
  return data;
};

export const useUpdateStaff = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateStaffRequest }) => updateStaff(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffs"] });
    },
  });
};

export const deleteStaff = async (id: string): Promise<void> => {
  await axios.delete(`/v1/master-items/${id}`);
};

export const useDeleteStaff = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteStaff,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["staffs"] });
    },
  });
};
