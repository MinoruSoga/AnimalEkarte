import { useMemo, useCallback } from "react";
import { useGetTrimmings } from "../api/get-trimmings";
import { useDeleteTrimming } from "../api/delete-trimming";
import type { ActiveFilter } from "@/components/shared/NotionFilter/types";

interface DateRange {
  from: string;
  to: string;
}

export function useFilterTrimmingRecords(
  searchTerm: string,
  dateRange: DateRange,
  activeFilters?: ActiveFilter[],
) {
  const { data: trimmingRecords = [], isLoading, error } = useGetTrimmings();
  const deleteMutation = useDeleteTrimming();
  const { from, to } = dateRange; // プリミティブを抽出 (rerender-dependencies)

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

    return result.filter((r) => {
      const matchesKeyword =
        searchTerm === "" ||
        r.ownerName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        r.petName.toLowerCase().includes(searchTerm.toLowerCase());

      // appointment_date はISO文字列なので日付部分（YYYY-MM-DD）で比較
      const recordDate = r.date.slice(0, 10);
      const matchesDate =
        (!from || recordDate >= from) &&
        (!to || recordDate <= to);

      return matchesKeyword && matchesDate;
    });
  }, [trimmingRecords, searchTerm, from, to, activeFilters]);

  const deleteRecord = useCallback((id: string) => {
    deleteMutation.mutate(id);
  }, [deleteMutation]);

  return { data: filteredRecords, isLoading, error, deleteRecord };
}
