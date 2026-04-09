// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";

// Relative
import type { BillingReview, ReturnBillingReviewInput } from "../types";

const billingReviewQueryKey = (medicalRecordId: string) =>
  ["medical-record", medicalRecordId, "billing-review"] as const;

// GET /v1/medical-records/:id/billing-review
const getBillingReview = async (medicalRecordId: string): Promise<BillingReview> => {
  const { data } = await axios.get<BillingReview>(
    `/v1/medical-records/${medicalRecordId}/billing-review`
  );
  return data;
};

export function useGetBillingReview(medicalRecordId: string) {
  return useQuery({
    queryKey: billingReviewQueryKey(medicalRecordId),
    queryFn: () => getBillingReview(medicalRecordId),
    enabled: !!medicalRecordId,
  });
}

// POST /v1/medical-records/:id/billing-review/confirm
export function useConfirmBillingReview(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { confirmed_by: number; memo?: string }) =>
      axios.post<BillingReview>(
        `/v1/medical-records/${medicalRecordId}/billing-review/confirm`,
        input
      ),
    onSuccess: () => {
      // Invalidate billing-review cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: billingReviewQueryKey(medicalRecordId),
      });
      // Invalidate accountings list cache so new billing appears immediately
      void queryClient.invalidateQueries({
        queryKey: ["accountings"],
      });
    },
  });
}

// POST /v1/medical-records/:id/billing-review/return
// rerender-defer-reads: userId は呼び出し側で useAuth から取得して渡す（hook 内の useAuth 依存を排除）
export function useReturnBillingReview(medicalRecordId: string, userId: number) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ReturnBillingReviewInput) =>
      axios.post<BillingReview>(
        `/v1/medical-records/${medicalRecordId}/billing-review/return`,
        {
          returned_by: userId,
          return_reason: input.return_reason,
        }
      ),
    onSuccess: () => {
      // Invalidate billing-review cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: billingReviewQueryKey(medicalRecordId),
      });
      // Invalidate accountings list cache in case billing status changes
      void queryClient.invalidateQueries({
        queryKey: ["accountings"],
      });
    },
  });
}
