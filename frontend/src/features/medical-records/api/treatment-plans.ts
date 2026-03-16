import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface UpdateTreatmentPlanRequest {
  plan?: string;
  assessment?: string;
  diagnosis_1_category_id?: number | null;
  diagnosis_1_name_id?: number | null;
  diagnosis_2_category_id?: number | null;
  diagnosis_2_name_id?: number | null;
}

export const useUpdateTreatmentPlan = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateTreatmentPlanRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/treatment-plans`, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
