import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";

export interface UpdateClinicalPlanRequest {
  treatment_policy?: string;
  diagnosis_details?: string;
  diagnosis_type_id?: number | null;
  diagnosis_name_id?: number | null;
  diagnosis_2_type_id?: number | null;
  diagnosis_2_name_id?: number | null;
}

export const useUpdateTreatmentPlan = (recordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateClinicalPlanRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/clinical-plan`, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
