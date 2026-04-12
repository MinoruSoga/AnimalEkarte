import type { Accounting, AccountingItem, PaymentInfo, Refund } from "../types";
import type { BackendAccounting, BackendAccountingItem } from "./types";
import type { BillingRefund } from "@/types/generated/models";

function transformAccountingItem(item: BackendAccountingItem): AccountingItem {
  const unitPrice = item.unit_price ?? 0;
  const quantity = item.quantity ?? 1;
  const taxRate = item.tax_rate ?? 0.1;
  return {
    id: String(item.id ?? 0),
    category: item.category as AccountingItem["category"],
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

function buildPaymentInfo(data: BackendAccounting): PaymentInfo | undefined {
  const payment = data.payments?.[0];
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
    method: (payment.method || "cash") as PaymentInfo["method"],
  };
}

export function transformToRefund(r: BillingRefund): Refund {
  return {
    id: String(r.id ?? 0),
    billingId: String(r.billing_id ?? 0),
    amount: r.amount,
    reason: r.reason,
    refundedAt: r.refunded_at,
    createdAt: r.created_at,
  };
}

// Backend → フロントエンド Accounting 型（一覧・詳細共通）
export function transformToAccounting(data: BackendAccounting): Accounting {
  return {
    id: String(data.id ?? 0),
    medicalRecordId: data.medical_record_id ? String(data.medical_record_id) : undefined,
    ownerId: String(data.owner_id ?? 0),
    ownerName: data.owner?.name ?? "",
    petId: String(data.pet_id ?? 0),
    petName: data.pet?.name ?? "",
    petSpecies: data.pet?.animal_species?.name,
    status: data.status as Accounting["status"],
    scheduledDate: data.scheduled_date ? data.scheduled_date.slice(0, 10) : "",
    completedAt: data.completed_at ?? undefined,
    items: (data.items ?? []).map(transformAccountingItem),
    payment: buildPaymentInfo(data),
    totalRefundedAmount: data.total_refunded_amount ?? 0,
    memo: data.memo || undefined,
  };
}
