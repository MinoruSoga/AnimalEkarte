import type { AccountingRecord } from "@/types";
import type { Accounting, AccountingItem, PaymentInfo } from "../types";
import type { BackendAccounting, BackendAccountingItem } from "./types";

function transformAccountingItem(item: BackendAccountingItem): AccountingItem {
  return {
    id: String(item.id ?? 0),
    category: item.category as AccountingItem["category"],
    name: item.name,
    unitPrice: item.unit_price ?? 0,
    quantity: item.quantity,
    taxRate: (item.tax_rate === 0.08 ? 0.08 : 0.1) as 0.1 | 0.08,
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

// Backend → フロントエンド Accounting 型（詳細画面用）
export function transformToAccounting(data: BackendAccounting): Accounting {
  return {
    id: String(data.id ?? 0),
    medicalRecordId: data.medical_record_id ? String(data.medical_record_id) : undefined,
    ownerId: String(data.owner_id ?? 0),
    ownerName: data.owner?.owner_name ?? "",
    petId: String(data.pet_id ?? 0),
    petName: data.pet?.name ?? "",
    petSpecies: data.pet?.animal_species?.name,
    status: data.status as Accounting["status"],
    scheduledDate: data.scheduled_date.slice(0, 10),
    completedAt: data.completed_at ?? undefined,
    items: (data.items ?? []).map(transformAccountingItem),
    payment: buildPaymentInfo(data),
    memo: data.memo || undefined,
  };
}

// Backend → AccountingRecord 型（一覧表示用）
export function transformAccounting(data: BackendAccounting): AccountingRecord {
  const payment = data.payments?.[0];

  const methodMap: Record<string, AccountingRecord["method"]> = {
    cash: "現金",
    credit_card: "クレジットカード",
    electronic_money: "電子マネー",
  };
  const method = methodMap[payment?.method ?? ""] ?? "-";

  const statusMap: Record<string, AccountingRecord["status"]> = {
    waiting: "会計待ち",
    pending: "会計待ち",
    completed: "会計済",
    cancelled: "キャンセル",
  };
  const status = statusMap[data.status] ?? "会計待ち";

  return {
    id: String(data.id ?? 0),
    date: data.scheduled_date.slice(0, 10),
    ownerName: data.owner?.owner_name ?? "",
    petName: data.pet?.name ?? "",
    amount: data.total_amount ?? payment?.billing_amount ?? 0,
    method,
    status,
    note: data.memo || undefined,
  };
}
