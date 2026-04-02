import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface CreateEstimateRequest {
  medical_record_id?: string;
  [key: string]: unknown;
}

export const useCreateEstimate = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateEstimateRequest) =>
      axios.post(`/v1/medical-records/${recordId}/estimates`, {
        ...input,
        medical_record_id: recordId,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
