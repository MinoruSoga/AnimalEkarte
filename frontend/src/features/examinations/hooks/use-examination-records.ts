import { useMemo } from "react";
import { useGetExaminations } from "../api/get-examinations";
import type { ExaminationFilters } from "../api/get-examinations";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";

export function useFilterExaminationRecords(
  searchTerm: string,
  filters?: ExaminationFilters,
  activeFilters?: ActiveFilter[],
) {
  const { data: examinationsData = [], isLoading, error } = useGetExaminations(filters);

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

    // テキスト検索
    if (!searchTerm) return result;
    const lowerTerm = searchTerm.toLowerCase();
    return result.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.testType.toLowerCase().includes(lowerTerm),
    );
  }, [examinationsData, searchTerm, activeFilters]);

  return { data: filteredRecords, allExaminations: examinationsData, isLoading, error };
}
