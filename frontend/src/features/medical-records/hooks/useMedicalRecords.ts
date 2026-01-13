import { useMemo } from "react";
import { MOCK_MEDICAL_RECORDS } from "../../../lib/constants";

export function useMedicalRecords(searchTerm: string) {
  const records = MOCK_MEDICAL_RECORDS;

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return records;
    const lowerTerm = searchTerm.toLowerCase();
    return records.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.recordNo.toLowerCase().includes(lowerTerm) ||
        r.chiefComplaint.toLowerCase().includes(lowerTerm)
    );
  }, [records, searchTerm]);

  return { data: filteredRecords };
}
