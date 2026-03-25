import { useMemo } from "react";
import { useGetMedicalRecords } from "../api/get-medical-records";
import type { MedicalRecordFilters } from "../api/get-medical-records";

export function useFilterMedicalRecords(
  searchTerm: string,
  filters?: MedicalRecordFilters,
) {
  const { data: records = [], isLoading, isError } = useGetMedicalRecords(filters);

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

  return { data: filteredRecords, isLoading, isError };
}
