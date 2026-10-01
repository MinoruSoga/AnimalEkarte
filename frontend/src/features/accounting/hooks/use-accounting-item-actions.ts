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
  /**
   * EMR-229: 手入力追加行の初期割引表示に使う飼主マスタ割引率(%)（unbilled-details の
   * owner_discount_rate）。省略時 0。POST body には載せず BE の自動割引解決に委ねるため
   * ローカルの optimistic 表示のみに使う。
   */
  ownerDiscountRate?: number;
}

/** EMR-230: 単価・数量・金額（小計）変更時のローカル再計算。BE BillingItem.CalculateTaxAmount と同式（割引後ベース・外税/内税は Math.round）。 */
function recomputeLineAmounts(
  item: Pick<AccountingItem, "discountAmount" | "taxType" | "taxRate">,
  unitPrice: number,
  quantity: number,
): { subtotal: number; taxAmount: number } {
  const subtotal = Math.max(Math.round(unitPrice * quantity) - item.discountAmount, 0);
  const taxAmount =
    item.taxType === "included"
      ? Math.round((subtotal * item.taxRate) / (1 + item.taxRate))
      : item.taxType === "excluded"
        ? Math.round(subtotal * item.taxRate)
        : 0;
  return { subtotal, taxAmount };
}

/** 既存会計で PATCH 可能な永続行かどうか（追加直後の仮ID `manual_<uuid>` は不可）。 */
function isPersistedItemId(itemId: string): boolean {
  return /^\d+$/.test(itemId);
}

