import { useCallback, type Dispatch, type SetStateAction } from "react";
import type { NavigateFunction } from "react-router";

import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { paths } from "@/config/paths";
import { formatCurrency } from "@/lib/format/number";
import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";

import { cancelAccounting } from "../api/cancel-accounting";
import { createRefund } from "../api/create-refund";
import type { PaymentMethod } from "../types";

interface UseAccountingSettlementActionsParams {
  accountingId: string | undefined;
  navigate: NavigateFunction;
  queryClient: QueryClient;
  setCancelConfirmOpen: Dispatch<SetStateAction<boolean>>;
  setPreviewOpen: Dispatch<SetStateAction<boolean>>;
  startCancelTransition: (callback: () => void) => void;
  startRefundTransition: (callback: () => void) => void;
}

export function useAccountingSettlementActions({
  accountingId,
  navigate,
  queryClient,
  setCancelConfirmOpen,
  setPreviewOpen,
  startCancelTransition,
  startRefundTransition,
}: UseAccountingSettlementActionsParams) {
  const handleRefund = useCallback(
    (amount: number, reason: string, paymentMethod?: PaymentMethod) => {
      if (!accountingId) return;
      startRefundTransition(async () => {
        try {
          await createRefund(accountingId, { amount, reason, paymentMethod });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountingRefunds(accountingId) });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
          toast.success(`${formatCurrency(amount)} の返金を登録しました`);
        } catch (error) {
          handleApiError(error, "返金の登録");
        }
      });
    },
    [accountingId, queryClient, startRefundTransition],
  );

  const handlePrint = useCallback(() => {
    setPreviewOpen(true);
  }, [setPreviewOpen]);

  const handleCancelConfirm = useCallback(() => {
    if (!accountingId) return;
    startCancelTransition(async () => {
      try {
        await cancelAccounting(accountingId);
        toast.success("会計をキャンセルしました");
        await queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
        navigate(paths.accounting.getHref());
      } catch (error) {
        handleApiError(error, "会計のキャンセル");
      } finally {
        setCancelConfirmOpen(false);
      }
    });
  }, [accountingId, queryClient, navigate, setCancelConfirmOpen, startCancelTransition]);

  return {
    handleCancelConfirm,
    handlePrint,
    handleRefund,
  };
}
