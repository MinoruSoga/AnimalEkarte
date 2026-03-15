// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";

// ── Types ──────────────────────────────────────────────────────────────

export interface Checkup {
  id: string;
  medical_record_id: string;
  checkup_type_id: string;
  date: string; // YYYY-MM-DD
  next_date?: string | null;
  doctor_id?: string | null;
  result: string;
  created_at: string;
  updated_at: string;
  // nested
  checkup_type?: { id: string; name: string } | null;
  doctor?: { id: string; name: string } | null;
}

export interface CreateCheckupInput {
  checkup_type_id: number;
  date: string;
  next_date?: string | null;
  doctor_id?: number | null;
  result?: string;
}

export interface UpdateCheckupInput {
  checkup_type_id?: number;
  date?: string;
  next_date?: string | null;
  doctor_id?: number | null;
  result?: string;
}

// ── Fetch ──────────────────────────────────────────────────────────────

const getCheckups = async (medicalRecordId: string): Promise<Checkup[]> => {
  const { data } = await axios.get<Checkup[]>(
    `/v1/medical-records/${medicalRecordId}/checkups`
  );
  return data;
};

export const useCheckups = (medicalRecordId: string) => {
  return useQuery({
    queryKey: ["checkups", medicalRecordId],
    queryFn: () => getCheckups(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
  });
};

// ── Create ─────────────────────────────────────────────────────────────

export const useCreateCheckup = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateCheckupInput) =>
      axios
        .post<Checkup>(`/v1/medical-records/${medicalRecordId}/checkups`, input)
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["checkups", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "検査結果追加");
    },
  });
};

// ── Update (PATCH) ─────────────────────────────────────────────────────

export const useUpdateCheckup = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ checkupId, input }: { checkupId: string; input: UpdateCheckupInput }) =>
      axios
        .patch<Checkup>(
          `/v1/medical-records/${medicalRecordId}/checkups/${checkupId}`,
          input
        )
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["checkups", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "検査結果更新");
    },
  });
};

// ── Delete ─────────────────────────────────────────────────────────────

export const useDeleteCheckup = (medicalRecordId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (checkupId: string) =>
      axios.delete(`/v1/medical-records/${medicalRecordId}/checkups/${checkupId}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["checkups", medicalRecordId] });
    },
    onError: (error) => {
      handleApiError(error, "検査結果削除");
    },
  });
};
