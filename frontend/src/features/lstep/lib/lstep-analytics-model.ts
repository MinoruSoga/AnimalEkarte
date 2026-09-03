import { PALETTE } from "@/lib/design-tokens";
import { todayJSTISO, toJSTWallDate } from "@/lib/jst-date";

import type { DeliveryStatsRow } from "../api/get-lstep-delivery-stats";
// FE4-12: 正本は trigger-types.ts の TriggerStatusLabels（内容同一）。参照名は既存流儀の STATUS_LABELS を維持。
export { TriggerStatusLabels as STATUS_LABELS } from "../constants/trigger-types";

export function currentYearMonth(): string {
  return todayJSTISO().slice(0, 7);
}

export function generateMonthOptions(count = 12): { value: string; label: string }[] {
  const options: { value: string; label: string }[] = [];
  const now = toJSTWallDate(new Date());
  for (let i = 0; i < count; i++) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const value = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
    const label = `${d.getFullYear()}年${d.getMonth() + 1}月`;
    options.push({ value, label });
  }
  return options;
}

export const STATS_STATUSES = ["fired", "excluded", "failed", "scheduled"] as const;
export type StatsStatus = (typeof STATS_STATUSES)[number];

export const STATUS_COLORS: Record<StatsStatus, string> = {
  fired: PALETTE.successGreen,
  excluded: PALETTE.chartAxisText,
  failed: PALETTE.danger,
  scheduled: PALETTE.accent,
};

export const CSV_STATUS_LABELS: Record<string, string> = {
  pending: "処理待ち",
  processing: "処理中",
  completed: "完了",
  failed: "失敗",
};

export interface CrossRow {
  trigger_type: string;
  fired: number;
  excluded: number;
  failed: number;
  scheduled: number;
  total: number;
}

export function buildCrossRows(rows: DeliveryStatsRow[]): CrossRow[] {
  const map = new Map<string, CrossRow>();
  for (const row of rows) {
    if (!map.has(row.trigger_type)) {
      map.set(row.trigger_type, {
        trigger_type: row.trigger_type,
        fired: 0,
        excluded: 0,
        failed: 0,
        scheduled: 0,
        total: 0,
      });
    }
    const entry = map.get(row.trigger_type)!;
    const status = row.status as StatsStatus;
    if (status in entry) {
      entry[status] += Number(row.count);
    }
    entry.total += Number(row.count);
  }
  return Array.from(map.values()).sort((a, b) => b.total - a.total);
}

export function formatPercent(rate: number): string {
  return `${(rate * 100).toFixed(1)}%`;
}
