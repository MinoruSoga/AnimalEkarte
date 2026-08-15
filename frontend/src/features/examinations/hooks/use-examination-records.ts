import { useMemo } from "react";
import { useGetExaminationsPage } from "../api/get-examinations";
import { normalizeKana } from "@/lib/normalize-kana";
import type { ExaminationFilters } from "../api/get-examinations";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { ExaminationRecord } from "../api/transforms";

// rerender-stable-empty-array: pageResult 未ロード時のフォールバックを毎レンダー新規配列にしない
// （useMemo の deps に使うため参照が安定している必要がある）
const EMPTY_EXAMINATIONS: ExaminationRecord[] = [];

interface PageArgs {
  page: number;
  limit: number;
}

/**
 * BUG-411: page/limit を受け取りサーバサイドページネーションで検査一覧を取得する。
 * status/testType/doctor フィルタとテキスト検索は backend が受け付けないため、
 * 修正前と同じくクライアントサイドで「取得済みデータの中でだけ」適用する
 * （修正前は取得済み20件の中、修正後は取得済み1ページの中 — 制約の性質は変わらず退行ではない）。
 */
export function useFilterExaminationRecords(
  searchTerm: string,
  pageArgs: PageArgs,
  filters?: ExaminationFilters,
  activeFilters?: ActiveFilter[],
) {
  const { data: pageResult, isLoading, error } = useGetExaminationsPage({
    ...filters,
    ...pageArgs,
  });
  const examinationsData = pageResult?.data ?? EMPTY_EXAMINATIONS;

  const filteredRecords = useMemo(() => {
    let result = examinationsData;

    // status フィルタ（クライアントサイド）
    const statusFilter = activeFilters?.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is":           return r.status === statusFilter.value;
          case "is_not":       return r.status !== statusFilter.value;
          case "is_empty":     return !r.status;
          case "is_not_empty": return !!r.status;
          default:             return r.status === statusFilter.value;
        }
      });
    }

    // testType フィルタ（クライアントサイド）
    const testTypeFilter = activeFilters?.find((f) => f.key === "testType");
    if (testTypeFilter && typeof testTypeFilter.value === "string") {
      result = result.filter((r) => {
        switch (testTypeFilter.condition) {
          case "is":           return r.testType === testTypeFilter.value;
          case "is_not":       return r.testType !== testTypeFilter.value;
          case "is_empty":     return !r.testType;
          case "is_not_empty": return !!r.testType;
          default:             return r.testType === testTypeFilter.value;
        }
      });
    }

    // doctor フィルタ（クライアントサイド）
    const doctorFilter = activeFilters?.find((f) => f.key === "doctor");
    if (doctorFilter && typeof doctorFilter.value === "string") {
      result = result.filter((r) => {
        switch (doctorFilter.condition) {
          case "is":           return r.doctor === doctorFilter.value;
          case "is_not":       return r.doctor !== doctorFilter.value;
          case "is_empty":     return !r.doctor;
          case "is_not_empty": return !!r.doctor;
          default:             return r.doctor === doctorFilter.value;
        }
      });
    }

    // テキスト検索
    if (!searchTerm) return result;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return result.filter(
      (r) =>
        normalizeKana(r.ownerName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(r.petName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(r.testType).toLowerCase().includes(normalizedTerm),
    );
  }, [examinationsData, searchTerm, activeFilters]);

  return {
    data: filteredRecords,
    allExaminations: examinationsData,
    isLoading,
    error,
    total: pageResult?.total ?? 0,
    page: pageResult?.page ?? pageArgs.page,
    limit: pageResult?.limit ?? pageArgs.limit,
  };
}
