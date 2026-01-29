import { useMemo } from "react";
import { getAccountingListSync } from "../api";

export function useAccountingRecords(searchTerm: string) {
  const records = getAccountingListSync();

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return records;
    const lowerTerm = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm)
    );
  }, [records, searchTerm]);

  return { data: filteredRecords };
}
