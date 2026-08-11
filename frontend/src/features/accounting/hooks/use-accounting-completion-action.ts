import { useActionState, useEffect, useRef, useState } from "react";
import type { Dispatch, RefObject, SetStateAction } from "react";
import type { NavigateFunction } from "react-router";
import type { QueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { paths } from "@/config/paths";
import { handleApiError } from "@/lib/handle-api-error";
import { jstDateStartISOString, jstNowISOString } from "@/lib/jst-date";
import { queryKeys } from "@/lib/query-keys";

import {
  completeAccounting,
  createAccountingCompletionIdempotencyKey,
} from "../api/complete-accounting";
import type { CompleteAccountingItemRequest } from "../api/types";
import { updateAccounting } from "../api/update-accounting";
import type { PaymentSplitRequest } from "../api/types";
import type { PaymentSplitDraft } from "../components/PaymentCard";
import type { Accounting, AccountingItem, PaymentInfo, PaymentMethod } from "../types";
import { buildPaymentSplitRequests, type AccountingFormState } from "../components/accounting-detail-model";

function toCompleteItems(
  items: ReadonlyArray<AccountingItem>,
): CompleteAccountingItemRequest[] {
  return items.map((item) => ({
    category: item.category,
    name: item.name,
    unit_price: item.unitPrice,
    quantity: item.quantity,
    discount_rate: item.discountRate,
    discount_amount: item.discountAmount,
    tax_type: item.taxType,
    tax_rate: item.taxRate,
    is_insurance_applicable: item.isInsuranceApplicable,
    source: item.source,
    ...(item.source === "manual" && item.category === "other" && item.otherReason !== undefined
      ? { other_reason: item.otherReason }
      : {}),
    merchandise_item_id: item.merchandiseItemId ? Number(item.merchandiseItemId) : undefined,
    vaccination_id: item.vaccinationId ? Number(item.vaccinationId) : undefined,
    treatment_id: item.treatmentId ? Number(item.treatmentId) : undefined,
    appointment_id: item.appointmentId ? Number(item.appointmentId) : undefined,
    trimming_course_id: item.trimmingCourseId ? Number(item.trimmingCourseId) : undefined,
    trimming_option_id: item.trimmingOptionId ? Number(item.trimmingOptionId) : undefined,
  }));
}

/** BUG-011: complete ヘッダ medical_record_id。会計本体 → 明細の一意値の順で解決。 */
function resolveCompleteMedicalRecordId(
  accounting: Accounting,
  items: ReadonlyArray<AccountingItem>,
): number | undefined {
  if (accounting.medicalRecordId) {
    const n = Number(accounting.medicalRecordId);
    if (Number.isFinite(n) && n > 0) return n;
  }
  const fromItems = new Set<number>();
  for (const item of items) {
    if (!item.medicalRecordId) continue;
    const n = Number(item.medicalRecordId);
    if (Number.isFinite(n) && n > 0) fromItems.add(n);
  }
  if (fromItems.size === 1) {
    return [...fromItems][0];
  }
  return undefined;
}

/** mutation 単位の Idempotency-Key を payload 指紋付きで保持する（失敗 retry で再利用）。 */
function buildCompletePayloadFingerprint(parts: {
  petId: string | number;
  ownerId: string | number;
  medicalRecordId?: number;
  scheduledDate: string;
  items: CompleteAccountingItemRequest[];
  paymentSplits: PaymentSplitRequest[];
  postCloseReason?: string;
  insuranceRatio: string | null;
  insuranceAmount: number | null;
}): string {
  return JSON.stringify(parts);
}

interface AccountingCalculation {
  subtotal: number;
  taxTotal: number;
  totalAmount: number;
  insuranceAmount: number;
  billingAmount: number;
}

interface UseAccountingCompletionActionArgs {
  accountingId?: string;
  accounting: Accounting | null;
  calculation: AccountingCalculation | null;
  displayItems: AccountingItem[];
  hasInsurance: boolean;
  insuranceRatio: string;
  paymentSplits: PaymentSplitDraft[];
  queryClient: QueryClient;
  navigate: NavigateFunction;
  setCompletedPayment: Dispatch<SetStateAction<PaymentInfo | null>>;
  postCloseReason?: string; // #115: 締め後編集理由
  /** BUG-001: 新規会計でペット生死が拒否対象のとき complete を発行しない */
  blockCreateReason?: string;
}

export function useAccountingCompletionAction({
  accountingId,
  accounting,
  calculation,
  displayItems,
  hasInsurance,
  insuranceRatio,
  paymentSplits,
  queryClient,
  navigate,
  setCompletedPayment,
  postCloseReason,
  blockCreateReason,
}: UseAccountingCompletionActionArgs) {
  const [editConfirmOpen, setEditConfirmOpen] = useState(false);
  const editConfirmedRef = useRef(false);
  const formRef = useRef<HTMLFormElement>(null);
  // BUG-018: mutation 単位の Idempotency-Key。失敗 retry で再利用し、成功後のみクリア。
  const completeIdempotencyKeyRef = useRef<string | null>(null);
  const completePayloadFingerprintRef = useRef<string | null>(null);
  const blockCreateReasonRef = useRef(blockCreateReason);
  useEffect(() => {
    blockCreateReasonRef.current = blockCreateReason;
  }, [blockCreateReason]);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: AccountingFormState, _formData: FormData): Promise<AccountingFormState> => {
      if (!accounting || !calculation) return { success: false, timestamp: Date.now() };

      // BUG-001: 死亡 / 未確認ペットの新規確定は API を叩かず fail-closed。
      if (!accountingId && blockCreateReasonRef.current) {
        toast.error(blockCreateReasonRef.current);
        return { success: false, timestamp: Date.now() };
      }

      if (accounting.status === "completed" && !editConfirmedRef.current) {
        setEditConfirmOpen(true);
        return { success: false, timestamp: Date.now() };
      }
      editConfirmedRef.current = false;

      const builtSplits: PaymentSplitRequest[] = buildPaymentSplitRequests(paymentSplits);

      const repMethod: PaymentMethod =
        builtSplits.some((split) => split.method === "cash") ? "cash" :
        builtSplits.some((split) => split.method === "credit_card") ? "credit_card" :
        "electronic_money";

      const cashSplit = builtSplits.find((split) => split.method === "cash");
      const totalReceived = cashSplit ? cashSplit.received_amount ?? 0 : calculation.billingAmount;
      const totalChange = cashSplit ? cashSplit.change_amount ?? 0 : 0;

      const paymentInfo: PaymentInfo = {
        subtotal: calculation.subtotal,
        taxTotal: calculation.taxTotal,
        totalAmount: calculation.totalAmount,
        insuranceAmount: calculation.insuranceAmount,
        discountAmount: 0,
        billingAmount: calculation.billingAmount,
        receivedAmount: totalReceived,
        changeAmount: totalChange,
        method: repMethod,
        insuranceRatio: hasInsurance ? parseFloat(insuranceRatio) : undefined,
      };

      try {
        if (!accountingId) {
          // BUG-018: 新規確定は header/items/payments を単一 complete command で原子的に送信する。
          // legacy create + sequential items は残置（他 consumer 用）だが本経路では呼ばない。
          const completeItems = toCompleteItems(displayItems);
          const medicalRecordId = resolveCompleteMedicalRecordId(accounting, displayItems);
          const scheduledDate = accounting.scheduledDate
            ? jstDateStartISOString(accounting.scheduledDate)
            : jstNowISOString();
          const insuranceRatioValue = hasInsurance ? parseFloat(insuranceRatio) : null;
          const insuranceAmountValue =
            calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null;
          const fingerprint = buildCompletePayloadFingerprint({
            petId: accounting.petId,
            ownerId: accounting.ownerId,
            medicalRecordId,
            scheduledDate,
            items: completeItems,
            paymentSplits: builtSplits,
            postCloseReason: postCloseReason || undefined,
            insuranceRatio: insuranceRatioValue !== null ? String(insuranceRatioValue) : null,
            insuranceAmount: insuranceAmountValue,
          });
          if (
            completeIdempotencyKeyRef.current === null ||
            completePayloadFingerprintRef.current !== fingerprint
          ) {
            completeIdempotencyKeyRef.current = createAccountingCompletionIdempotencyKey();
            completePayloadFingerprintRef.current = fingerprint;
          }
          const idempotencyKey = completeIdempotencyKeyRef.current;
          const created = await completeAccounting(
            {
              pet_id: Number(accounting.petId),
              owner_id: Number(accounting.ownerId),
              medical_record_id: medicalRecordId ?? null,
              scheduled_date: scheduledDate,
              has_insurance: hasInsurance,
              insurance_ratio: insuranceRatioValue,
              insurance_amount: insuranceAmountValue,
              items: completeItems,
              payment_splits: builtSplits,
              post_close_reason: postCloseReason || undefined,
            },
            idempotencyKey,
          );
          completeIdempotencyKeyRef.current = null;
          completePayloadFingerprintRef.current = null;
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
          toast.success("会計を登録・完了しました");
          navigate(paths.accounting.detail.getHref(created.id));
        } else {
          await updateAccounting(accountingId, {
            status: "completed",
            subtotal: calculation.subtotal,
            tax_total: calculation.taxTotal,
            total_amount: calculation.totalAmount,
            insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
            insurance_amount: calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
            billing_amount: calculation.billingAmount,
            received_amount: totalReceived,
            change_amount: totalChange,
            payment_method: repMethod,
            payment_splits: builtSplits,
            completed_at: jstNowISOString(),
            post_close_reason: postCloseReason || undefined,
          });
          setCompletedPayment(paymentInfo);
          queryClient.invalidateQueries({ queryKey: queryKeys.accountings.all() });
          toast.success("会計を完了しました");
        }
        return { success: true, timestamp: Date.now() };
      } catch (error) {
        handleApiError(error, "会計の処理");
        return { success: false, timestamp: Date.now() };
      }
    },
    { success: false, timestamp: 0 },
  );

  useEffect(() => {
    if (formState.success === false && formState.timestamp > 0) {
      const element = document.getElementById("receivedAmount");
      if (element) {
        element.focus();
        element.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }
  }, [formState.success, formState.timestamp]);

  const confirmCompletedEdit = () => {
    editConfirmedRef.current = true;
    setEditConfirmOpen(false);
    requestAnimationFrame(() => {
      formRef.current?.requestSubmit();
    });
  };

  return {
    editConfirmOpen,
    setEditConfirmOpen,
    confirmCompletedEdit,
    formRef: formRef as RefObject<HTMLFormElement>,
    formAction,
    formState,
    isPending,
  };
}
