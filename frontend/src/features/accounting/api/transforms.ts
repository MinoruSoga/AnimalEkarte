import type { AccountingRecord } from "@/types";
import type { Accounting, AccountingItem, PaymentInfo } from "../types";
import type { BackendAccounting, BackendAccountingItem } from "./types";

function transformAccountingItem(item: BackendAccountingItem): AccountingItem {
  return {
    id: item.id,
    code: item.code,
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
  if (data.status !== "回収済" || data.billing_amount == null) {
    return undefined;
  }
  return {
    subtotal: data.subtotal ?? 0,
    taxTotal: data.tax_total ?? 0,
    totalAmount: data.total_amount ?? 0,
    insuranceName: data.insurance_name,
    insuranceRatio: data.insurance_ratio ?? undefined,
    insuranceAmount: data.insurance_amount ?? 0,
    discountAmount: data.discount_amount ?? 0,
    billingAmount: data.billing_amount ?? 0,
    receivedAmount: data.received_amount ?? 0,
    changeAmount: data.change_amount ?? 0,
    method: (data.payment_method || "cash") as PaymentInfo["method"],
  };
}

function mapStatus(status: BackendAccounting["status"]): Accounting["status"] {
  const map: Record<BackendAccounting["status"], Accounting["status"]> = {
    未収: "waiting",
    保留: "pending",
    回収済: "completed",
    キャンセル: "cancelled",
  };
  return map[status] ?? "waiting";
}

// Backend → フロントエンド Accounting 型（詳細画面用）
export function transformToAccounting(data: BackendAccounting): Accounting {
  return {
    id: data.id,
    medicalRecordId: data.medical_record_id ?? undefined,
    ownerId: data.owner_id,
    ownerName: data.owner?.name ?? "",
    petId: data.pet_id,
    petName: data.pet?.name ?? "",
    petSpecies: data.pet?.species,
    status: mapStatus(data.status),
    scheduledDate: data.scheduled_date.slice(0, 10),
    completedAt: data.completed_at ?? undefined,
    items: (data.accounting_items ?? []).map(transformAccountingItem),
    payment: buildPaymentInfo(data),
    memo: data.memo,
  };
}

// Backend → AccountingRecord 型（一覧表示用）
export function transformAccounting(data: BackendAccounting): AccountingRecord {
  const methodMap: Record<string, AccountingRecord["method"]> = {
    現金: "現金",
    クレジットカード: "クレジットカード",
    電子マネー: "電子マネー",
  };
  const method = methodMap[data.payment_method ?? ""] ?? "-";

  const statusMap: Record<BackendAccounting["status"], AccountingRecord["status"]> = {
    未収: "未収",
    保留: "未収",
    回収済: "回収済",
    キャンセル: "キャンセル",
  };
  const status = statusMap[data.status] ?? "未収";

  return {
    id: data.id,
    date: data.scheduled_date.slice(0, 10),
    ownerName: data.owner?.name ?? "",
    petName: data.pet?.name ?? "",
    amount: data.total_amount ?? data.billing_amount ?? 0,
    method,
    status,
    note: data.memo,
  };
}
