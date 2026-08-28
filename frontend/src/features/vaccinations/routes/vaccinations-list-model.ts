import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { paths } from "@/config/paths";
import { todayJSTISO } from "@/lib/jst-date";
import type { VaccinationFilters } from "../api/get-vaccinations";
import type { VaccinationRecord } from "@/types";

export const VACCINATION_LIST_CHART_TAB = "予防接種";
export const VACCINATION_LIST_ID_PARAM = "vaccinationId";

export function vaccinationListDetailHref(input: {
  id: string;
  medicalRecordId?: string;
}): string {
  if (input.medicalRecordId) {
    const params = new URLSearchParams({
      tab: VACCINATION_LIST_CHART_TAB,
      [VACCINATION_LIST_ID_PARAM]: input.id,
    });
    return `${paths.medicalRecords.detail.getHref(input.medicalRecordId)}?${params.toString()}`;
  }
  return paths.vaccinations.detail.getHref(input.id);
}

export function vaccinationCreateHref(petId: string): string {
  // BUG-501: 一覧→新規はカルテ未保存ゲートを避ける独立フォームへ（仕様 15 §1.1 / S03#1）
  const params = new URLSearchParams({ petId });
  return `${paths.vaccinations.new.getHref()}?${params.toString()}`;
}

export interface VaccinationListQueryInput {
  /** PropertyFilter date-range value */
  dateRange?: { from?: string; to?: string };
  /** Free-text search (pet/owner/vaccine). Trimmed; empty → omitted. */
  search?: string;
  page?: number;
  limit?: number;
  /** Injectable clock for tests; defaults to todayJSTISO(). */
  today?: string;
}

/**
 * BUG-502: immutable default list query options.
 *
 * Far-future seed rows (e.g. 実施日 2029-12-01) used to fill the HISTORY_FETCH_LIMIT
 * window under server `date DESC`, hiding same-day S03 registrations and their
 * near-term next_date. Default end_date = today (JST) keeps the window on real
 * 実績; search is always server-scoped so PACO etc. are not client-filtered out
 * of an unscoped 100-row page.
 */
export function buildVaccinationListQueryOptions(
  input: VaccinationListQueryInput = {},
): VaccinationFilters {
  const today = input.today ?? todayJSTISO();
  const from = input.dateRange?.from?.trim() || undefined;
  const to = input.dateRange?.to?.trim() || undefined;
  const search = input.search?.trim() || undefined;

  // Fail closed: reject inverted range by dropping both bounds (caller shows full default window).
  if (from && to && from > to) {
    return {
      page: input.page ?? 1,
      limit: input.limit ?? HISTORY_FETCH_LIMIT,
      search,
      endDate: today,
    };
  }

  return {
    page: input.page ?? 1,
    limit: input.limit ?? HISTORY_FETCH_LIMIT,
    startDate: from,
    // User-supplied `to` wins; otherwise cap at today so future seed dates leave the window.
    endDate: to ?? today,
    search,
  };
}

/**
 * BUG-502: stable default ordering for the list UI when the user has not chosen a column sort.
 * Prefer near-term next_date (asc), then newer created/id as tie-break via original index.
 * Far-future next_date peers sink; missing next_date sorts last.
 */
export function orderVaccinationListRows(
  rows: readonly VaccinationRecord[],
): VaccinationRecord[] {
  return rows
    .map((row, index) => ({ row, index }))
    .sort((a, b) => {
      const aNext = a.row.nextDate?.trim() ?? "";
      const bNext = b.row.nextDate?.trim() ?? "";
      const aHas = aNext.length > 0;
      const bHas = bNext.length > 0;
      if (aHas && bHas) {
        const byNext = aNext.localeCompare(bNext);
        if (byNext !== 0) return byNext;
      } else if (aHas !== bHas) {
        return aHas ? -1 : 1;
      }
      const byDate = (b.row.date ?? "").localeCompare(a.row.date ?? "");
      if (byDate !== 0) return byDate;
      return a.index - b.index;
    })
    .map(({ row }) => row);
}
