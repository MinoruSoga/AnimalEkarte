import { useMemo } from "react";
import { useGetVaccinations } from "../api/get-vaccinations";
import { normalizeKana } from "@/lib/normalize-kana";
import type { VaccinationFilters } from "../api/get-vaccinations";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";

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
          case "is":
            return r.doctor === doctorFilter.value;
          case "is_not":
            return r.doctor !== doctorFilter.value;
          case "is_empty":
            return !r.doctor;
          case "is_not_empty":
            return !!r.doctor;
          default:
            return r.doctor === doctorFilter.value;
        }
      });
    }

    if (!searchTerm) return result;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return result.filter(
      (r) =>
        normalizeKana(r.ownerName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(r.petName).toLowerCase().includes(normalizedTerm) ||
        normalizeKana(r.vaccineName).toLowerCase().includes(normalizedTerm),
    );
  }, [vaccinationsData, searchTerm, activeFilters]);

  return { data: filteredRecords, allVaccinations: vaccinationsData, isLoading, error };
}
