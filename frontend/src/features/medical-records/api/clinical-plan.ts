// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// ── Types ─────────────────────────────────────────────────────────────

function transformClinicalPlan(item: {
  id: string;
  medical_record_id: string;
  physical_exam: string;
  diagnosis_type_id?: string | null;
  diagnosis_name_id?: string | null;
  diagnosis_details: string;
  treatment_policy: string;
  created_at: string;
  updated_at: string;
  diagnosis_type?: { id: string; name: string } | null;
  diagnosis_name?: { id: string; name: string } | null;
}) {
  return {
    id: item.id,
    medical_record_id: item.medical_record_id,
    physical_exam: item.physical_exam,
    diagnosis_type_id: item.diagnosis_type_id,
    diagnosis_name_id: item.diagnosis_name_id,
    diagnosis_details: item.diagnosis_details,
    treatment_policy: item.treatment_policy,
    created_at: item.created_at,
    updated_at: item.updated_at,
    diagnosis_type: item.diagnosis_type,
    diagnosis_name: item.diagnosis_name,
  };
}
export type ClinicalPlan = ReturnType<typeof transformClinicalPlan>;

export interface UpdateClinicalPlanInput {
  physical_exam?: string;
  diagnosis_type_id?: number | null;
  diagnosis_name_id?: number | null;
  diagnosis_details?: string;
  treatment_policy?: string;
}

// ── Fetch ─────────────────────────────────────────────────────────────

const getClinicalPlan = async (medicalRecordId: string): Promise<ClinicalPlan> => {
  const { data } = await axios.get<Parameters<typeof transformClinicalPlan>[0]>(
    `/v1/medical-records/${medicalRecordId}/clinical-plan`
  );
  return transformClinicalPlan(data);
};

export const useGetClinicalPlan = (medicalRecordId: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.clinicalPlan(medicalRecordId),
    queryFn: () => getClinicalPlan(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
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
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.clinicalPlan(medicalRecordId) });
      toast.success("保存しました");
    },
    onError: (error) => {
      handleApiError(error, "保存");
    },
  });
};
