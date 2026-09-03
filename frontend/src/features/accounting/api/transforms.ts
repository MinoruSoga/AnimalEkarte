import type { BackendAccounting, BackendAccountingItem } from "./types";
import type { BillingRefund, Payment, PaymentSplit } from "@/types/generated/models";
import type { ItemCategory, AccountingStatus, PaymentMethod } from "../types";
import { DEFAULT_STANDARD_TAX_RATE } from "@/constants/tax";

/** Payment にバックエンドが付与する結合フィールドを加えたローカル拡張型 */
type PaymentWithStaff = Payment & {
  paid_by_name?: string;
};

export function transformAccountingItem(item: BackendAccountingItem) {
  const unitPrice = item.unit_price ?? 0;
  const quantity = item.quantity ?? 1;
  const taxRate = item.tax_rate ?? DEFAULT_STANDARD_TAX_RATE;
  return {
    id: String(item.id ?? 0),
    category: item.category as ItemCategory,
    name: item.name,
    unitPrice,
    quantity,
    discountRate: item.discount_rate ?? 0,
    discountAmount: item.discount_amount ?? 0,
    taxType: item.tax_type ?? "excluded",
    taxRate,
    taxAmount: item.tax_amount ?? 0,
    subtotal: item.subtotal ?? Math.round(unitPrice * quantity) - (item.discount_amount ?? 0),
    isInsuranceApplicable: item.is_insurance_applicable,
    source: item.source as "medical_record" | "manual" | "hospitalization" | "trimming",
    otherReason: item.other_reason,
    merchandiseItemId: item.merchandise_item_id ? String(item.merchandise_item_id) : undefined,
    vaccinationId: item.vaccination_id ? String(item.vaccination_id) : undefined,
    examId: item.exam_id ? String(item.exam_id) : undefined,
    treatmentId: item.treatment_id ? String(item.treatment_id) : undefined,
    medicalRecordId: item.medical_record_id ? String(item.medical_record_id) : undefined,
    appointmentId: item.appointment_id ? String(item.appointment_id) : undefined,
    trimmingCourseId: item.trimming_course_id ? String(item.trimming_course_id) : undefined,
    trimmingOptionId: item.trimming_option_id ? String(item.trimming_option_id) : undefined,
  };
}

export type AccountingItem = ReturnType<typeof transformAccountingItem>;

function buildPaymentInfo(data: BackendAccounting) {
  const payment = data.payments?.[0] as PaymentWithStaff | undefined;
  if (data.status !== "completed" || payment == null) {
    return undefined;
  }
  return {
    subtotal: data.subtotal ?? 0,
    taxTotal: data.tax_total ?? 0,
    totalAmount: data.total_amount ?? 0,
    insuranceName: payment.insurance_name || undefined,
    insuranceRatio: payment.insurance_ratio ?? undefined,
    insuranceAmount: payment.insurance_amount ?? 0,
    discountAmount: payment.discount_amount ?? 0,
    billingAmount: payment.billing_amount ?? 0,
    receivedAmount: payment.received_amount ?? 0,
    changeAmount: payment.change_amount ?? 0,
    method: (payment.method || "cash") as PaymentMethod,
    paidByName: payment.paid_by_name,
  };
}

export function transformToRefund(r: BillingRefund & { refunded_by_name?: string }) {
  return {
    id: String(r.id ?? 0),
    billingId: String(r.billing_id ?? 0),
    amount: r.amount,
    reason: r.reason,
    refundedBy: r.refunded_by ?? null,
    refundedByName: r.refunded_by_name ?? "",
    paymentMethod: r.payment_method ?? undefined,
    refundedAt: r.refunded_at,
    createdAt: r.created_at,
  };
}

export type Refund = ReturnType<typeof transformToRefund>;

type PaymentSplitWithStaff = PaymentSplit & { paid_by_name?: string };

function transformPaymentSplit(s: PaymentSplitWithStaff) {
  return {
    id: String(s.id ?? 0),
    method: (s.method || "cash") as PaymentMethod,
    paymentMethodId: s.payment_method_id != null ? String(s.payment_method_id) : undefined,
    amount: s.amount ?? 0,
    receivedAmount: s.received_amount ?? 0,
    changeAmount: s.change_amount ?? 0,
    paidByName: s.paid_by_name || undefined,
  };
}

// Backend → フロントエンド Accounting 型（一覧・詳細共通）
export function transformToAccounting(data: BackendAccounting) {
  const splits = data.payment_splits;
  return {
    id: String(data.id ?? 0),
    clinicId: String(data.clinic_id),
    medicalRecordId: data.medical_record_id ? String(data.medical_record_id) : undefined,
    ownerId: String(data.owner_id ?? 0),
    ownerName: data.owner?.name ?? "",
    petId: String(data.pet_id ?? 0),
    petName: data.pet?.name ?? "",
    petSpecies: data.pet?.animal_species?.name,
    status: data.status as AccountingStatus,
    scheduledDate: data.scheduled_date ? data.scheduled_date.slice(0, 10) : "",
    completedAt: data.completed_at ?? undefined,
    items: (data.items ?? []).map(transformAccountingItem),
    payment: buildPaymentInfo(data),
    paymentSplits:
      splits && splits.length > 0
        ? splits.map((s) => transformPaymentSplit(s as PaymentSplitWithStaff))
        : undefined,
    totalRefundedAmount: data.total_refunded_amount ?? 0,
    /** BUG-007: 未収残高（waiting 全額 or クレジット訂正後の patient_due−支払額） */
    outstandingAmount: data.outstanding_amount ?? 0,
    memo: data.memo || undefined,
  };
}

export type Accounting = ReturnType<typeof transformToAccounting>;
