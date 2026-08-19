import { AGGREGATION_CPM_STAGE_SHORT_LABELS } from "@/lib/cpm-stage";

import type { AggregationOwner } from "../api/get-aggregations";
import type { AggregationTab } from "./aggregation-filter-panel-model";

interface CsvColumnDef {
  header: string;
  getValue: (o: AggregationOwner) => string;
}

/**
 * Escape a free-text CSV cell and neutralize spreadsheet formula injection.
 * Leading = + - @ (and tab/CR, which can hide those prefixes) get a single-quote
 * prefix so Excel/Sheets treat the value as text. Then double-quote escape.
 * Mirrors backend sanitizeCSVCell (lstep_tag_summary_service.go).
 */
export function escapeCsvTextCell(value: string): string {
  let cell = value;
  if (cell.length > 0) {
    const first = cell[0];
    if (
      first === "=" ||
      first === "+" ||
      first === "-" ||
      first === "@" ||
      first === "\t" ||
      first === "\r"
    ) {
      cell = "'" + cell;
    }
  }
  return `"${cell.replace(/"/g, '""')}"`;
}

const CSV_COMMON_COLUMNS: CsvColumnDef[] = [
  { header: "owner_id", getValue: (o) => o.owner_id },
  {
    header: "owner_name",
    getValue: (o) => escapeCsvTextCell(o.owner_name),
  },
  // ISSUE-180: CPM セグメント列（全タブ共通・owner 直後）。短縮ラベルを出力する。
  // 画面の一覧列 (AggregationOwnerTableColumns) と表示を揃える。
  {
    header: "cpm_stage",
    getValue: (o) => (o.cpm_stage ? AGGREGATION_CPM_STAGE_SHORT_LABELS[o.cpm_stage] : ""),
  },
];

const CSV_COLUMNS: Record<AggregationTab, CsvColumnDef[]> = {
  revenue: [
    ...CSV_COMMON_COLUMNS,
    { header: "annual_amount", getValue: (o) => String(o.annual_amount ?? "") },
    { header: "billing_count", getValue: (o) => String(o.billing_count ?? "") },
    {
      header: "period_visit_count",
      getValue: (o) => String(o.period_visit_count ?? ""),
    },
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "first_visit_date", getValue: (o) => o.first_visit_date ?? "" },
  ],
  visit: [
    ...CSV_COMMON_COLUMNS,
    {
      header: "period_visit_count",
      getValue: (o) => String(o.period_visit_count ?? ""),
    },
    { header: "total_visit_count", getValue: (o) => String(o.total_visit_count) },
    {
      header: "annual_visit_count",
      getValue: (o) => String(o.annual_visit_count),
    },
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    { header: "first_visit_date", getValue: (o) => o.first_visit_date ?? "" },
  ],
  // 仕様書 §4.3 表示項目に合わせる: 飼主名 / CPM / 最終来院日 / 経過日数 / 分類 / 累計来院回数 / 年間来院回数 / 累計診療費
  // 画面 (AggregationOwnerTableColumns) と1対1で揃えること。drift 防止。
  last_visit: [
    ...CSV_COMMON_COLUMNS,
    { header: "last_visit_date", getValue: (o) => o.last_visit_date ?? "" },
    {
      header: "days_since_last_visit",
      getValue: (o) => String(o.days_since_last_visit ?? ""),
    },
    { header: "last_visit_bucket", getValue: (o) => o.last_visit_bucket ?? "" },
    { header: "total_visit_count", getValue: (o) => String(o.total_visit_count) },
    {
      header: "annual_visit_count",
      getValue: (o) => String(o.annual_visit_count),
    },
    {
      header: "total_amount",
      getValue: (o) => String(o.total_amount ?? o.total_fee ?? ""),
    },
  ],
};

export function buildCsvContent(
  owners: AggregationOwner[],
  tab: AggregationTab
): string {
  const columns = CSV_COLUMNS[tab];
  const header = columns.map((col) => col.header).join(",");

  const rows = owners.map((o) =>
    columns.map((col) => col.getValue(o)).join(",")
  );

  return [header, ...rows].join("\n");
}
