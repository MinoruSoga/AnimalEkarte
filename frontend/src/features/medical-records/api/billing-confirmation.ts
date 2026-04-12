// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";

// Relative
import type { BillingConfirmation, ReturnBillingConfirmationInput } from "../types";

const billingConfirmationQueryKey = (medicalRecordId: string) =>
  ["medical-record", medicalRecordId, "billing-confirmation"] as const;

// GET /v1/medical-records/:id/billing-confirmation
const getBillingConfirmation = async (medicalRecordId: string): Promise<BillingConfirmation> => {
  const { data } = await axios.get<BillingConfirmation>(
    `/v1/medical-records/${medicalRecordId}/billing-confirmation`
  );
  return data;
};

export function useGetBillingConfirmation(medicalRecordId: string) {
  return useQuery({
    queryKey: billingConfirmationQueryKey(medicalRecordId),
    queryFn: () => getBillingConfirmation(medicalRecordId),
    enabled: !!medicalRecordId,
  });
}

// POST /v1/medical-records/:id/billing-confirmation/confirm
export function useConfirmBillingConfirmation(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { confirmed_by: number; memo?: string }) =>
      axios.post<BillingConfirmation>(
        `/v1/medical-records/${medicalRecordId}/billing-confirmation/confirm`,
        input
      ),
    onSuccess: () => {
      // Invalidate billing-confirmation cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: billingConfirmationQueryKey(medicalRecordId),
      });
      // Invalidate accountings list cache so new billing appears immediately
      void queryClient.invalidateQueries({
        queryKey: ["accountings"],
      });
    },
    onError: (error) => handleApiError(error, "会計確認"),
  });
}

// POST /v1/medical-records/:id/billing-confirmation/return
// rerender-defer-reads: userId は呼び出し側で useAuth から取得して渡す（hook 内の useAuth 依存を排除）
export function useReturnBillingConfirmation(medicalRecordId: string, userId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ReturnBillingConfirmationInput) =>
      axios.post<BillingConfirmation>(
        `/v1/medical-records/${medicalRecordId}/billing-confirmation/return`,
        {
          returned_by: userId,
          return_reason: input.return_reason,
        }
      ),
    onSuccess: () => {
      // Invalidate billing-confirmation cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: billingConfirmationQueryKey(medicalRecordId),
      });
      // Invalidate accountings list cache in case billing status changes
      void queryClient.invalidateQueries({
        queryKey: ["accountings"],
      });
    },
    onError: (error) => handleApiError(error, "会計差戻"),
  });
}
