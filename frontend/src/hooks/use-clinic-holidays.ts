import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ClinicHoliday } from "@/types/generated/models";

// GET /v1/clinic-holidays?year_month=YYYY-MM
const getClinicHolidays = async (yearMonth?: string): Promise<ClinicHoliday[]> => {
  const params: Record<string, string> = {};
  if (yearMonth) params.year_month = yearMonth;
  const { data } = await axios.get<ClinicHoliday[]>("/v1/clinic-holidays", { params });
  return data;
};

/**
 * 指定年月の休診日一覧を取得する（共有フック）。
 * features/shifts と同一 query key を使用し React Query キャッシュを共有。
 */
export function useGetClinicHolidays(yearMonth: string) {
  return useQuery({
    queryKey: queryKeys.clinicHolidays.byMonth(yearMonth),
    queryFn: () => getClinicHolidays(yearMonth),
    enabled: !!yearMonth,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
