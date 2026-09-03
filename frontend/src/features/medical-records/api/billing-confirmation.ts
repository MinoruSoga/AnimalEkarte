// React/Framework
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

// Internal
import { axios } from "@/lib/axios";
import { handleApiError } from "@/lib/handle-api-error";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";

// Relative
import type { BillingConfirmation, ReturnBillingConfirmationInput } from "../types";

// GET /v1/medical-records/:id/billing-confirmation
const getBillingConfirmation = async (medicalRecordId: string): Promise<BillingConfirmation> => {
  const { data } = await axios.get<BillingConfirmation>(
    `/v1/medical-records/${medicalRecordId}/billing-confirmation`,
  );
  return data;
};

export function useGetBillingConfirmation(medicalRecordId: string) {
  return useQuery({
    queryKey: queryKeys.medicalRecords.billingConfirmation(medicalRecordId),
    queryFn: () => getBillingConfirmation(medicalRecordId),
    enabled: !!medicalRecordId,
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}

// POST /v1/medical-records/:id/billing-confirmation/confirm
export function useCreateBillingConfirmation(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: { memo?: string }) =>
      axios.post<BillingConfirmation>(
        `/v1/medical-records/${medicalRecordId}/billing-confirmation/confirm`,
        { memo: input.memo },
      ),
    onSuccess: () => {
      // Invalidate billing-confirmation cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.billingConfirmation(medicalRecordId),
      });
      // Invalidate accountings list cache so new billing appears immediately
      void queryClient.invalidateQueries({
        queryKey: queryKeys.accountings.all(),
      });
    },
    onError: (error) => handleApiError(error, "会計確認"),
  });
}

// POST /v1/medical-records/:id/billing-confirmation/return
export function useCreateBillingReturn(medicalRecordId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ReturnBillingConfirmationInput) =>
      axios.post<BillingConfirmation>(
        `/v1/medical-records/${medicalRecordId}/billing-confirmation/return`,
        {
          return_reason: input.return_reason,
        },
      ),
    onSuccess: () => {
      // Invalidate billing-confirmation cache for current medical record
      void queryClient.invalidateQueries({
        queryKey: queryKeys.medicalRecords.billingConfirmation(medicalRecordId),
      });
      // Invalidate accountings list cache in case billing status changes
      void queryClient.invalidateQueries({
        queryKey: queryKeys.accountings.all(),
      });
    },
    onError: (error) => handleApiError(error, "会計差戻"),
  });
}
