import { AlertCircle, Calendar, PawPrint } from "lucide-react";
import type {
  ActiveFilter,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { paths } from "@/config/paths";
import { addDaysISO, todayISODate } from "@/lib/iso-date";
import { normalizedIncludes } from "@/lib/normalize-kana";
import type { CheckupFilters } from "../types";
import type { CheckupRecord } from "../api/transforms";

/**
 * EMR-223: 動物種フィルタの「その他」値。
 * マスタ種「その他」（seed id=6）だけでなく、種名が 犬/猫 以外の行全般を指す。
 */
const SPECIES_FILTER_OTHER_VALUE = "その他";

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
    key: "species",
    label: "動物種",
    type: "select",
    icon: PawPrint,
    options: [
      { value: "犬", label: "犬" },
      { value: "猫", label: "猫" },
      { value: SPECIES_FILTER_OTHER_VALUE, label: SPECIES_FILTER_OTHER_VALUE },
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

export function buildCheckupListFilters(
  activeFilters: ActiveFilter[],
): Pick<CheckupFilters, "startDate" | "endDate" | "nextStartDate" | "nextEndDate"> {
  const today = todayISODate();
  const dateFilter = activeFilters.find((f) => f.key === "date")?.value as
    { from?: string; to?: string } | undefined;
  const alertStatus = activeFilters.find((f) => f.key === "alertStatus")?.value as
    string | undefined;

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
  return checkups.filter(
    (c) =>
      normalizedIncludes(c.petName, deferredSearch) ||
      normalizedIncludes(c.ownerName, deferredSearch) ||
      normalizedIncludes(c.checkupTypeName, deferredSearch) ||
      normalizedIncludes(c.result, deferredSearch),
  );
}

/**
 * EMR-223: 動物種で現在ページの行をクライアント側に絞り込む。
 * filterCheckupsBySearch と同じく取得済みページに対する後処理フィルタであり、
 * GET /v1/checkups のパラメータ（buildCheckupListFilters の返り値）には混入しない。
 * petId が無い行・種名が未解決の行は、フィルタ適用中は除外する。
 */
export function filterCheckupsBySpecies(
  checkups: CheckupRecord[],
  speciesValue: string | undefined,
  speciesNameByPetId: ReadonlyMap<string, string>,
): CheckupRecord[] {
  if (!speciesValue) return checkups;
  return checkups.filter((c) => {
    const speciesName = c.petId ? speciesNameByPetId.get(c.petId) : undefined;
    if (!speciesName) return false;
    if (speciesValue === SPECIES_FILTER_OTHER_VALUE) {
      return speciesName !== "犬" && speciesName !== "猫";
    }
    return speciesName === speciesValue;
  });
}

export function nextListSearchParamsWithPage(prev: URLSearchParams, page: number): URLSearchParams {
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
