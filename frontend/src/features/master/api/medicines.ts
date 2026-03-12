import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Medicine } from "@/types";

// Backend types (snake_case)
interface BackendMedicine {
  id: string;
  code: string;
  name: string;
  dosage_form: string;
  medicine_unit: string;
  price: number;
  default_quantity?: number;
  inventory_id?: string;
  description: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateMedicineRequest {
  code: string;
  name: string;
  dosage_form: string;
  medicine_unit: string;
  price: number;
  default_quantity?: number;
  inventory_id?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateMedicineRequest {
  code?: string;
  name?: string;
  dosage_form?: string;
  medicine_unit?: string;
  price?: number;
  default_quantity?: number;
  inventory_id?: string;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

function transformMedicine(data: BackendMedicine): Medicine {
  return {
    id: data.id,
    code: data.code,
    name: data.name,
    dosageForm: data.dosage_form as Medicine["dosageForm"],
    medicineUnit: data.medicine_unit as Medicine["medicineUnit"],
    price: data.price,
    defaultQuantity: data.default_quantity,
    inventoryId: data.inventory_id,
    description: data.description,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export const getAllMedicines = async (): Promise<Medicine[]> => {
  const { data } = await axios.get<BackendMedicine[]>("/v1/masters/medicines");
  return data.map(transformMedicine);
};

export const useGetAllMedicines = () => {
  return useQuery({
    queryKey: ["medicines"],
    queryFn: getAllMedicines,
  });
};

export const getMedicineById = async (id: string): Promise<Medicine> => {
  const { data } = await axios.get<BackendMedicine>(`/v1/masters/medicines/${id}`);
  return transformMedicine(data);
};

export const useGetMedicineById = (id: string) => {
  return useQuery({
    queryKey: ["medicines", id],
    queryFn: () => getMedicineById(id),
    enabled: !!id,
  });
};

export const createMedicine = async (req: CreateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.post<BackendMedicine>("/v1/masters/medicines", req);
  return transformMedicine(data);
};

export const useCreateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createMedicine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medicines"] });
    },
  });
};

export const updateMedicine = async (id: string, req: UpdateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.patch<BackendMedicine>(`/v1/masters/medicines/${id}`, req);
  return transformMedicine(data);
};

export const useUpdateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMedicineRequest }) =>
      updateMedicine(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medicines"] });
    },
  });
};

export const deleteMedicine = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/medicines/${id}`);
};

export const useDeleteMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteMedicine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medicines"] });
    },
  });
};
