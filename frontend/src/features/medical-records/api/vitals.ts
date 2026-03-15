// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import type { Vital, CreateVitalInput, UpdateVitalInput } from "../types";

// ── Fetch ─────────────────────────────────────────────────────────────

const getVitals = async (medicalRecordId: string): Promise<Vital[]> => {
  const { data } = await axios.get<Vital[]>(
    `/v1/medical-records/${medicalRecordId}/vitals`
  );
  return data;
};

export const useVitals = (medicalRecordId: string) => {
  return useQuery({
    queryKey: ["vitals", medicalRecordId],
    queryFn: () => getVitals(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
  });
};

// ── Create ────────────────────────────────────────────────────────────

export const useCreateVital = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateVitalInput) =>
      axios
        .post<Vital>(`/v1/medical-records/${medicalRecordId}/vitals`, input)
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vitals", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "バイタル追加");
    },
  });
};

// ── Update (PATCH) ────────────────────────────────────────────────────

export const useUpdateVital = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ vitalId, input }: { vitalId: string; input: UpdateVitalInput }) =>
      axios
        .patch<Vital>(
          `/v1/medical-records/${medicalRecordId}/vitals/${vitalId}`,
          input
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vitals", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "バイタル更新");
    },
  });
};

// ── Delete ────────────────────────────────────────────────────────────

export const useDeleteVital = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (vitalId: string) =>
      axios.delete(`/v1/medical-records/${medicalRecordId}/vitals/${vitalId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["vitals", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "バイタル削除");
    },
  });
};
