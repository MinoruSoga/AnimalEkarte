import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
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
  });
