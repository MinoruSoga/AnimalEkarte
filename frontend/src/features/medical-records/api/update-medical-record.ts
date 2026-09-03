import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import { transformMedicalRecord } from "./transforms";
import type { MedicalRecord } from "./transforms";
import type { BackendMedicalRecord, UpdateMedicalRecordRequest } from "./types";

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）を保存する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
const updateMedicalRecord = async (
  id: string,
  req: UpdateMedicalRecordRequest,
  clinicId?: string,
): Promise<MedicalRecord> => {
  const { data } = await axios.patch<BackendMedicalRecord>(
    `/v1/medical-records/${id}`,
    req,
    clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined,
  );
  return transformMedicalRecord(data);
};

export const useUpdateMedicalRecord = (clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, req }: { id: string; req: UpdateMedicalRecordRequest }) =>
      updateMedicalRecord(id, req, clinicId),
    onSuccess: (_data, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(id) });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
