import { useCallback, type Dispatch, type SetStateAction } from "react";

import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TaxType } from "@/types/generated/models";

import { createBillingItem } from "../api/create-billing-item";
import { deleteBillingItem } from "../api/delete-billing-item";
import { updateBillingItem } from "../api/update-billing-item";
import type { AccountingItem, ItemCategory } from "../types";

interface UseAccountingItemActionsParams {
  accountingId: string | undefined;
  baseItems: AccountingItem[];
  queryClient: QueryClient;
  setLocalItems: Dispatch<SetStateAction<AccountingItem[] | null>>;
  setNewItemOpen: Dispatch<SetStateAction<boolean>>;
  startAddItemTransition: (callback: () => void) => void;
  startDeleteItemTransition: (callback: () => void) => void;
  startItemUpdateTransition: (callback: () => void) => void;
}

export function useAccountingItemActions({
  accountingId,
  baseItems,
  queryClient,
  setLocalItems,
  setNewItemOpen,
  startAddItemTransition,
  startDeleteItemTransition,
  startItemUpdateTransition,
}: UseAccountingItemActionsParams) {
  const handleAddItem = useCallback(
    (name: string, price: string, category: string, taxRate?: number) => {
      const unitPrice = parseInt(price, 10);
      const qty = 1;
      const rate = taxRate ?? 0.1;
      const tempId = `manual_${crypto.randomUUID()}`;
      const newItem: AccountingItem = {
        id: tempId,
        category: category as ItemCategory,
        name,
        unitPrice,
        quantity: qty,
        discountRate: 0,
        discountAmount: 0,
        taxType: "excluded" as TaxType,
        taxRate: rate,
        taxAmount: Math.round(unitPrice * qty * rate),
        subtotal: unitPrice * qty,
        isInsuranceApplicable: false,
        source: "manual",
      };

      setLocalItems((prev) => [...(prev ?? baseItems), newItem]);
      setNewItemOpen(false);

      if (accountingId) {
        startAddItemTransition(async () => {
          try {
            await createBillingItem({
              billing_id: Number(accountingId),
              category,
              name,
              unit_price: unitPrice,
              quantity: qty,
              tax_type: "excluded",
              tax_rate: rate,
              is_insurance_applicable: false,
              source: "manual",
            });
            await queryClient.refetchQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
            setLocalItems(null);
            toast.success("明細を追加しました");
          } catch (error) {
            setLocalItems((prev) => (prev ?? []).filter((i) => i.id !== tempId));
            handleApiError(error, "明細の追加");
          }
        });
      } else {
        toast.success("明細を追加しました");
      }
    },
    [accountingId, baseItems, queryClient, setLocalItems, setNewItemOpen, startAddItemTransition],
  );

  const handleDeleteItem = useCallback(
    (itemId: string) => {
      if (!accountingId || itemId.startsWith("manual_")) {
        setLocalItems((prev) => (prev ?? baseItems).filter((i) => i.id !== itemId));
        return;
      }

      const rollbackItems = baseItems;
      setLocalItems((prev) => (prev ?? baseItems).filter((i) => i.id !== itemId));
      startDeleteItemTransition(async () => {
        try {
          await deleteBillingItem(itemId);
          await queryClient.refetchQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
          setLocalItems(null);
          toast.success("明細を削除しました");
        } catch (error) {
          setLocalItems(rollbackItems);
          handleApiError(error, "明細の削除");
        }
      });
    },
    [accountingId, baseItems, queryClient, setLocalItems, startDeleteItemTransition],
  );

  const handleUpdateItemTax = useCallback(
    (itemId: string, taxType: TaxType, taxRate: number) => {
      if (!accountingId) return;
      startItemUpdateTransition(async () => {
        try {
          await updateBillingItem(itemId, { tax_type: taxType, tax_rate: taxRate });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "税区分の更新");
        }
      });
    },
    [accountingId, queryClient, startItemUpdateTransition],
  );

  const handleUpdateItemDiscount = useCallback(
    (itemId: string, discountAmount: number) => {
      if (!accountingId) return;
      startItemUpdateTransition(async () => {
        try {
          await updateBillingItem(itemId, { discount_amount: discountAmount });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "割引の更新");
        }
      });
    },
    [accountingId, queryClient, startItemUpdateTransition],
  );

  return {
    handleAddItem,
    handleDeleteItem,
    handleUpdateItemTax,
    handleUpdateItemDiscount,
  };
}
