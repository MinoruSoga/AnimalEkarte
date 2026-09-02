import type { CheckupRecord } from "../api/transforms";

export function toCheckupHistoryItems(records: readonly CheckupRecord[]) {
  return records.map((record) => ({
    id: String(record.id),
    date: record.date,
    title: record.checkupTypeName || "健診",
    subtitle: [record.doctorName, record.result].filter(Boolean).join(" / ") || undefined,
  }));
}
