import { useMemo } from "react";
import { useGetMedicalRecords } from "../api/get-medical-records";
import type { MedicalRecordFilters } from "../api/get-medical-records";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";

export function useFilterMedicalRecords(
  searchTerm: string,
  filters?: MedicalRecordFilters,
  activeFilters?: ActiveFilter[],
) {
  const { data: records = [], isLoading, isError } = useGetMedicalRecords(filters);

  const filteredRecords = useMemo(() => {
    let result = records;

    // status フィルタ（クライアントサイド）
    const statusFilter = activeFilters?.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is":          return r.status === statusFilter.value;
          case "is_not":      return r.status !== statusFilter.value;
          case "is_empty":    return !r.status;
          case "is_not_empty": return !!r.status;
          default:            return r.status === statusFilter.value;
        }
      });
    }

    // doctor フィルタ（クライアントサイド）
    const doctorFilter = activeFilters?.find((f) => f.key === "doctor");
    if (doctorFilter && typeof doctorFilter.value === "string") {
      result = result.filter((r) => {
        switch (doctorFilter.condition) {
          case "is":           return r.doctor === doctorFilter.value;
          case "is_not":       return r.doctor !== doctorFilter.value;
          case "is_empty":     return !r.doctor;
          case "is_not_empty": return !!r.doctor;
          default:             return r.doctor === doctorFilter.value;
        }
      });
    }

    // species フィルタ（クライアントサイド）
    const speciesFilter = activeFilters?.find((f) => f.key === "species");
    if (speciesFilter && typeof speciesFilter.value === "string") {
      result = result.filter((r) => {
        switch (speciesFilter.condition) {
          case "is":           return r.species === speciesFilter.value;
          case "is_not":       return r.species !== speciesFilter.value;
          case "is_empty":     return !r.species;
          case "is_not_empty": return !!r.species;
          default:             return r.species === speciesFilter.value;
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
        r.recordNo.toLowerCase().includes(lowerTerm) ||
        r.chiefComplaint.toLowerCase().includes(lowerTerm),
    );
  }, [records, searchTerm, activeFilters]);

  return { data: filteredRecords, allRecords: records, isLoading, isError };
}
