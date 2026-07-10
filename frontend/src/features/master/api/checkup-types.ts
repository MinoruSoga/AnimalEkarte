import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformCheckupType } from "@/lib/transforms/treatment";
import type { CheckupTypeItem } from "@/lib/transforms/treatment";
import type { CheckupType } from "@/types/generated/models";
import type {
  CreateCheckupTypeRequest,
  UpdateCheckupTypeRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

const CHECKUP_TYPES_QUERY_KEY = ["masters", "checkup-types"] as const;

export type { CheckupTypeItem };

const getAllCheckupTypes = async (): Promise<CheckupTypeItem[]> => {
  const { data } = await axios.get<CheckupType[]>("/v1/masters/checkup-types");
  return data.map(transformCheckupType);
};

export const useGetAllCheckupTypes = () =>
  useQuery({
    queryKey: CHECKUP_TYPES_QUERY_KEY,
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
    onSuccess: () => qc.invalidateQueries({ queryKey: CHECKUP_TYPES_QUERY_KEY }),
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
      const { data } = await axios.patch<CheckupType>(
        `/v1/masters/checkup-types/${id}`,
        req,
      );
      return transformCheckupType(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: CHECKUP_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteCheckupType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/checkup-types/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: CHECKUP_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderCheckupTypes = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/checkup-types/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: CHECKUP_TYPES_QUERY_KEY }),
    onError: (error) => handleApiError(error, "操作"),
  });
};
