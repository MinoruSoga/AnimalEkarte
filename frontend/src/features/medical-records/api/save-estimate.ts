import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TaxType } from "@/types/generated/models";
import type { LegacyEstimateResponse } from "@/types/estimate-legacy-wire";

export interface SaveEstimateItemPayload {
  name: string;
  category?: string;
  unit_price: number;
  quantity: number;
  discount_rate: number;
  discount_amount: number;
  is_insurance_applicable: boolean;
  tax_type?: TaxType;
  tax_rate?: number;
  sort_order: number;
  /**
   * PR665: 既存行との一致ヒント。現在の見積に保存済みと証明できる行（取得または
   * 保存応答で id を確認できた行）のみ送る。新規クライアント行では送らない。
   */
  existing_item_id?: number;
}

export interface SaveEstimatePayload {
  title: string;
  subtotal: number;
  tax_total: number;
  total_amount: number;
  discount_amount: number;
  /** PR669: 保険適用額。省略時は既存互換（BE は保存済み値を保持）。 */
  insurance_amount?: number;
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

const getEstimatesByRecord = async (
  medicalRecordId: string,
): Promise<LegacyEstimateResponse | null> => {
  const { data } = await axios.get<{ data: LegacyEstimateResponse[] }>("/v1/estimates", {
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
    mutationFn: async (payload: SaveEstimatePayload): Promise<LegacyEstimateResponse> => {
      const { data } = await axios.post<LegacyEstimateResponse>("/v1/estimates", payload);
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
    mutationFn: async ({
      id,
      payload,
    }: UpdateEstimateRecordVariables): Promise<LegacyEstimateResponse> => {
      const { data } = await axios.patch<LegacyEstimateResponse>(`/v1/estimates/${id}`, payload);
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
