// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// External
import { toast } from "sonner";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

// Relative
import type {
  Treatment,
  CreateTreatmentInput,
  UpdateTreatmentInput,
  BulkReorderTreatmentsInput,
} from "../types";

// ── Fetch ─────────────────────────────────────────────────────────────

export const getTreatments = async (medicalRecordId: string): Promise<Treatment[]> => {
  const { data } = await axios.get<Treatment[]>(
    `/v1/medical-records/${medicalRecordId}/treatments`
  );
  return data;
};

export const useTreatments = (medicalRecordId: string) => {
  return useQuery({
    queryKey: ["treatments", medicalRecordId],
    queryFn: () => getTreatments(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
  });
};

// ── Create ────────────────────────────────────────────────────────────

export const useCreateTreatment = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateTreatmentInput) =>
      axios
        .post<Treatment>(`/v1/medical-records/${medicalRecordId}/treatments`, input)
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["treatments", medicalRecordId] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "追加に失敗しました");
    },
  });
};

// ── Update (PATCH) ────────────────────────────────────────────────────

export const useUpdateTreatment = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ treatmentId, input }: { treatmentId: string; input: UpdateTreatmentInput }) =>
      axios
        .patch<Treatment>(
          `/v1/medical-records/${medicalRecordId}/treatments/${treatmentId}`,
          input
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["treatments", medicalRecordId] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "更新に失敗しました");
    },
  });
};

// ── Delete ────────────────────────────────────────────────────────────

export const useDeleteTreatment = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (treatmentId: string) =>
      axios.delete(`/v1/medical-records/${medicalRecordId}/treatments/${treatmentId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["treatments", medicalRecordId] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "削除に失敗しました");
    },
  });
};

// ── Reorder (PUT bulk update) ─────────────────────────────────────────

export const useReorderTreatments = (medicalRecordId: string) => {
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
        payload
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["treatments", medicalRecordId] });
    },
    onError: (error: Error) => {
      toast.error(error.message || "並び替えに失敗しました");
    },
  });
};
