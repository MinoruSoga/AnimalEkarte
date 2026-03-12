import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Cage } from "@/types";

// Backend types (snake_case)
interface BackendCage {
  id: string;
  code: string;
  name: string;
  cage_type: string;
  cage_size: string;
  body_size?: string;
  billing_unit: string;
  price: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateCageRequest {
  code: string;
  name: string;
  cage_type: string;
  cage_size: string;
  body_size?: string;
  billing_unit: string;
  price: number;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateCageRequest {
  code?: string;
  name?: string;
  cage_type?: string;
  cage_size?: string;
  body_size?: string;
  billing_unit?: string;
  price?: number;
  is_active?: boolean;
  sort_order?: number;
}

function transformCage(data: BackendCage): Cage {
  return {
    id: data.id,
    code: data.code,
    name: data.name,
    cageType: data.cage_type as Cage["cageType"],
    cageSize: data.cage_size as Cage["cageSize"],
    bodySize: data.body_size as Cage["bodySize"],
    billingUnit: data.billing_unit as Cage["billingUnit"],
    price: data.price,
    isActive: data.is_active,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export const getAllCages = async (): Promise<Cage[]> => {
  const { data } = await axios.get<BackendCage[]>("/v1/masters/cages");
  return data.map(transformCage);
};

export const useGetAllCages = () => {
  return useQuery({
    queryKey: ["cages"],
    queryFn: getAllCages,
  });
};

export const getCageById = async (id: string): Promise<Cage> => {
  const { data } = await axios.get<BackendCage>(`/v1/masters/cages/${id}`);
  return transformCage(data);
};

export const useGetCageById = (id: string) => {
  return useQuery({
    queryKey: ["cages", id],
    queryFn: () => getCageById(id),
    enabled: !!id,
  });
};

export const createCage = async (req: CreateCageRequest): Promise<Cage> => {
  const { data } = await axios.post<BackendCage>("/v1/masters/cages", req);
  return transformCage(data);
};

export const useCreateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cages"] });
    },
  });
};

export const updateCage = async (id: string, req: UpdateCageRequest): Promise<Cage> => {
  const { data } = await axios.patch<BackendCage>(`/v1/masters/cages/${id}`, req);
  return transformCage(data);
};

export const useUpdateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateCageRequest }) =>
      updateCage(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cages"] });
    },
  });
};

export const deleteCage = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/cages/${id}`);
};

export const useDeleteCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["cages"] });
    },
  });
};
