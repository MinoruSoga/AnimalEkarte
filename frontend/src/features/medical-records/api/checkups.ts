// React/Framework
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

// ── Types ──────────────────────────────────────────────────────────────

function transformCheckup(item: {
  id: string;
  medical_record_id: string;
  checkup_type_id: string;
  date: string;
  next_date?: string | null;
  doctor_id?: string | null;
  result: string;
  created_at: string;
  updated_at: string;
  checkup_type?: { id: string; name: string } | null;
  doctor?: { id: string; name: string } | null;
}) {
  return {
    id: item.id,
    medical_record_id: item.medical_record_id,
    checkup_type_id: item.checkup_type_id,
    date: item.date,
    next_date: item.next_date,
    doctor_id: item.doctor_id,
    result: item.result,
    created_at: item.created_at,
    updated_at: item.updated_at,
    checkup_type: item.checkup_type,
    doctor: item.doctor,
  };
}
export type Checkup = ReturnType<typeof transformCheckup>;

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
  doctor_id_clear?: boolean;
  result?: string;
}

// ── Fetch ──────────────────────────────────────────────────────────────

const getCheckups = async (medicalRecordId: string): Promise<Checkup[]> => {
  const { data } = await axios.get<Parameters<typeof transformCheckup>[0][]>(
    `/v1/medical-records/${medicalRecordId}/checkups`,
  );
  return data.map(transformCheckup);
};

export const useGetCheckups = (medicalRecordId: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.checkups(medicalRecordId),
    queryFn: () => getCheckups(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
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
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.checkups(medicalRecordId),
      });
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
        .patch<Checkup>(`/v1/medical-records/${medicalRecordId}/checkups/${checkupId}`, input)
        .then((r) => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.checkups(medicalRecordId),
      });
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
      queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.checkups(medicalRecordId),
      });
    },
    onError: (error) => {
      handleApiError(error, "検査結果削除");
    },
  });
};
