import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Estimate } from "@/types/generated/models";

export interface SaveEstimatePayload {
  title: string;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  discount_amount: number;
  comment: string;
  notes: string;
  medical_record_id?: number;
}

const getEstimatesByRecord = async (
  medicalRecordId: string,
): Promise<Estimate | null> => {
  const { data } = await axios.get<{ data: Estimate[] }>("/v1/estimates", {
    params: { medical_record_id: Number(medicalRecordId), limit: 1 },
  });
  return data.data?.[0] ?? null;
};

export const useGetEstimateByRecord = (medicalRecordId?: string) => {
  return useQuery({
    queryKey: ["estimate", "record", medicalRecordId],
    queryFn: () => getEstimatesByRecord(medicalRecordId!),
    enabled: !!medicalRecordId,
  });
};

export const useCreateEstimateRecord = (medicalRecordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: SaveEstimatePayload) =>
      axios.post<Estimate>("/v1/estimates", payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["estimate", "record", medicalRecordId] });
    },
  });
};

export const useUpdateEstimateRecord = (estimateId: number, medicalRecordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: Partial<SaveEstimatePayload>) =>
      axios.patch<Estimate>(`/v1/estimates/${estimateId}`, payload),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["estimate", "record", medicalRecordId] });
    },
  });
};
