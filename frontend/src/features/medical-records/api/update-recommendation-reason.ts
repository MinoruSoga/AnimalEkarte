import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { RecommendationReason } from "../constants/recommendation-reason";
import { transformMedicalRecord, type MedicalRecord } from "./transforms";
import type { BackendMedicalRecord } from "./types";

interface UpdateRecommendationReasonBody {
  reason: RecommendationReason | null;
}

const updateRecommendationReason = async (
  id: string,
  body: UpdateRecommendationReasonBody,
): Promise<MedicalRecord> => {
  const { data } = await axios.patch<BackendMedicalRecord>(
    `/v1/medical-records/${id}/recommendation-reason`,
    body,
  );
  return transformMedicalRecord(data);
};

export function useUpdateRecommendationReason(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: UpdateRecommendationReasonBody) =>
      updateRecommendationReason(medicalRecordId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.all() });
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.detail(medicalRecordId) });
    },
    onError: (error) => {
      handleApiError(error, "推奨理由更新");
    },
  });
}
