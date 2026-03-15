import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformConsultation } from "@/lib/transforms/treatment";
import type { ConsultationItem } from "@/lib/transforms/treatment";
import type { Consultation } from "@/types/generated/models";
import type {
  CreateConsultationRequest,
  UpdateConsultationRequest,
  ReorderTreatmentRequest,
} from "@/types/treatment";

export type { ConsultationItem };

export const getAllConsultations = async (): Promise<ConsultationItem[]> => {
  const { data } = await axios.get<Consultation[]>("/v1/masters/consultations");
  return data.map(transformConsultation);
};

export const useGetAllConsultations = () =>
  useQuery({
    queryKey: ["consultations"],
    queryFn: getAllConsultations,
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });

export const useCreateConsultation = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (req: CreateConsultationRequest): Promise<ConsultationItem> => {
      const { data } = await axios.post<Consultation>("/v1/masters/consultations", req);
      return transformConsultation(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["consultations"] }),
  });
};

export const useUpdateConsultation = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      req,
    }: {
      id: string;
      req: UpdateConsultationRequest;
    }): Promise<ConsultationItem> => {
      const { data } = await axios.patch<Consultation>(
        `/v1/masters/consultations/${id}`,
        req,
      );
      return transformConsultation(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["consultations"] }),
  });
};

export const useDeleteConsultation = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => axios.delete(`/v1/masters/consultations/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["consultations"] }),
  });
};

export const useReorderConsultations = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (req: ReorderTreatmentRequest) =>
      axios.patch("/v1/masters/consultations/reorder", req),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["consultations"] }),
  });
};
