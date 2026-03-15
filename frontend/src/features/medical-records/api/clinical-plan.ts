// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

// ── Types ─────────────────────────────────────────────────────────────

export interface ClinicalPlan {
  id: string;
  medical_record_id: string;
  physical_exam: string;
  diagnosis_category_id?: string | null;
  diagnosis_name_id?: string | null;
  diagnosis_details: string;
  treatment_policy: string;
  created_at: string;
  updated_at: string;
  // nested relations
  diagnosis_category?: { id: string; name: string } | null;
  diagnosis_name?: { id: string; name: string } | null;
}

export interface UpdateClinicalPlanInput {
  physical_exam?: string;
  diagnosis_category_id?: number | null;
  diagnosis_name_id?: number | null;
  diagnosis_details?: string;
  treatment_policy?: string;
}

// ── Fetch ─────────────────────────────────────────────────────────────

export const getClinicalPlan = async (medicalRecordId: string): Promise<ClinicalPlan> => {
  const { data } = await axios.get<ClinicalPlan>(
    `/v1/medical-records/${medicalRecordId}/clinical-plan`
  );
  return data;
};

export const useClinicalPlan = (medicalRecordId: string) => {
  return useQuery({
    queryKey: ["clinical-plan", medicalRecordId],
    queryFn: () => getClinicalPlan(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
  });
};

// ── Update (PATCH) ────────────────────────────────────────────────────

export const useUpdateClinicalPlan = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: UpdateClinicalPlanInput) =>
      axios
        .patch<ClinicalPlan>(
          `/v1/medical-records/${medicalRecordId}/clinical-plan`,
          input
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["clinical-plan", medicalRecordId] });
      toast.success("保存しました");
    },
    onError: (error: Error) => {
      toast.error(error.message || "保存に失敗しました");
    },
  });
};
