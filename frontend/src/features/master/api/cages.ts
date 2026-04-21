import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { Cage as ModelCage } from "@/types/generated/models";

const CAGES_QUERY_KEY = ["masters", "cages"] as const;

// Strict union types (models.ts uses string, but we keep these for form safety)
export type CageType = "icu" | "dog" | "cat" | "general";
export type CageSize = "small" | "medium" | "large";

export interface CreateCageRequest {
  name: string;
  cage_type: CageType;
  cage_size: CageSize;
  price?: number;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

export interface UpdateCageRequest {
  name?: string;
  cage_type?: CageType;
  cage_size?: CageSize;
  price?: number;
  description?: string;
  is_active?: boolean;
  sort_order?: number;
}

function transformCage(data: ModelCage) {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    cageType: data.cage_type as CageType,
    cageSize: data.cage_size as CageSize,
    price: data.price ?? 0,
    isActive: data.is_active,
    description: data.description,
    sortOrder: data.sort_order,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  };
}

export type Cage = ReturnType<typeof transformCage>;

export const getAllCages = async (): Promise<Cage[]> => {
  const { data } = await axios.get<ModelCage[]>("/v1/masters/cages");
  return data.map(transformCage);
};

export const useGetAllCages = () => {
  return useQuery({
    queryKey: CAGES_QUERY_KEY,
    queryFn: getAllCages,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const getCageById = async (id: string): Promise<Cage> => {
  const { data } = await axios.get<ModelCage>(`/v1/masters/cages/${id}`);
  return transformCage(data);
};

export const createCage = async (req: CreateCageRequest): Promise<Cage> => {
  const { data } = await axios.post<ModelCage>("/v1/masters/cages", req);
  return transformCage(data);
};

export const useCreateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CAGES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const updateCage = async (id: string, req: UpdateCageRequest): Promise<Cage> => {
  const { data } = await axios.patch<ModelCage>(`/v1/masters/cages/${id}`, req);
  return transformCage(data);
};

export const useUpdateCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateCageRequest }) =>
      updateCage(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CAGES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

export const deleteCage = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/cages/${id}`);
};

export const reorderCages = async (req: { ids: number[] }): Promise<void> => {
  await axios.patch("/v1/masters/cages/reorder", req);
};

export const useDeleteCage = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteCage,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CAGES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const useReorderCages = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderCages,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: CAGES_QUERY_KEY });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
