import { todayJSTISO } from "@/lib/jst-date";
import type { AggregationParams } from "../api/get-aggregations";
import type { AggregationTab } from "../components/aggregation-filter-panel-model";

export const DEFAULT_AGGREGATION_TAB: AggregationTab = "revenue";

const AGGREGATION_TABS: readonly AggregationTab[] = ["revenue", "visit", "last_visit"] as const;

const CURRENT_YEAR = Number(todayJSTISO().slice(0, 4));

export const TAB_DEFAULT_PARAMS: Record<AggregationTab, AggregationParams> = {
  revenue: {
    page: 1,
    per_page: 50,
    year: CURRENT_YEAR,
    amount_basis: "gross_total_amount",
    sort: "annual_amount",
    order: "desc",
  },
  // BUG-012: 来院/最終来院は売上0除外の対象外。UIに include_zero が無いため明示 true。
  visit: {
    page: 1,
    per_page: 50,
    period_preset: "last_12_months",
    sort: "period_visit_count",
    order: "desc",
    include_zero: true,
  },
  last_visit: {
    page: 1,
    per_page: 50,
    last_visit_bucket: "over_3m",
    sort: "last_visit_date",
    order: "asc",
    include_zero: true,
  },
};

export function validateTab(value: unknown): AggregationTab | null {
  return AGGREGATION_TABS.find((t) => t === value) ?? null;
}

// エラーメッセージは「データの読み込みに失敗しました」をベースとし、
// 利用者に原因 (HTTP ステータス等) が分かる形で補助情報を併記する。
// axios エラーは Error インスタンスで `error.message` が `Request failed with status code 404` 等になるため、
// それをそのまま表示すると業務利用者には伝わりにくい。
export function formatAggregationError(error: unknown): string {
  const baseMessage = "データの読み込みに失敗しました";
  if (error !== null && typeof error === "object") {
    const response = (error as { response?: { status?: number; statusText?: string; data?: { error?: string } } }).response;
    if (response?.status) {
      const detail = response.statusText
        ? `HTTP ${response.status} ${response.statusText}`
        : `HTTP ${response.status}`;
      const apiError = response.data?.error;
      return apiError
        ? `${baseMessage} (${detail}: ${apiError})`
        : `${baseMessage} (${detail})`;
    }
    if (error instanceof Error && error.message && !error.message.startsWith("Request failed")) {
      return `${baseMessage}: ${error.message}`;
    }
  }
  return baseMessage;
}

export function downloadCsv(content: string, filename: string): void {
  const bom = "\uFEFF";
  const blob = new Blob([bom + content], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(url);
}

export const AGGREGATION_TAB_ITEMS = [
  { value: "revenue" as const, label: "売上ランキング" },
  { value: "visit" as const, label: "来院回数" },
  { value: "last_visit" as const, label: "最終来院" },
];
