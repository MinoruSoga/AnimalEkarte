import { useMemo } from "react";
import { MOCK_VACCINATION_RECORDS } from "../data";

export function useVaccinations(searchTerm: string) {
  const records = MOCK_VACCINATION_RECORDS;

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return records;
    const lowerTerm = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.vaccineName.toLowerCase().includes(lowerTerm)
    );
  }, [records, searchTerm]);

  return { data: filteredRecords };
}
