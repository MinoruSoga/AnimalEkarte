import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformBackendMedicineToFrontend } from "@/lib/transforms/medicine";
import type { Medicine } from "@/lib/transforms/medicine";
import type { Medicine as MedicineModel } from "@/types/generated/models";
import type {
  CreateMedicineRequest,
  UpdateMedicineRequest,
  ReorderMedicinesRequest,
} from "@/types/medicine";

const getAllMedicines = async (): Promise<Medicine[]> => {
  const { data } = await axios.get<MedicineModel[]>("/v1/masters/medicines");
  return (data ?? []).map(transformBackendMedicineToFrontend);
};

export const useGetAllMedicines = () => {
  return useQuery({
    queryKey: queryKeys.masters.category("medicines"),
    queryFn: getAllMedicines,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
};

const createMedicine = async (req: CreateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.post<MedicineModel>("/v1/masters/medicines", req);
  return transformBackendMedicineToFrontend(data);
};

export const useCreateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createMedicine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("medicines") });
    },
    onError: (error) => handleApiError(error, "作成"),
  });
};

const updateMedicine = async (id: string, req: UpdateMedicineRequest): Promise<Medicine> => {
  const { data } = await axios.patch<MedicineModel>(`/v1/masters/medicines/${id}`, req);
  return transformBackendMedicineToFrontend(data);
};

export const useUpdateMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMedicineRequest }) =>
      updateMedicine(id, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("medicines") });
    },
    onError: (error) => handleApiError(error, "更新"),
  });
};

const deleteMedicine = async (id: string): Promise<void> => {
  await axios.delete(`/v1/masters/medicines/${id}`);
};

export const useDeleteMedicine = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: deleteMedicine,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("medicines") });
    },
    onError: (error) => handleApiError(error, "削除"),
  });
};

const reorderMedicines = async (req: ReorderMedicinesRequest): Promise<void> => {
  await axios.patch("/v1/masters/medicines/reorder", req);
};

export const useReorderMedicines = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: reorderMedicines,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.masters.category("medicines") });
    },
    onError: (error) => handleApiError(error, "並び替え"),
  });
};
