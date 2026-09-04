// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// Relative
import type { Vital, CreateVitalInput, UpdateVitalInput } from "../types";

// P2-15 (PR #186 review): 拠点横断で開いたカルテ（record.clinicId）の子リソースを操作する場合、
// グローバル選択クリニックではなくレコード自身の clinicId を X-Clinic-ID として送る必要がある。
// clinicId 省略時は axios インターセプタがグローバル選択値にフォールバックする（従来挙動を維持）。
function clinicHeaderConfig(clinicId?: string) {
  return clinicId ? { headers: { "X-Clinic-ID": clinicId } } : undefined;
}

// ── Fetch ─────────────────────────────────────────────────────────────

const getVitals = async (medicalRecordId: string, clinicId?: string): Promise<Vital[]> => {
  const { data } = await axios.get<Vital[]>(
    `/v1/medical-records/${medicalRecordId}/vitals`,
    clinicHeaderConfig(clinicId),
  );
  return data;
};

export const useGetVitals = (medicalRecordId: string, clinicId?: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.vitals(medicalRecordId, clinicId),
    queryFn: () => getVitals(medicalRecordId, clinicId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

// ── Create ────────────────────────────────────────────────────────────

export const useCreateVital = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateVitalInput) =>
      axios
        .post<Vital>(
          `/v1/medical-records/${medicalRecordId}/vitals`,
          input,
          clinicHeaderConfig(clinicId),
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.vitals(medicalRecordId) });
    },
    onError: (error) => {
      handleApiError(error, "バイタル追加");
    },
  });
};

// ── Update (PATCH) ────────────────────────────────────────────────────

export const useUpdateVital = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ vitalId, input }: { vitalId: string; input: UpdateVitalInput }) =>
      axios
        .patch<Vital>(
          `/v1/medical-records/${medicalRecordId}/vitals/${vitalId}`,
          input,
          clinicHeaderConfig(clinicId),
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.vitals(medicalRecordId) });
    },
    onError: (error) => {
      handleApiError(error, "バイタル更新");
    },
  });
};

// ── Delete ────────────────────────────────────────────────────────────

export const useDeleteVital = (medicalRecordId: string, clinicId?: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (vitalId: string) =>
      axios.delete(
        `/v1/medical-records/${medicalRecordId}/vitals/${vitalId}`,
        clinicHeaderConfig(clinicId),
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.medicalRecords.vitals(medicalRecordId) });
    },
    onError: (error) => {
      handleApiError(error, "バイタル削除");
    },
  });
};
