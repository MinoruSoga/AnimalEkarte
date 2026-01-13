import { useMemo } from "react";
import { MOCK_ACCOUNTING_LIST } from "../mockData";

export function useAccountingRecords(searchTerm: string) {
  const records = MOCK_ACCOUNTING_LIST;

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
