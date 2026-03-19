import { useMemo, useCallback } from "react";
import { useGetTrimmings } from "../api/get-trimmings";
import { useDeleteTrimming } from "../api/delete-trimming";

interface DateRange {
  from: string;
  to: string;
}

export function useFilterTrimmingRecords(searchTerm: string, dateRange: DateRange) {
  const { data: trimmingRecords = [], isLoading, error } = useGetTrimmings();
  const deleteMutation = useDeleteTrimming();
  const { from, to } = dateRange; // プリミティブを抽出 (rerender-dependencies)

  const filteredRecords = useMemo(() => {
    return trimmingRecords.filter((r) => {
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
  }, [trimmingRecords, searchTerm, from, to]); // オブジェクトではなくプリミティブを使用

  const deleteRecord = useCallback((id: string) => {
    deleteMutation.mutate(id);
  }, [deleteMutation]);

  return { data: filteredRecords, isLoading, error, deleteRecord };
}
