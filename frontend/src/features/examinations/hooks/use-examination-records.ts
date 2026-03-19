import { useMemo } from "react";
import { useGetExaminations } from "../api/get-examinations";
import type { ExaminationFilters } from "../api/get-examinations";

export function useFilterExaminationRecords(
  searchTerm: string,
  filters?: ExaminationFilters,
) {
  const { data: examinationsData = [], isLoading, error } = useGetExaminations(filters);

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return examinationsData;
    const lowerTerm = searchTerm.toLowerCase();
    return examinationsData.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.testType.toLowerCase().includes(lowerTerm)
    );
  }, [examinationsData, searchTerm]);

  return { data: filteredRecords, isLoading, error };
}
