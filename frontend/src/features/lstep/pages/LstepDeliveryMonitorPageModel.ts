import { formatJSTWallDate, formatJSTWallTime, todayJSTISO, toJSTWallDate } from "@/lib/jst-date";
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
  return todayJSTISO();
}

export function formatDeliveryMonitorDatetime(iso: string): string {
  const date = toJSTWallDate(iso);
  return `${formatJSTWallDate(date)} ${formatJSTWallTime(date)}`;
}
