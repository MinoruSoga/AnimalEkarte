import { AlertCircle, Calendar } from "lucide-react";
import type { ActiveFilter, FilterProperty, SortProperty } from "@/components/shared/PropertyFilter/types";
import { paths } from "@/config/paths";
import { addDaysISO, todayISODate } from "@/lib/iso-date";
import { normalizeKana } from "@/lib/normalize-kana";
import type { CheckupFilters } from "../api/types";
import type { CheckupRecord } from "../api/transforms";

export const FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "alertStatus",
    label: "期限状態",
    type: "select",
    icon: AlertCircle,
    options: [
      { value: "overdue", label: "期限切れ" },
      { value: "upcoming30", label: "期限間近 (30日以内)" },
    ],
  },
  {
    key: "date",
    label: "日付",
    type: "date-range",
    icon: Calendar,
  },
];

export const PAGE_SIZE = 20;

export const CHECKUPS_SORT_PROPERTIES: SortProperty[] = [
  { key: "date", label: "実施日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "checkupTypeName", label: "健診種別" },
  { key: "nextDate", label: "次回予定" },
];

export function buildCheckupListFilters(activeFilters: ActiveFilter[]): Pick<
  CheckupFilters,
  "startDate" | "endDate" | "nextStartDate" | "nextEndDate"
> {
  const today = todayISODate();
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string }
    | undefined;
  const alertStatus = activeFilters.find((f) => f.key === "alertStatus")?.value as
    | string
    | undefined;

  let nextStartDate: string | undefined;
  let nextEndDate: string | undefined;

  if (alertStatus === "overdue") {
    nextEndDate = addDaysISO(today, -1);
  } else if (alertStatus === "upcoming30") {
    nextStartDate = today;
    nextEndDate = addDaysISO(today, 30);
  }

  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
    nextStartDate,
    nextEndDate,
  };
}

export function filterCheckupsBySearch(
  checkups: CheckupRecord[],
  deferredSearch: string,
): CheckupRecord[] {
  if (!deferredSearch) return checkups;
  const normalizedTerm = normalizeKana(deferredSearch).toLowerCase();
  return checkups.filter(
    (c) =>
      normalizeKana(c.petName).toLowerCase().includes(normalizedTerm) ||
      normalizeKana(c.ownerName).toLowerCase().includes(normalizedTerm) ||
      normalizeKana(c.checkupTypeName).toLowerCase().includes(normalizedTerm) ||
      c.result.toLowerCase().includes(normalizedTerm),
  );
}

export function nextListSearchParamsWithPage(
  prev: URLSearchParams,
  page: number,
): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (page === 1) {
    next.delete("page");
  } else {
    next.set("page", String(page));
  }
  return next;
}

export function nextListSearchParamsWithoutPage(prev: URLSearchParams): URLSearchParams {
  const next = new URLSearchParams(prev);
  next.delete("page");
  return next;
}

export function checkupChartHref(medicalRecordId: string, checkupId: string): string {
  const params = new URLSearchParams({ tab: "定期健診", checkupId });
  return `${paths.medicalRecords.detail.getHref(medicalRecordId)}?${params.toString()}`;
}
