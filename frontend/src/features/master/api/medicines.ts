import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformBackendMedicineToFrontend } from "@/lib/transforms/medicine";
import type { Medicine } from "@/lib/transforms/medicine";
import type { Medicine as MedicineModel } from "@/types/generated/models";
import type { CreateMedicineRequest, UpdateMedicineRequest, ReorderMedicinesRequest } from "@/types/medicine";

interface MedicinesResponse {
  data: MedicineModel[];
  total: number;
  page: number;
  limit: number;
}

export const getAllMedicines = async (): Promise<Medicine[]> => {
  const { data } = await axios.get<MedicinesResponse>("/v1/masters/medicines", {
    params: { limit: 100 },
  });
  return (data.data ?? []).map(transformBackendMedicineToFrontend);
};

export const useGetAllMedicines = () => {
  return useQuery({
    queryKey: ["masters", "medicines"],
    queryFn: getAllMedicines,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const getMedicineById = async (id: string): Promise<Medicine> => {
  const { data } = await axios.get<MedicineModel>(`/v1/masters/medicines/${id}`);
  return transformBackendMedicineToFrontend(data);
};

export const useGetMedicineById = (id: string) => {
  return useQuery({
    queryKey: ["masters", "medicines", id],
    queryFn: () => getMedicineById(id),
    enabled: !!id,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

export const createMedicine = async (req: CreateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.post<MedicineModel>("/v1/masters/medicines", req);
  return transformBackendMedicineToFrontend(data);
};

export const useCreateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createMedicine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "medicines"] });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

export const updateMedicine = async (id: string, req: UpdateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.patch<MedicineModel>(`/v1/masters/medicines/${id}`, req);
  return transformBackendMedicineToFrontend(data);
};

export const useUpdateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMedicineRequest }) =>
      updateMedicine(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "medicines"] });
    },
    onError: (error) => handleApiError(error, "更新"),
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
      queryClient.invalidateQueries({ queryKey: ["masters", "medicines"] });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

export const reorderMedicines = async (req: ReorderMedicinesRequest): Promise<void> => {
  await axios.patch("/v1/masters/medicines/reorder", req);
};

export const useReorderMedicines = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderMedicines,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["masters", "medicines"] });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
