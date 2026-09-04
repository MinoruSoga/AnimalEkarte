import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformVaccine } from "@/lib/transforms/treatment";
import type { VaccineItem } from "@/lib/transforms/treatment";
import type { Vaccine } from "@/types/generated/models";
import type {
  CreateVaccineRequest,
  UpdateVaccineRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { VaccineItem };

const getAllVaccinesMaster = async (): Promise<VaccineItem[]> => {
  const { data } = await axios.get<Vaccine[]>("/v1/masters/vaccines");
  return data.map(transformVaccine);
};

export const useGetAllVaccinesMaster = () =>
  useQuery({
    queryKey: queryKeys.masters.category("vaccines"),
    queryFn: getAllVaccinesMaster,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateVaccineRequest): Promise<VaccineItem> => {
      const { data } = await axios.post<Vaccine>("/v1/masters/vaccines", req);
      return transformVaccine(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateVaccineRequest;
    }): Promise<VaccineItem> => {
      const { data } = await axios.patch<Vaccine>(`/v1/masters/vaccines/${id}`, req);
      return transformVaccine(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteVaccineMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/vaccines/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderVaccinesMaster = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) => axios.patch("/v1/masters/vaccines/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("vaccines") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};
