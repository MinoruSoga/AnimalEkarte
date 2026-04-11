import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExaminationType } from "@/lib/transforms/treatment";
import type { ExamTypeField } from "@/lib/transforms/treatment";
import type { ExaminationType } from "@/types/generated/models";
import type {
  CreateExaminationTypeRequest,
  UpdateExaminationTypeRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { ExamTypeField };

export const getAllExaminationTypes = async (): Promise<ExamTypeField[]> => {
  const { data } = await axios.get<ExaminationType[]>("/v1/masters/examination-types");
  return data.map(transformExaminationType);
};

export const useGetAllExaminationTypes = () =>
  useQuery({
    queryKey: ["masters", "examination-types"],
    queryFn: getAllExaminationTypes,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateExaminationType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateExaminationTypeRequest): Promise<ExamTypeField> => {
      const { data } = await axios.post<ExaminationType>(
        "/v1/masters/examination-types",
        req,
      );
      return transformExaminationType(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["masters", "examination-types"] }),
  });
};

export const useUpdateExaminationType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateExaminationTypeRequest;
    }): Promise<ExamTypeField> => {
      const { data } = await axios.patch<ExaminationType>(
        `/v1/masters/examination-types/${id}`,
        req,
      );
      return transformExaminationType(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["masters", "examination-types"] }),
  });
};

export const useDeleteExaminationType = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/examination-types/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["masters", "examination-types"] }),
  });
};

export const useReorderExaminationTypes = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/examination-types/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["masters", "examination-types"] }),
  });
};
