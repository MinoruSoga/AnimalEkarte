import { useMemo } from "react";
import { useGetExaminations } from "../api/get-examinations";
import type { ExaminationFilters } from "../api/get-examinations";

export function useExaminationRecords(
  searchTerm: string,
  filters?: ExaminationFilters,
) {
  const { data = [], isLoading, error } = useGetExaminations(filters);

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return data;
    const lowerTerm = searchTerm.toLowerCase();
    return data.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.testType.toLowerCase().includes(lowerTerm)
    );
  }, [data, searchTerm]);

  return { data: filteredRecords, isLoading, error };
}
