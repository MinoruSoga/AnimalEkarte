import { useMemo, useCallback } from "react";
import type { MutateOptions } from "@tanstack/react-query";
import { useGetTrimmingsPage } from "../api/get-trimmings";
import { useDeleteTrimming } from "../api/delete-trimming";
import { normalizeKana } from "@/lib/normalize-kana";
import type { ActiveFilter } from "@/components/shared/PropertyFilter/types";
import type { TrimmingFilters } from "../api/get-trimmings";

export function useFilterTrimmingRecords(
  searchTerm: string,
  filters?: TrimmingFilters,
  activeFilters?: ActiveFilter[],
) {
  const { data: trimmingPage, isLoading, error } = useGetTrimmingsPage(filters);
  const trimmingRecords = useMemo(
    () => trimmingPage?.items ?? [],
    [trimmingPage?.items],
  );
  const isTruncated = trimmingPage?.isTruncated ?? false;
  const { mutate: deleteTrimmingFn } = useDeleteTrimming();

  const filteredRecords = useMemo(() => {
    let result = trimmingRecords;

    // status フィルタ（クライアントサイド）
    const statusFilter = activeFilters?.find((f) => f.key === "status");
    if (statusFilter && typeof statusFilter.value === "string") {
      result = result.filter((r) => {
        switch (statusFilter.condition) {
          case "is":           return r.status === statusFilter.value;
          case "is_not":       return r.status !== statusFilter.value;
          case "is_empty":     return !r.status;
          case "is_not_empty": return !!r.status;
          default:             return r.status === statusFilter.value;
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

    // staff フィルタ（クライアントサイド）
    const staffFilter = activeFilters?.find((f) => f.key === "staff");
    if (staffFilter && typeof staffFilter.value === "string") {
      result = result.filter((r) => {
        switch (staffFilter.condition) {
          case "is":           return r.staff === staffFilter.value;
          case "is_not":       return r.staff !== staffFilter.value;
          case "is_empty":     return !r.staff;
          case "is_not_empty": return !!r.staff;
          default:             return r.staff === staffFilter.value;
        }
      });
    }

    // テキスト検索（日付フィルタはサーバーサイドに移行済み）。#231: 犬種も検索対象に含める。
    if (searchTerm === "") return result;
    const normalizedTerm = normalizeKana(searchTerm).toLowerCase();
    return result.filter((r) =>
      normalizeKana(r.ownerName).toLowerCase().includes(normalizedTerm) ||
      normalizeKana(r.petName).toLowerCase().includes(normalizedTerm) ||
      normalizeKana(r.breed).toLowerCase().includes(normalizedTerm),
    );
  }, [trimmingRecords, searchTerm, activeFilters]);

  const deleteRecord = useCallback(
    (id: string, options?: MutateOptions<void, Error, string>) => {
      deleteTrimmingFn(id, options);
    },
    [deleteTrimmingFn],
  );

  return {
    data: filteredRecords,
    allTrimmings: trimmingRecords,
    isTruncated,
    isLoading,
    error,
    deleteRecord,
  };
}
