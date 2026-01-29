import { useMemo } from "react";
import { MOCK_EXAMINATION_RECORDS } from "@/config/mock-data";

export function useExaminationRecords(searchTerm: string) {
  const records = MOCK_EXAMINATION_RECORDS;

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return records;
    const lowerTerm = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.testType.toLowerCase().includes(lowerTerm)
    );
  }, [records, searchTerm]);

  return { data: filteredRecords };
}
