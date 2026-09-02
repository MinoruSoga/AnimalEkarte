import { Building2, Calendar, PawPrint } from "lucide-react";
import { normalizeKana } from "@/lib/normalize-kana";
import { uniqueSortedOptions } from "@/lib/unique-sorted-options";
import type {
  ActiveFilter,
  ActiveSort,
  FilterProperty,
  SortProperty,
} from "@/components/shared/PropertyFilter/types";
import { CONDITIONS_NO_EMPTY } from "@/components/shared/PropertyFilter/types";
import type { Hospitalization } from "@/types";
import type { HospitalizationFilters } from "../api/get-hospitalizations";
import {
  HOSPITALIZATION_FILTER_STATUS,
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  type HospitalizationFilterStatus,
} from "../constants";

type SortKey = "startDate" | "ownerName" | "petName" | "species" | "status";

const HOSPITALIZATION_STATIC_FILTER_PROPERTIES: FilterProperty[] = [
  {
    key: "startDate",
    label: "入院日",
    type: "date-range",
    icon: Calendar,
  },
  {
    key: "hospitalizationType",
    label: "入院区分",
    type: "select",
    icon: Building2,
    conditions: CONDITIONS_NO_EMPTY,
    options: [
      { value: "入院", label: "入院" },
      { value: "ホテル", label: "ホテル" },
    ],
  },
];

export const HOSPITALIZATION_SORT_PROPERTIES: SortProperty[] = [
  { key: "startDate", label: "入院日" },
  { key: "ownerName", label: "飼主名" },
  { key: "petName", label: "ペット名" },
  { key: "species", label: "種" },
  { key: "status", label: "ステータス" },
];

export const HOSPITALIZATION_TAB_ITEMS = [
  { value: HOSPITALIZATION_FILTER_STATUS.ACTIVE, label: "入院中" },
  { value: HOSPITALIZATION_FILTER_STATUS.RESERVED, label: "予約" },
  { value: HOSPITALIZATION_FILTER_STATUS.DISCHARGED, label: "退院済" },
  { value: HOSPITALIZATION_FILTER_STATUS.ALL, label: "すべて" },
];

export function isValidFilterStatus(v: string): v is HospitalizationFilterStatus {
  return Object.values(HOSPITALIZATION_FILTER_STATUS).includes(v as HospitalizationFilterStatus);
}

export type ViewMode = "list" | "board";

export function isValidViewMode(v: string): v is ViewMode {
  return v === "list" || v === "board";
}

export function buildHospitalizationListQueryFilters(
  activeFilters: ActiveFilter[],
  statusFilter: HospitalizationFilterStatus,
  serverPage: number,
): HospitalizationFilters {
  const dateFilter = activeFilters.find((f) => f.key === "startDate")?.value as
    | { from?: string; to?: string }
    | undefined;
  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
    statusFilter,
    page: serverPage,
    limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  };
}

export function buildHospitalizationFilterProperties(
  hospitalizations: Hospitalization[],
): FilterProperty[] {
  const speciesOptions = uniqueSortedOptions(hospitalizations, (h) => h.species);
  return [
    ...HOSPITALIZATION_STATIC_FILTER_PROPERTIES,
    { key: "species", label: "種", type: "select" as const, icon: PawPrint, conditions: CONDITIONS_NO_EMPTY, options: speciesOptions },
  ];
}

export function applyHospitalizationClientFilters(
  hospitalizations: Hospitalization[],
  deferredSearchTerm: string,
  activeFilters: ActiveFilter[],
): Hospitalization[] {
  let result = hospitalizations;

  if (deferredSearchTerm) {
    const normalizedTerm = normalizeKana(deferredSearchTerm).toLowerCase();
    result = result.filter(
      (h) =>
        normalizeKana(h.ownerName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(h.petName).toLowerCase().includes(normalizedTerm) ||
        h.hospitalizationNo.toLowerCase().includes(normalizedTerm),
    );
  }

  const typeFilter = activeFilters.find((f) => f.key === "hospitalizationType");
  if (typeFilter && typeof typeFilter.value === "string") {
    result = result.filter((h) => {
      switch (typeFilter.condition) {
        case "is":           return h.hospitalizationType === typeFilter.value;
        case "is_not":       return h.hospitalizationType !== typeFilter.value;
        default:             return h.hospitalizationType === typeFilter.value;
      }
    });
  }

  const speciesFilter = activeFilters.find((f) => f.key === "species");
  if (speciesFilter && typeof speciesFilter.value === "string") {
    result = result.filter((h) => {
      switch (speciesFilter.condition) {
        case "is":           return h.species === speciesFilter.value;
        case "is_not":       return h.species !== speciesFilter.value;
        case "is_empty":     return !h.species;
        case "is_not_empty": return !!h.species;
        default:             return h.species === speciesFilter.value;
      }
    });
  }

  return result;
}

export function sortHospitalizations(
  hospitalizations: Hospitalization[],
  activeSorts: ActiveSort[],
): Hospitalization[] {
  if (activeSorts.length === 0) return [...hospitalizations];
  const sorted = [...hospitalizations];
  sorted.sort((a, b) => {
    for (const sort of activeSorts) {
      const key = sort.key as SortKey;
      const aVal = String(a[key] ?? "");
      const bVal = String(b[key] ?? "");
      const cmp = aVal.localeCompare(bVal, "ja");
      if (cmp !== 0) return sort.direction === "asc" ? cmp : -cmp;
    }
    return 0;
  });
  return sorted;
}

export interface ServerPagePagination<T> {
  paginatedData: T[];
  totalPages: number;
  totalCount: number;
  startIndex: number;
  endIndex: number;
  currentPage: number;
}

export function buildServerPagePagination<T>(input: {
  rows: T[];
  total: number;
  page: number;
  limit: number;
}): ServerPagePagination<T> {
  const totalPages = Math.max(1, Math.ceil(input.total / input.limit));
  return {
    paginatedData: input.rows,
    totalPages,
    totalCount: input.total,
    startIndex: input.total === 0 ? 0 : (input.page - 1) * input.limit + 1,
    endIndex: Math.min(input.page * input.limit, input.total),
    currentPage: input.page,
  };
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
