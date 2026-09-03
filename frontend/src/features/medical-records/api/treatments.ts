// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// Relative
import type {
  Treatment,
  CreateTreatmentInput,
  UpdateTreatmentInput,
  BulkReorderTreatmentsInput,
} from "../types";

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）の子リソースを操作する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
function clinicHeaderConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}

// ── Fetch ─────────────────────────────────────────────────────────────

const getTreatments = async (medicalRecordId: string, clinicId?: string): Promise<Treatment[]> => {
  const { data } = await axios.get<Treatment[]>(
    `/v1/medical-records/${medicalRecordId}/treatments`,
    clinicHeaderConfig(clinicId),
  );
  return data;
};

export const useGetTreatments = (medicalRecordId: string, clinicId?: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.treatments(medicalRecordId, clinicId),
    queryFn: () => getTreatments(medicalRecordId, clinicId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

// ── Create ────────────────────────────────────────────────────────────

export const useCreateTreatment = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateTreatmentInput) =>
      axios
        .post<Treatment>(
          `/v1/medical-records/${medicalRecordId}/treatments`,
          input,
          clinicHeaderConfig(clinicId),
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.treatments(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "治療追加");
    },
  });
};

// ── Update (PATCH) ────────────────────────────────────────────────────

export const useUpdateTreatment = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ treatmentId, input }: { treatmentId: string; input: UpdateTreatmentInput }) =>
      axios
        .patch<Treatment>(
          `/v1/medical-records/${medicalRecordId}/treatments/${treatmentId}`,
          input,
          clinicHeaderConfig(clinicId),
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.treatments(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "治療更新");
    },
  });
};

// ── Delete ────────────────────────────────────────────────────────────

export const useDeleteTreatment = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (treatmentId: string) =>
      axios.delete(
        `/v1/medical-records/${medicalRecordId}/treatments/${treatmentId}`,
        clinicHeaderConfig(clinicId),
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.treatments(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "治療削除");
    },
  });
};

// ── Reorder (PUT bulk update) ─────────────────────────────────────────

export const useReorderTreatments = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: BulkReorderTreatmentsInput) => {
      // バックエンドは数値IDを期待するため string -> number 変換
      const payload = {
        treatments: input.treatments.map((t) => ({
          id: Number(t.id),
          sort_order: t.sort_order,
        })),
      };
      return axios.put(
        `/v1/medical-records/${medicalRecordId}/treatments`,
        payload,
        clinicHeaderConfig(clinicId),
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.treatments(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "治療並び替え");
    },
  });
};