/** 空文字は送らない。trim 後の理由のみ API に載せる（BUG-021 add/delete）。 */
export function buildPostCloseReasonField(postCloseReason?: string): {
  post_close_reason?: string;
} {
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
  ownerDiscountRate = 0,
}: UseAccountingItemActionsParams) {
  const permissionsRef = useRef(permissions);
  useLayoutEffect(() => {
    permissionsRef.current = permissions;
  }, [permissions]);
  const isMutationAllowed = useCallback(
    (action: keyof AccountingItemMutationPermissions) => permissionsRef.current[action] === true,
    [],
  );

  /** 新規会計ドラフト（accountingId なし）の明細編集は localItems へ immutable に反映する。 */
  const patchLocalItem = useCallback(
    (itemId: string, apply: (item: AccountingItem) => AccountingItem) => {
      setLocalItems((prev) => (prev ?? baseItems).map((i) => (i.id === itemId ? apply(i) : i)));
    },
    [baseItems, setLocalItems],
  );

  const handleAddItem = useCallback(
    ({
      name,
      price,
      category,
      otherReason,
      taxType,
      taxRate,
      merchandiseItemId,
    }: AddAccountingItemInput) => {
      if (!isMutationAllowed(accountingId ? "canEdit" : "canCreate")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      const unitPrice = parseInt(price, 10);
      const qty = 1;
      const rate = taxRate ?? DEFAULT_STANDARD_TAX_RATE;
      // EMR-65: マスタ登録の税区分を優先し、未指定時のみ従来の外税既定
      const resolvedTaxType: TaxType = taxType ?? "excluded";
      const tempId = `manual_${crypto.randomUUID()}`;
      const manualOtherReason = category === "other" ? otherReason : undefined;
      // EMR-229: 飼主マスタ割引率を optimistic な初期表示へ反映する。
      // request body には discount を載せない（BE がキャンペーンとの大きい方で再解決する）。
      const prefilledRate = Math.min(Math.max(ownerDiscountRate, 0), 100);
      const discountAmount = Math.round((unitPrice * qty * prefilledRate) / 100);
      // BE BillingItem.CalculateTaxAmount と同一規則（base = unitPrice*qty − discountAmount）。
      const { subtotal, taxAmount } = recomputeLineAmounts(
        { discountAmount, taxType: resolvedTaxType, taxRate: rate },
        unitPrice,
        qty,
      );
      const newItem: AccountingItem = {
        id: tempId,
        category: category as ItemCategory,
        name,
        unitPrice,
        quantity: qty,
        discountRate: prefilledRate,
        discountAmount,
        taxType: resolvedTaxType,
        taxRate: rate,
        taxAmount,
        subtotal,
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
              tax_type: resolvedTaxType,
              tax_rate: rate,
              is_insurance_applicable: false,
              source: "manual",
              ...(manualOtherReason !== undefined ? { other_reason: manualOtherReason } : {}),
              merchandise_item_id: merchandiseItemId ? Number(merchandiseItemId) : undefined,
              ...buildPostCloseReasonField(postCloseReason),
            };
            await createBillingItem(request);
            await queryClient.refetchQueries({
              queryKey: queryKeys.accountings.detail(accountingId),
            });
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
      ownerDiscountRate,
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
          await queryClient.refetchQueries({
            queryKey: queryKeys.accountings.detail(accountingId),
          });
          setLocalItems(null);
          toast.success("明細を削除しました");
        } catch (error) {
          setLocalItems(rollbackItems);
          handleApiError(error, "明細の削除");
        }
      });
    },
    [
      accountingId,
      baseItems,
      isMutationAllowed,
      postCloseReason,
      queryClient,
      setLocalItems,
      startDeleteItemTransition,
    ],
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

  // EMR-229: 項目名の編集。社販処理の目安として年月追記などを行う。
  // 新規会計ドラフトは localItems、既存会計は PATCH /v1/billing-items/:id {name}。
  const handleUpdateItemName = useCallback(
    (itemId: string, name: string) => {
      const trimmed = name.trim();
      if (trimmed === "") {
        toast.error("項目名は必須です");
        return;
      }
      if (!isMutationAllowed(accountingId ? "canEdit" : "canCreate")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      if (!accountingId) {
        patchLocalItem(itemId, (i) => ({ ...i, name: trimmed }));
        return;
      }
      if (!isPersistedItemId(itemId)) {
        // 追加直後の仮ID行は作成 POST 反映後に編集可能になる（仮IDで PATCH すると失敗する）。
        toast.error("明細の登録が反映されてから編集してください");
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
            name: trimmed,
            ...(gate.reason ? { post_close_reason: gate.reason } : {}),
          };
          await updateBillingItem(itemId, req);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "項目名の更新");
        }
      });
    },
    [
      accountingId,
      accountingStatus,
      canPostCloseEdit,
      isMutationAllowed,
      isScheduledDateClosed,
      patchLocalItem,
      postCloseReason,
      queryClient,
      startItemUpdateTransition,
    ],
  );

  // EMR-230: 数量の編集（単価×数量−割引=金額の明細内訳に対応）。数量は正の値のみ。
  const handleUpdateItemQuantity = useCallback(
    (itemId: string, quantity: number) => {
      if (!Number.isFinite(quantity) || quantity <= 0) {
        toast.error("数量は正の値で入力してください");
        return;
      }
      if (!isMutationAllowed(accountingId ? "canEdit" : "canCreate")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      if (!accountingId) {
        patchLocalItem(itemId, (i) => ({
          ...i,
          quantity,
          ...recomputeLineAmounts(i, i.unitPrice, quantity),
        }));
        return;
      }
      if (!isPersistedItemId(itemId)) {
        toast.error("明細の登録が反映されてから編集してください");
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
            quantity,
            ...(gate.reason ? { post_close_reason: gate.reason } : {}),
          };
          await updateBillingItem(itemId, req);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "数量の更新");
        }
      });
    },
    [
      accountingId,
      accountingStatus,
      canPostCloseEdit,
      isMutationAllowed,
      isScheduledDateClosed,
      patchLocalItem,
      postCloseReason,
      queryClient,
      startItemUpdateTransition,
    ],
  );

  // EMR-230: 金額（税抜小計 = 単価×数量−割引額）の直接編集。
  // 新規カラムを増やさず unit_price = round((amount + discountAmount) / quantity) に換算して PATCH する
  // （サーバ側が小計・税・合計を再計算して正本化する）。
  // 数量>1 で割り切れない金額は表現不能なため round が最も近い表示可能値に寄せ、
  // 再計算後の表示が入力量と最大 数量/2 円ずれうる（スキーマ上の不可避な制約）。
  const handleUpdateItemAmount = useCallback(
    (item: AccountingItem, amount: number) => {
      if (!Number.isFinite(amount) || amount < 0) {
        toast.error("金額は0以上の値で入力してください");
        return;
      }
      if (!Number.isFinite(item.quantity) || item.quantity <= 0) {
        toast.error("数量が0のため金額を編集できません");
        return;
      }
      if (!isMutationAllowed(accountingId ? "canEdit" : "canCreate")) {
        toast.error("この操作を行う権限がありません");
        return;
      }
      const unitPrice = Math.round((amount + item.discountAmount) / item.quantity);
      if (!accountingId) {
        patchLocalItem(item.id, (i) => ({
          ...i,
          unitPrice,
          ...recomputeLineAmounts(i, unitPrice, i.quantity),
        }));
        return;
      }
      if (!isPersistedItemId(item.id)) {
        toast.error("明細の登録が反映されてから編集してください");
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
            unit_price: unitPrice,
            ...(gate.reason ? { post_close_reason: gate.reason } : {}),
          };
          await updateBillingItem(item.id, req);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.detail(accountingId) });
        } catch (error) {
          handleApiError(error, "金額の更新");
        }
      });
    },
    [
      accountingId,
      accountingStatus,
      canPostCloseEdit,
      isMutationAllowed,
      isScheduledDateClosed,
      patchLocalItem,
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
    handleUpdateItemName,
    handleUpdateItemQuantity,
    handleUpdateItemAmount,
  };
}
