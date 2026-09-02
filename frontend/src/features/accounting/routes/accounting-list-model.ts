import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { AccountingPageFilters, AccountingSelectFilterOp } from "../api/get-accountings";

export const TABS = [
  { value: "list", label: "会計一覧" },
  { value: "daily", label: "当日会計" },
  { value: "unpaid", label: "未納者一覧" },
] as const;

export const CLINIC_TOGGLE_RESET_PARAMS = ["page"] as const;

export const ACCOUNTING_PAGE_SIZE = 20;

export type AccountingListTab = "list" | "daily" | "unpaid";

export function accountingListTabFromParam(tabParam: string | null): AccountingListTab {
  if (tabParam === "unpaid") return "unpaid";
  if (tabParam === "daily") return "daily";
  return "list";
}

export function nextAccountingListTabSearchParams(
  prev: URLSearchParams,
  tab: string,
): URLSearchParams {
  const next = new URLSearchParams(prev);
  if (tab === "unpaid") {
    next.set("tab", "unpaid");
    next.delete("page");
    next.delete("daily_date");
  } else if (tab === "daily") {
    next.set("tab", "daily");
    next.delete("page");
    next.delete("group_by");
    next.delete("reference_date");
  } else {
    next.delete("tab");
    next.delete("page");
    next.delete("group_by");
    next.delete("reference_date");
    next.delete("daily_date");
  }
  return next;
}

function extractSelectFilter(
  filters: ActiveFilter[],
  key: string,
): { value?: string; op?: AccountingSelectFilterOp } {
  const f = filters.find((x) => x.key === key);
  if (!f) return {};
  const op = (f.condition ?? "is") as AccountingSelectFilterOp;
  if (op === "is_empty" || op === "is_not_empty") {
    return { op };
  }
  if (typeof f.value === "string" && f.value !== "") {
    return { value: f.value, op };
  }
  return {};
}

export function buildAccountingListPageFilters(input: {
  activeFilters: ActiveFilter[];
  deferredSearch: string;
  isMultiClinic: boolean;
  selectedClinicIds: string[];
  urlPage: number;
}): AccountingPageFilters {
  const dateFilter = input.activeFilters.find((f) => f.key === "date")?.value as
    | { from?: string; to?: string }
    | undefined;
  const status = extractSelectFilter(input.activeFilters, "status");
  const paymentMethod = extractSelectFilter(input.activeFilters, "paymentMethod");
  return {
    startDate: dateFilter?.from,
    endDate: dateFilter?.to,
    search: input.deferredSearch.trim() || undefined,
    status: status.value,
    statusOp: status.op,
    paymentMethod: paymentMethod.value,
    paymentMethodOp: paymentMethod.op,
    clinicIds: input.isMultiClinic ? input.selectedClinicIds : undefined,
    page: input.urlPage,
    limit: ACCOUNTING_PAGE_SIZE,
  };
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
