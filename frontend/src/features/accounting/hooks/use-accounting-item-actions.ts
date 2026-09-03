import { useCallback, useLayoutEffect, useRef, type Dispatch, type SetStateAction } from "react";

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
import type { UpdateBillingItemRequest } from "../api/types";
import type { AccountingItem, AddAccountingItemInput, ItemCategory } from "../types";

type CreateManualBillingItemRequest = CreateBillingItemRequest & {
  other_reason?: string;
};

/** FE-RC-001: fieldset disabled 等の render 側ガードをバイパスされても各 handler で再検証する。 */
export interface AccountingItemMutationPermissions {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
}

const DENIED_ACCOUNTING_ITEM_PERMISSIONS: Readonly<AccountingItemMutationPermissions> = {
  canCreate: false,
  canEdit: false,
  canDelete: false,
};

interface UseAccountingItemActionsParams {
  accountingId: string | undefined;
  /** 会計 status（completed 時は明細 PATCH に修正理由必須 — BUG-009） */
  accountingStatus?: string;
  /** #115 / BUG-009 / BUG-021: 締め後・確定済み修正理由 */
  postCloseReason?: string;
  /** 確定済み明細修正に必要な権限 */
  canPostCloseEdit?: boolean;
  /** 対象日がレジ締め済みか */
  isScheduledDateClosed?: boolean;
  baseItems: AccountingItem[];
  queryClient: QueryClient;
  setLocalItems: Dispatch<SetStateAction<AccountingItem[] | null>>;
  setNewItemOpen: Dispatch<SetStateAction<boolean>>;
  startAddItemTransition: (callback: () => void) => void;
  startDeleteItemTransition: (callback: () => void) => void;
  startItemUpdateTransition: (callback: () => void) => void;
  /** FE-RC-001: handler 開始時に再検証する canCreate/canEdit/canDelete */
  permissions?: Readonly<AccountingItemMutationPermissions>;
}

/** 空文字は送らない。trim 後の理由のみ API に載せる（BUG-021 add/delete）。 */
export function buildPostCloseReasonField(postCloseReason?: string): { post_close_reason?: string } {
  const trimmed = postCloseReason?.trim();
  return trimmed ? { post_close_reason: trimmed } : {};
}

function buildPostClosePayload(args: {
  accountingStatus?: string;
  postCloseReason?: string;
  canPostCloseEdit?: boolean;
  isScheduledDateClosed?: boolean;
}): { ok: true; reason?: string } | { ok: false } {
  const isCompleted = args.accountingStatus === "completed";
  const needsReason = isCompleted || Boolean(args.isScheduledDateClosed);
  const optionalReason = (args.postCloseReason ?? "").trim() || undefined;
  // 通常フロー: 理由は任意配線のみ（BUG-021）。BE が締め時に必須検証する。
  if (!needsReason) {
    return optionalReason ? { ok: true, reason: optionalReason } : { ok: true };
  }
  // 確定済み / レジ締め済み: 権限 + 理由必須（BUG-009）
  if (isCompleted && !args.canPostCloseEdit) {
    toast.error("確定済み会計の明細修正には締め後編集権限が必要です");
    return { ok: false };
  }
  if (args.isScheduledDateClosed && !args.canPostCloseEdit) {
    toast.error("レジ締め済み期間の明細修正には締め後編集権限が必要です");
    return { ok: false };
  }
  if (!optionalReason) {
    toast.error(
      isCompleted
        ? "確定済み会計の明細を修正するには修正理由を入力してください"
        : "レジ締め済み期間の明細を修正するには修正理由を入力してください",
    );
    return { ok: false };
  }
  return { ok: true, reason: optionalReason };
}

export function useAccountingItemActions({
  accountingId,
  accountingStatus,
  postCloseReason,
  canPostCloseEdit,
  isScheduledDateClosed,
  baseItems,
  queryClient,
  setLocalItems,
  setNewItemOpen,
  startAddItemTransition,
  startDeleteItemTransition,
  startItemUpdateTransition,
  permissions = DENIED_ACCOUNTING_ITEM_PERMISSIONS,
}: UseAccountingItemActionsParams) {
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = permissions;
  }, [permissions]);
  const isMutationAllowed = useCallback(
    (action: keyof AccountingItemMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  const handleAddItem = useCallback(
    ({ name, price, category, otherReason, taxRate, merchandiseItemId }: AddAccountingItemInput) => {
      if (!isMutationAllowed(accountingId ? "canEdit" : "canCreate")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
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
      isMutationAllowed,
      postCloseReason,
      queryClient,
      setLocalItems,
      setNewItemOpen,
      startAddItemTransition,
    ],
  );

  const handleDeleteItem = useCallback(
    (itemId: string) => {
      if (!isMutationAllowed("canDelete")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
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
    [accountingId, baseItems, isMutationAllowed, postCloseReason, queryClient, setLocalItems, startDeleteItemTransition],
  );

  const handleUpdateItemTax = useCallback(
    (itemId: string, taxType: TaxType, taxRate: number) => {
      if (!accountingId) return;
      if (!isMutationAllowed("canEdit")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      const gate = buildPostClosePayload({
        accountingStatus,
        postCloseReason,
        canPostCloseEdit,
        isScheduledDateClosed,
      });
      if (!gate.ok) return;
      startItemUpdateTransition(async () => {
        try {
          const req: UpdateBillingItemRequest = {
            tax_type: taxType,
            tax_rate: taxRate,
            ...(gate.reason ? { post_close_reason: gate.reason } : {}),
          };
          await updateBillingItem(itemId, req);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "税区分の更新");
        }
      });
    },
    [
      accountingId,
      accountingStatus,
      canPostCloseEdit,
      isMutationAllowed,
      isScheduledDateClosed,
      postCloseReason,
      queryClient,
      startItemUpdateTransition,
    ],
  );

  const handleUpdateItemDiscount = useCallback(
    (itemId: string, discountAmount: number) => {
      if (!accountingId) return;
      if (!isMutationAllowed("canEdit")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      const gate = buildPostClosePayload({
        accountingStatus,
        postCloseReason,
        canPostCloseEdit,
        isScheduledDateClosed,
      });
      if (!gate.ok) return;
      startItemUpdateTransition(async () => {
        try {
          const req: UpdateBillingItemRequest = {
            discount_amount: discountAmount,
            ...(gate.reason ? { post_close_reason: gate.reason } : {}),
          };
          await updateBillingItem(itemId, req);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "割引の更新");
        }
      });
    },
    [
      accountingId,
      accountingStatus,
      canPostCloseEdit,
      isMutationAllowed,
      isScheduledDateClosed,
      postCloseReason,
      queryClient,
      startItemUpdateTransition,
    ],
  );

  return {
    handleAddItem,
    handleDeleteItem,
    handleUpdateItemTax,
    handleUpdateItemDiscount,
  };
}
