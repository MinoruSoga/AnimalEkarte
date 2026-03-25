// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { useAuth } from "@/features/auth/hooks/use-auth";

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
      void queryClient.invalidateQueries({
        queryKey: billingReviewQueryKey(medicalRecordId),
      });
    },
  });
}

// POST /v1/medical-records/:id/billing-review/return
export function useReturnBillingReview(medicalRecordId: string) {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  return useMutation({
    mutationFn: (input: ReturnBillingReviewInput) =>
      axios.post<BillingReview>(
        `/v1/medical-records/${medicalRecordId}/billing-review/return`,
        {
          returned_by: Number(user?.id ?? 0),
          return_reason: input.return_reason,
        }
      ),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: billingReviewQueryKey(medicalRecordId),
      });
    },
  });
}
