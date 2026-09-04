import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { Estimate } from "@/types/generated/models";

interface SaveEstimateItemPayload {
  name: string;
  category?: string;
  unit_price: number;
  quantity: number;
  discount_rate: number;
  discount_amount: number;
  is_insurance_applicable: boolean;
  sort_order: number;
}

export interface SaveEstimatePayload {
  title: string;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  discount_amount: number;
  comment: string;
  notes: string;
  medical_record_id?: number;
  items?: SaveEstimateItemPayload[];
}

export interface UpdateEstimateRecordVariables {
  /** BUG-016: id は mutation 変数で渡し、hook 生成時のクロージャ id=0 を避ける */
  id: number;
  payload: Partial<SaveEstimatePayload>;
}

const getEstimatesByRecord = async (medicalRecordId: string): Promise<Estimate | null> => {
  const { data } = await axios.get<{ data: Estimate[] }>("/v1/estimates", {
    params: { medical_record_id: Number(medicalRecordId), limit: 1 },
  });
  return data.data?.[0] ?? null;
};

export const useGetEstimateByRecord = (medicalRecordId?: string) => {
  return useQuery({
    queryKey: queryKeys.medicalRecords.estimate(medicalRecordId!),
    queryFn: () => getEstimatesByRecord(medicalRecordId!),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};

export const useCreateEstimateRecord = (medicalRecordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: SaveEstimatePayload): Promise<Estimate> => {
      const { data } = await axios.post<Estimate>("/v1/estimates", payload);
      return data;
    },
    onSuccess: (created) => {
      // 直後の再保存で existing が null のまま POST されるのを防ぐ
      qc.setQueryData(queryKeys.medicalRecords.estimate(medicalRecordId), created);
      void qc.invalidateQueries({
        queryKey: queryKeys.medicalRecords.estimate(medicalRecordId),
      });
    },
    onError: (error) => handleApiError(error, "見積登録"),
  });
};

export const useUpdateEstimateRecord = (medicalRecordId: string) => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, payload }: UpdateEstimateRecordVariables): Promise<Estimate> => {
      const { data } = await axios.patch<Estimate>(`/v1/estimates/${id}`, payload);
      return data;
    },
    onSuccess: (updated) => {
      qc.setQueryData(queryKeys.medicalRecords.estimate(medicalRecordId), updated);
      void qc.invalidateQueries({
        queryKey: queryKeys.medicalRecords.estimate(medicalRecordId),
      });
    },
    onError: (error) => handleApiError(error, "見積更新"),
  });
};
