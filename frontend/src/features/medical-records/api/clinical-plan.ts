// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";

function transformClinicalPlan(item: {
  id: string;
  medical_record_id: string;
  physical_exam: string;
  diagnosis_type_id?: string | null;
  diagnosis_name_id?: string | null;
  diagnosis_2_type_id?: string | null;
  diagnosis_2_name_id?: string | null;
  diagnosis_details: string;
  treatment_policy: string;
  created_at: string;
  updated_at: string;
  diagnosis_type?: { id: string; name: string } | null;
  diagnosis_name?: { id: string; name: string } | null;
  diagnosis_2_type?: { id: string; name: string } | null;
  diagnosis_2_name?: { id: string; name: string } | null;
  version: number;
}) {
  return {
    id: item.id,
    medical_record_id: item.medical_record_id,
    physical_exam: item.physical_exam,
    diagnosis_type_id: item.diagnosis_type_id,
    diagnosis_name_id: item.diagnosis_name_id,
    diagnosis_2_type_id: item.diagnosis_2_type_id,
    diagnosis_2_name_id: item.diagnosis_2_name_id,
    diagnosis_details: item.diagnosis_details,
    treatment_policy: item.treatment_policy,
    created_at: item.created_at,
    updated_at: item.updated_at,
    diagnosis_type: item.diagnosis_type,
    diagnosis_name: item.diagnosis_name,
    diagnosis_2_type: item.diagnosis_2_type,
    diagnosis_2_name: item.diagnosis_2_name,
    version: item.version,
  };
}
export type ClinicalPlan = ReturnType<typeof transformClinicalPlan>;

export interface UpdateClinicalPlanInput {
  physical_exam?: string;
  diagnosis_type_id?: number | null;
  diagnosis_name_id?: number | null;
  diagnosis_2_type_id?: number | null;
  diagnosis_2_name_id?: number | null;
  diagnosis_details?: string;
  treatment_policy?: string;
  version?: number;
}

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）の子リソースを操作する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
function clinicHeaderConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}

const getClinicalPlan = async (
  medicalRecordId: string,
  clinicId?: string
): Promise<ClinicalPlan> => {
  const { data } = await axios.get<Parameters<typeof transformClinicalPlan>[0]>(
    `/v1/medical-records/${medicalRecordId}/clinical-plan`,
    clinicHeaderConfig(clinicId)
  );
  return transformClinicalPlan(data);
};

export const useGetClinicalPlan = (medicalRecordId: string, clinicId?: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.clinicalPlan(medicalRecordId, clinicId),
    queryFn: () => getClinicalPlan(medicalRecordId, clinicId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export const useUpdateClinicalPlan = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: UpdateClinicalPlanInput) =>
      axios
        .patch<Parameters<typeof transformClinicalPlan>[0]>(
          `/v1/medical-records/${medicalRecordId}/clinical-plan`,
          input,
          clinicHeaderConfig(clinicId)
        )
        .then((r) => transformClinicalPlan(r.data)),
    onSuccess: (data) => {
      // BUG-010: toast は親 save action が一括で出す。mutation 側では出さない。
      // version CAS 用に応答を即キャッシュへ書き、invalidate 待ちで stale version を送らない。
      queryClient.setQueryData(
        queryKeys.medicalRecords.clinicalPlan(medicalRecordId, clinicId),
        data,
      );
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.clinicalPlan(medicalRecordId),
      });
    },
    // onError は置かない: mutateAsync 呼び出し側 (save action) が handleApiError する。
    // ここに置くと二重 toast になる。
  });
};
