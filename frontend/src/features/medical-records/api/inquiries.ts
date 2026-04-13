import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

export interface UpdateInquiryRequest {
  chief_complaint?: string;
  chief_complaint_type_id?: number | null;
  notes?: string;
}

export const useUpdateInquiry = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateInquiryRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/inquiries`, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
    onError: (error) => handleApiError(error, "問診更新"),
  });
};
