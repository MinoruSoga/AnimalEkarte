import type { DeliveryTriggerSummaryResponse } from "../api/get-lstep-delivery-trigger-summary";

export type DeliverySummaryKey = Exclude<
  keyof DeliveryTriggerSummaryResponse,
  "excluded_reason_breakdown"
>;

export const DELIVERY_STATUS_CARDS: readonly {
  key: DeliverySummaryKey;
  label: string;
}[] = [
  { key: "scheduled", label: "予定" },
  { key: "fired", label: "送信済" },
  { key: "excluded", label: "除外" },
  { key: "failed", label: "失敗" },
  { key: "suppressed_by_priority", label: "優先度抑制" },
];

export function getTodayDateString(): string {
  return new Date().toISOString().slice(0, 10);
}

export function formatDeliveryMonitorDatetime(iso: string): string {
  const date = new Date(iso);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  const hours = String(date.getHours()).padStart(2, "0");
  const minutes = String(date.getMinutes()).padStart(2, "0");
  return `${year}-${month}-${day} ${hours}:${minutes}`;
}
