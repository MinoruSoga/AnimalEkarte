import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformCheckupType } from "@/lib/transforms/treatment";
import type { CheckupTypeItem } from "@/lib/transforms/treatment";
import type { CheckupType } from "@/types/generated/models";
import type {
  CreateCheckupTypeRequest,
  UpdateCheckupTypeRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { CheckupTypeItem };

const getAllCheckupTypes = async (): Promise<CheckupTypeItem[]> => {
  const { data } = await axios.get<CheckupType[]>("/v1/masters/checkup-types");
  return data.map(transformCheckupType);
};

export const useGetAllCheckupTypes = () =>
  useQuery({
    queryKey: queryKeys.masters.category("checkup-types"),
    queryFn: getAllCheckupTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateCheckupTypeRequest): Promise<CheckupTypeItem> => {
      const { data } = await axios.post<CheckupType>("/v1/masters/checkup-types", req);
      return transformCheckupType(data);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateCheckupTypeRequest;
    }): Promise<CheckupTypeItem> => {
      const { data } = await axios.patch<CheckupType>(`/v1/masters/checkup-types/${id}`, req);
      return transformCheckupType(data);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/checkup-types/${id}`),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderCheckupTypes = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/checkup-types/reorder", req),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: queryKeys.masters.category("checkup-types") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};
