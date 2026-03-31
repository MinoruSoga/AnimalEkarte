import { useMemo } from "react";
import { useGetVaccinations } from "../api/get-vaccinations";
import type { VaccinationFilters } from "../api/get-vaccinations";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";

export function useFilterVaccinations(
  searchTerm: string,
  filters?: VaccinationFilters,
  activeFilters?: ActiveFilter[],
) {
  const { data: vaccinationsData = [], isLoading, error } = useGetVaccinations(filters);

  const filteredRecords = useMemo(() => {
    let result = vaccinationsData;

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

    if (!searchTerm) return result;
    const lowerTerm = searchTerm.toLowerCase();
    return result.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.vaccineName.toLowerCase().includes(lowerTerm),
    );
  }, [vaccinationsData, searchTerm, activeFilters]);

  return { data: filteredRecords, allVaccinations: vaccinationsData, isLoading, error };
}
