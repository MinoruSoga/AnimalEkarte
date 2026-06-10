import { useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { useAuth } from "@/hooks/use-auth";
import { useGetPet } from "@/hooks/use-pet";
import { calculateBillingTotals } from "@/lib/calculations";

import { getUnbilledItems } from "../api/get-unbilled-items";
import { useGetUngroupedSameDay } from "../api/get-ungrouped-items";
import type { PaymentSplitDraft } from "../components/PaymentCard";
import type { Accounting, AccountingItem, PaymentInfo } from "../types";
import { createInitialPaymentSplits } from "./AccountingDetailModel";

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

  const { data: newPetData } = useGetPet(newPetId);

  const baseAccounting = useMemo<Accounting | null>(() => {
    if (accountingId) {
      return fetchedAccounting ?? null;
    }
    const stateItems = locationState?.accountingItems ?? [];
    return {
      id: "acc_new",
      clinicId: currentClinicId ?? "",
      ownerId: newPetData?.ownerId ?? "",
      ownerName: newPetData?.ownerName ?? "飼い主様",
      petId: newPetId,
      petName: newPetData?.name ?? "ペット",
      petSpecies: newPetData?.species ?? "犬",
      status: "waiting",
      scheduledDate: new Date().toISOString().split("T")[0],
      items: stateItems,
      payment: undefined,
      totalRefundedAmount: 0,
    };
  }, [accountingId, currentClinicId, fetchedAccounting, locationState, newPetData, newPetId]);

  const baseItems = useMemo(() => baseAccounting?.items ?? [], [baseAccounting]);

  const { data: unbilledItems } = useQuery({
    queryKey: ["unbilledItems", newPetId],
    queryFn: () => getUnbilledItems(newPetId),
    enabled: !accountingId && !!newPetId,
    staleTime: 30_000,
  });

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

    const billingResult = calculateBillingTotals(
      accounting.items,
      0,
      0,
      0.10,
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
