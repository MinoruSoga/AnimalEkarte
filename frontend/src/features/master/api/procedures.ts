import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { transformProcedure } from "@/lib/transforms/treatment";
import type { ProcedureItem } from "@/lib/transforms/treatment";
import type { Procedure } from "@/types/generated/models";
import type {
  CreateProcedureRequest,
  UpdateProcedureRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { ProcedureItem };

const getAllProcedures = async (): Promise<ProcedureItem[]> => {
  const { data } = await axios.get<Procedure[]>("/v1/masters/procedures");
  return data.map(transformProcedure);
};

export const useGetAllProcedures = () =>
  useQuery({
    queryKey: queryKeys.masters.category("procedures"),
    queryFn: getAllProcedures,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateProcedure = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateProcedureRequest): Promise<ProcedureItem> => {
      const { data } = await axios.post<Procedure>("/v1/masters/procedures", req);
      return transformProcedure(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("procedures") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useUpdateProcedure = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateProcedureRequest;
    }): Promise<ProcedureItem> => {
      const { data } = await axios.patch<Procedure>(`/v1/masters/procedures/${id}`, req);
      return transformProcedure(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("procedures") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useDeleteProcedure = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/procedures/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("procedures") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};

export const useReorderProcedures = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/procedures/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: queryKeys.masters.category("procedures") }),
    onError: (error) => handleApiError(error, "操作"),
  });
};
