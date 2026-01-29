import { useMemo } from "react";
import { MOCK_TRIMMING_RECORDS } from "@/config/mock-data";

interface DateRange {
  from: string;
  to: string;
}

export function useTrimmingRecords(searchTerm: string, dateRange: DateRange) {
  const records = MOCK_TRIMMING_RECORDS;

  const filteredRecords = useMemo(() => {
    return records.filter((r) => {
        const matchesKeyword = 
            searchTerm === "" ||
            r.ownerName.toLowerCase().includes(searchTerm.toLowerCase()) ||
            r.petName.toLowerCase().includes(searchTerm.toLowerCase());
    
        const matchesDate = 
            (!dateRange.from || r.date >= dateRange.from) &&
            (!dateRange.to || r.date <= dateRange.to);
    
        return matchesKeyword && matchesDate;
      });
  }, [records, searchTerm, dateRange]);

  return { data: filteredRecords };
}
