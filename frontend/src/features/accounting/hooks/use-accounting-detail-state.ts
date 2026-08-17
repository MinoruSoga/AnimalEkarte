import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { useAuth } from "@/hooks/use-auth";
import { useGetPet } from "@/hooks/use-pet";
import { calculateBillingTotals } from "@/lib/calculations";
import { todayJSTISO } from "@/lib/jst-date";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";

import { getUnbilledItemDetails, type UnbilledWarning } from "../api/get-unbilled-items";
import { useGetUngroupedSameDay } from "../api/get-ungrouped-items";
import type { PaymentSplitDraft } from "../components/PaymentCard";
import type { Accounting, AccountingItem, PaymentInfo } from "../types";
import { createInitialPaymentSplits } from "../components/accounting-detail-model";

interface UseAccountingDetailStateArgs {
  accountingId?: string;
  locationSearch: string;
  locationState: { accountingItems?: AccountingItem[] } | null;
  fetchedAccounting?: Accounting;
}

export function useAccountingDetailState({
  accountingId,
  locationSearch,
  locationState,
  fetchedAccounting,
}: UseAccountingDetailStateArgs) {
  const { currentClinicId } = useAuth();
  const newPetId = useMemo(() => {
    if (accountingId) return "";
    return new URLSearchParams(locationSearch).get("petId") ?? "";
  }, [accountingId, locationSearch]);

  const {
    data: newPetData,
    isPending: newPetPending,
    isError: newPetError,
    isSuccess: newPetSuccess,
  } = useGetPet(newPetId);

  const baseAccounting = useMemo<Accounting | null>(() => {
    if (accountingId) {
      return fetchedAccounting ?? null;
    }
    if (currentClinicId === null) {
      return null;
    }
    const stateItems = locationState?.accountingItems ?? [];
    return {
      id: "acc_new",
      clinicId: currentClinicId,
      ownerId: newPetData?.ownerId ?? "",
      ownerName: newPetData?.ownerName ?? "飼い主様",
      petId: newPetId,
      petName: newPetData?.name ?? "ペット",
      petSpecies: newPetData?.species ?? "犬",
      status: "waiting",
      scheduledDate: todayJSTISO(),
      items: stateItems,
      payment: undefined,
      totalRefundedAmount: 0,
    };
  }, [accountingId, currentClinicId, fetchedAccounting, locationState, newPetData, newPetId]);

  const baseItems = useMemo(() => baseAccounting?.items ?? [], [baseAccounting]);

  // BUG-001: 死亡ペットへの /accounting/new?petId= 直打ちは FE で確定前に拒否する（BE と文言整合）。
  // 選択 UI と同様、生存が明示されるまで fail-closed（pending / error / 不明 status）。
  const isNewAccountingPetDeceased = Boolean(
    !accountingId && newPetId && newPetSuccess && newPetData?.status === "死亡",
  );
  const blocksDeceasedOrUnconfirmedPet = Boolean(
    !accountingId &&
      newPetId &&
      (newPetPending ||
        newPetError ||
        !newPetSuccess ||
        !newPetData ||
        newPetData.status !== "生存"),
  );
  // 表示メッセージは settle 後のみ（pending 中は fieldset disabled のみ）。
  const deceasedPetBlockMessage = (() => {
    if (accountingId || !newPetId || newPetPending) return undefined;
    if (isNewAccountingPetDeceased) return "死亡したペットは会計を作成できません";
    if (newPetError || (newPetSuccess && newPetData?.status !== "生存")) {
      return "ペットの生死状態を確認できないため、新規会計を作成できません";
    }
    return undefined;
  })();

  // BUG-013: new accounting consumer uses details envelope (items + typed warnings).
  // Fail-closed while pending/error: do not treat "unknown warnings" as clear.
  const {
    data: unbilledDetails,
    isPending: unbilledDetailsPending,
    isError: unbilledDetailsError,
    isSuccess: unbilledDetailsSuccess,
  } = useQuery({
    queryKey: queryKeys.unbilledItems(newPetId),
    queryFn: () => getUnbilledItemDetails(newPetId),
    enabled: !accountingId && !!newPetId && !isNewAccountingPetDeceased,
    staleTime: QUERY_STALE_TIMES.SHORT,
  });

  const unbilledItems = unbilledDetails?.items;
  const unbilledWarnings: UnbilledWarning[] = useMemo(
    () => unbilledDetails?.warnings ?? [],
    [unbilledDetails?.warnings],
  );
  const hasBlockingUnbilledWarning = useMemo(
    () => unbilledWarnings.some((w) => w.blocking && w.count > 0),
    [unbilledWarnings],
  );
  // New accounting only: ready when details succeeded (or not needed for existing accounting).
  // Deceased path skips unbilled fetch — treat as ready so the deceased banner is the sole block reason.
  const unbilledDetailsReady =
    Boolean(accountingId) || !newPetId || isNewAccountingPetDeceased || unbilledDetailsSuccess;
  const blocksNewAccountingSubmit =
    !accountingId &&
    !!newPetId &&
    (blocksDeceasedOrUnconfirmedPet ||
      unbilledDetailsPending ||
      unbilledDetailsError ||
      hasBlockingUnbilledWarning);

  const editableBaseItems = useMemo(
    () => (!accountingId && unbilledItems && unbilledItems.length > 0 ? unbilledItems : baseItems),
    [accountingId, baseItems, unbilledItems],
  );

  // #77: 同日同ペットの未会計対象化項目(取り残し)を検出して警告するためのサマリ。新規会計時のみ。
  const ungroupedScheduledDate = baseAccounting?.scheduledDate ?? "";
  const { data: ungroupedSummary } = useGetUngroupedSameDay(
    newPetId,
    ungroupedScheduledDate,
    !accountingId && !!newPetId,
  );

  const [localItems, setLocalItems] = useState<AccountingItem[] | null>(null);

  const displayItems = useMemo(
    () => localItems ?? editableBaseItems,
    [editableBaseItems, localItems],
  );

  const [completedPayment, setCompletedPayment] = useState<PaymentInfo | null>(null);

  const accounting = useMemo<Accounting | null>(() => {
    if (!baseAccounting) return null;
    return {
      ...baseAccounting,
      items: displayItems,
      payment: completedPayment ?? baseAccounting.payment,
      status: completedPayment ? "completed" : baseAccounting.status,
    };
  }, [baseAccounting, completedPayment, displayItems]);

  const [hasInsurance, setHasInsurance] = useState(false);
  const [insuranceRatio, setInsuranceRatio] = useState("0.5");
  const [paymentSplits, setPaymentSplits] = useState<PaymentSplitDraft[]>([]);

  const calculation = useMemo(() => {
    if (!accounting) return null;

    // BUG-006: 明細ごとの taxRate を尊重（10% 固定を渡さない）。
    // calculateBillingTotals が item.taxRate 未指定時のみ既定 10% にフォールバックする。
    const billingResult = calculateBillingTotals(
      accounting.items,
      0,
      0,
      undefined,
      hasInsurance ? parseFloat(insuranceRatio) : 0,
    );

    return {
      subtotal: billingResult.subtotal,
      taxTotal: billingResult.tax,
      totalAmount: billingResult.total,
      insuranceAmount: billingResult.insuranceAmount,
      billingAmount: billingResult.billingAmount,
    };
  }, [accounting, hasInsurance, insuranceRatio]);

  // Sync fetched accounting state to form fields (P1 bug fix: avoid race condition on mount)
  // ⚠️ レンダー中比較 (inline-comparison) に書き換えてはならない:
  // この state は useAccountingCompletionAction の useActionState action closure から参照される。
  // render-phase setState はマウント時の再レンダーパスで action closure を更新しないため、
  // マウント後に再レンダーなしで完了をサブミットすると action がデフォルト値
  // (paymentSplits=[] / hasInsurance=false) を見る。effect 同期なら commit 後の
  // 通常再レンダーで action が更新されるので安全。
  /* eslint-disable react-hooks/set-state-in-effect */
  useEffect(() => {
    if (!fetchedAccounting?.payment) return;
    setHasInsurance((fetchedAccounting.payment.insuranceAmount ?? 0) < 0);
    setInsuranceRatio(fetchedAccounting.payment.insuranceRatio?.toString() ?? "0.5");
    setPaymentSplits(createInitialPaymentSplits(fetchedAccounting));
  }, [fetchedAccounting]);
  /* eslint-enable react-hooks/set-state-in-effect */

  return {
    newPetId,
    baseAccounting,
    ungroupedSummary,
    unbilledWarnings,
    hasBlockingUnbilledWarning,
    unbilledDetailsReady,
    unbilledDetailsError,
    blocksNewAccountingSubmit,
    isNewAccountingPetDeceased,
    deceasedPetBlockMessage,
    baseItems: editableBaseItems,
    displayItems,
    localItems,
    setLocalItems,
    accounting,
    calculation,
    completedPayment,
    setCompletedPayment,
    hasInsurance,
    setHasInsurance,
    insuranceRatio,
    setInsuranceRatio,
    paymentSplits,
    setPaymentSplits,
  };
}
