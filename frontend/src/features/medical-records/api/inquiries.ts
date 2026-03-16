import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface UpdateInquiryRequest {
  chief_complaint?: string;
  chief_complaint_category_id?: number | null;
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
  });
};
