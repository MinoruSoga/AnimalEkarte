import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { MonthlyReportResponse } from "@/types/generated/models";

export const getMonthlyReport = async (
  year: number,
  month: number,
): Promise<MonthlyReportResponse> => {
  const { data } = await axios.get<MonthlyReportResponse>("/v1/reports/monthly", {
    params: { year, month },
  });
  return data;
};

export const useGetMonthlyReport = (year: number, month: number) =>
  useQuery({
    queryKey: ["monthly-report", year, month],
    queryFn: () => getMonthlyReport(year, month),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
