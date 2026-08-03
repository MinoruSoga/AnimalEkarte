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
}: UseAccountingCompletionActionArgs) {
  const [editConfirmOpen, setEditConfirmOpen] = useState(false);
  const editConfirmedRef = useRef(false);
  const formRef = useRef<HTMLFormElement>(null);

  const [formState, formAction, isPending] = useActionState(
    async (_prevState: AccountingFormState, _formData: FormData): Promise<AccountingFormState> => {
      if (!accounting || !calculation) return { success: false, timestamp: Date.now() };

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
          const idempotencyKey = createAccountingCompletionIdempotencyKey();
          const created = await completeAccounting(
            {
              pet_id: Number(accounting.petId),
              owner_id: Number(accounting.ownerId),
              scheduled_date: accounting.scheduledDate
                ? jstDateStartISOString(accounting.scheduledDate)
                : jstNowISOString(),
              has_insurance: hasInsurance,
              insurance_ratio: hasInsurance ? parseFloat(insuranceRatio) : null,
              insurance_amount: calculation.insuranceAmount !== 0 ? calculation.insuranceAmount : null,
              items: toCompleteItems(displayItems),
              payment_splits: builtSplits,
              post_close_reason: postCloseReason || undefined,
            },
            idempotencyKey,
          );
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
