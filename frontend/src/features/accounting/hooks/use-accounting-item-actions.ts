import { useCallback, type Dispatch, type SetStateAction } from "react";

import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { handleApiError } from "@/lib/handle-api-error";
import { queryKeys } from "@/lib/query-keys";
import type { TaxType } from "@/types/generated/models";
import { DEFAULT_STANDARD_TAX_RATE } from "@/constants/tax";

import { createBillingItem } from "../api/create-billing-item";
import type { CreateBillingItemRequest } from "../api/create-billing-item";
import { deleteBillingItem } from "../api/delete-billing-item";
import { updateBillingItem } from "../api/update-billing-item";
import type { AccountingItem, AddAccountingItemInput, ItemCategory } from "../types";

type CreateManualBillingItemRequest = CreateBillingItemRequest & {
  other_reason?: string;
};

interface UseAccountingItemActionsParams {
  accountingId: string | undefined;
  baseItems: AccountingItem[];
  queryClient: QueryClient;
  setLocalItems: Dispatch<SetStateAction<AccountingItem[] | null>>;
  setNewItemOpen: Dispatch<SetStateAction<boolean>>;
  startAddItemTransition: (callback: () => void) => void;
  startDeleteItemTransition: (callback: () => void) => void;
  startItemUpdateTransition: (callback: () => void) => void;
  /** #115 / BUG-021: 締め後編集理由（空なら送らない。BE が締め時に必須検証） */
  postCloseReason?: string;
}

/** 空文字は送らない。trim 後の理由のみ API に載せる。 */
export function buildPostCloseReasonField(postCloseReason?: string): { post_close_reason?: string } {
  const trimmed = postCloseReason?.trim();
  return trimmed ? { post_close_reason: trimmed } : {};
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
  postCloseReason,
}: UseAccountingItemActionsParams) {
  const handleAddItem = useCallback(
    ({ name, price, category, otherReason, taxRate, merchandiseItemId }: AddAccountingItemInput) => {
      const unitPrice = parseInt(price, 10);
      const qty = 1;
      const rate = taxRate ?? DEFAULT_STANDARD_TAX_RATE;
      const tempId = `manual_${crypto.randomUUID()}`;
      const manualOtherReason = category === "other" ? otherReason : undefined;
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
        ...(manualOtherReason !== undefined ? { otherReason: manualOtherReason } : {}),
        merchandiseItemId,
      };

      setLocalItems((prev) => [...(prev ?? baseItems), newItem]);
      setNewItemOpen(false);

      if (accountingId) {
        startAddItemTransition(async () => {
          try {
            const request: CreateManualBillingItemRequest = {
              billing_id: Number(accountingId),
              category,
              name,
              unit_price: unitPrice,
              quantity: qty,
              tax_type: "excluded",
              tax_rate: rate,
              is_insurance_applicable: false,
              source: "manual",
              ...(manualOtherReason !== undefined ? { other_reason: manualOtherReason } : {}),
              merchandise_item_id: merchandiseItemId ? Number(merchandiseItemId) : undefined,
              ...buildPostCloseReasonField(postCloseReason),
            };
            await createBillingItem(request);
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
    [
      accountingId,
      baseItems,
      postCloseReason,
      queryClient,
      setLocalItems,
      setNewItemOpen,
      startAddItemTransition,
    ],
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
          const reasonField = buildPostCloseReasonField(postCloseReason);
          await deleteBillingItem(
            itemId,
            Object.keys(reasonField).length > 0 ? reasonField : undefined,
          );
          await queryClient.refetchQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
          setLocalItems(null);
          toast.success("明細を削除しました");
        } catch (error) {
          setLocalItems(rollbackItems);
          handleApiError(error, "明細の削除");
        }
      });
    },
    [accountingId, baseItems, postCloseReason, queryClient, setLocalItems, startDeleteItemTransition],
  );

  const handleUpdateItemTax = useCallback(
    (itemId: string, taxType: TaxType, taxRate: number) => {
      if (!accountingId) return;
      startItemUpdateTransition(async () => {
        try {
          await updateBillingItem(itemId, {
            tax_type: taxType,
            tax_rate: taxRate,
            ...buildPostCloseReasonField(postCloseReason),
          });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "税区分の更新");
        }
      });
    },
    [accountingId, postCloseReason, queryClient, startItemUpdateTransition],
  );

  const handleUpdateItemDiscount = useCallback(
    (itemId: string, discountAmount: number) => {
      if (!accountingId) return;
      startItemUpdateTransition(async () => {
        try {
          await updateBillingItem(itemId, {
            discount_amount: discountAmount,
            ...buildPostCloseReasonField(postCloseReason),
          });
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "割引の更新");
        }
      });
    },
    [accountingId, postCloseReason, queryClient, startItemUpdateTransition],
  );

  return {
    handleAddItem,
    handleDeleteItem,
    handleUpdateItemTax,
    handleUpdateItemDiscount,
  };
}
