import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）を削除する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
const deleteMedicalRecord = async (id: string, clinicId?: string): Promise<void> => {
  await axios.delete(
    `/v1/medical-records/${id}`,
    clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined,
  );
};

export const useDeleteMedicalRecord = (clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteMedicalRecord(id, clinicId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.all() });
    },
    onError: (error) => handleApiError(error, "カルテ削除"),
  });
};
