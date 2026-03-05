import { useMemo } from "react";
import { useGetTrimmings, useDeleteTrimming } from "../api";

interface DateRange {
  from: string;
  to: string;
}

export function useTrimmingRecords(searchTerm: string, dateRange: DateRange) {
  const { data = [], isLoading, error } = useGetTrimmings();
  const deleteMutation = useDeleteTrimming();

  const filteredRecords = useMemo(() => {
    return data.filter((r) => {
      const matchesKeyword =
        searchTerm === "" ||
        r.ownerName.toLowerCase().includes(searchTerm.toLowerCase()) ||
        r.petName.toLowerCase().includes(searchTerm.toLowerCase());

      // appointment_date はISO文字列なので日付部分（YYYY-MM-DD）で比較
      const recordDate = r.date.slice(0, 10);
      const matchesDate =
        (!dateRange.from || recordDate >= dateRange.from) &&
        (!dateRange.to || recordDate <= dateRange.to);

      return matchesKeyword && matchesDate;
    });
  }, [data, searchTerm, dateRange]);

  const deleteRecord = (id: string) => {
    deleteMutation.mutate(id);
  };

  return { data: filteredRecords, isLoading, error, deleteRecord };
}
