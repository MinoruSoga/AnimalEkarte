import { useMemo } from "react";
import { useGetVaccinations } from "../api/get-vaccinations";

export function useVaccinations(searchTerm: string) {
  const { data = [], isLoading, error } = useGetVaccinations();

  const filteredRecords = useMemo(() => {
    if (!searchTerm) return data;
    const lowerTerm = searchTerm.toLowerCase();
    return data.filter(
      (r) =>
        r.ownerName.toLowerCase().includes(lowerTerm) ||
        r.petName.toLowerCase().includes(lowerTerm) ||
        r.vaccineName.toLowerCase().includes(lowerTerm)
    );
  }, [data, searchTerm]);

  return { data: filteredRecords, isLoading, error };
}
