import { useMemo } from "react";
import { useGetExaminations } from "../api";

export function useExaminationRecords(searchTerm: string) {
  const { data = [], isLoading, error } = useGetExaminations();

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
