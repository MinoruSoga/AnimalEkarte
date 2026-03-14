// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";

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
    staleTime: 2 * 60 * 1000,
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
    onError: (error: Error) => {
      toast.error(error.message || "追加に失敗しました");
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
    onError: (error: Error) => {
      toast.error(error.message || "更新に失敗しました");
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
    onError: (error: Error) => {
      toast.error(error.message || "削除に失敗しました");
    },
  });
};
