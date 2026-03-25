import { useMutation, useQueryClient } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { handleApiError } from "@/lib/handle-api-error";
import type { UpdateBillingItemRequest, BackendAccountingItem } from "./types";
import type { AccountingItem } from "../types";
import { transformToAccounting } from "./transforms";
import type { BackendAccounting } from "./types";

export const updateBillingItem = async (
  itemId: string,
  req: UpdateBillingItemRequest,
): Promise<BackendAccountingItem> => {
  const { data } = await axios.patch<BackendAccountingItem>(
    `/v1/billing-items/${itemId}`,
    req,
  );
  return data;
};

export const useUpdateBillingItem = (billingId: string) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({
      itemId,
      req,
    }: {
      itemId: string;
      req: UpdateBillingItemRequest;
    }) => updateBillingItem(itemId, req),
    onSuccess: () => {
      // Billing 合計が BE 側で再計算されるため詳細を再取得
      queryClient.invalidateQueries({
        queryKey: queryKeys.accountings.detail(billingId),
      });
    },
    onError: (error) => {
      handleApiError(error, "更新");
    },
  });
};
