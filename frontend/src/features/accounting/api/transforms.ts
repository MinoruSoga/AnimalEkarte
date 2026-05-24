import type { BackendAccounting, BackendAccountingItem } from "./types";
import type { BillingRefund, Payment, PaymentSplit } from "@/types/generated/models";
import type { ItemCategory, AccountingStatus, PaymentMethod } from "../types";

/** Payment にバックエンドが付与する結合フィールドを加えたローカル拡張型 */
type PaymentWithStaff = Payment & {
  paid_by_name?: string;
};

export function transformAccountingItem(item: BackendAccountingItem) {
  const unitPrice = item.unit_price ?? 0;
  const quantity = item.quantity ?? 1;
  const taxRate = item.tax_rate ?? 0.1;
  return {
    id: String(item.id ?? 0),
    category: item.category as ItemCategory,
    name: item.name,
    unitPrice,
    quantity,
    taxType: item.tax_type ?? "excluded",
    taxRate,
    taxAmount: item.tax_amount ?? 0,
    subtotal: item.subtotal ?? Math.round(unitPrice * quantity),
    isInsuranceApplicable: item.is_insurance_applicable,
    source: item.source as "medical_record" | "manual",
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

export type PaymentInfo = NonNullable<ReturnType<typeof buildPaymentInfo>>;

export function transformToRefund(r: BillingRefund & { refunded_by_name?: string }) {
  return {
    id: String(r.id ?? 0),
    billingId: String(r.billing_id ?? 0),
    amount: r.amount,
    reason: r.reason,
    refundedBy: r.refunded_by ?? null,
    refundedByName: r.refunded_by_name ?? "",
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
    paymentSplits: splits && splits.length > 0
      ? splits.map((s) => transformPaymentSplit(s as PaymentSplitWithStaff))
      : undefined,
    totalRefundedAmount: data.total_refunded_amount ?? 0,
    memo: data.memo || undefined,
  };
}

export type Accounting = ReturnType<typeof transformToAccounting>;
