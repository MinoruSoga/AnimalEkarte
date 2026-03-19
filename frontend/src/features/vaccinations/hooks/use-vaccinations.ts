import { useMemo } from "react";
import { useGetVaccinations } from "../api/get-vaccinations";
import type { VaccinationFilters } from "../api/get-vaccinations";

export function useFilterVaccinations(
  searchTerm: string,
  filters?: VaccinationFilters,
) {
  const { data: vaccinationsData = [], isLoading, error } = useGetVaccinations(filters);

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return vaccinationsData;
    const lowerTerm = searchTerm.toLowerCase();
    return vaccinationsData.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.vaccineName.toLowerCase().includes(lowerTerm)
    );
  }, [vaccinationsData, searchTerm]);

  return { data: filteredRecords, isLoading, error };
}
