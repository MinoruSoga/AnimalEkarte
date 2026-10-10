import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import type { ClinicHoliday } from "@/types/generated/models";

// GET /v1/clinic-holidays?year_month=YYYY-MM
const getClinicHolidays = async (
  yearMonth?: string,
  clinicId?: string,
): Promise<ClinicHoliday[]> => {
  const params: Record<string, string> = {};
  if (yearMonth) params.year_month = yearMonth;
  const { data } = await axios.get<ClinicHoliday[]>("/v1/clinic-holidays", {
    params,
    ...(clinicId ? { headers: { "X-Clinic-ID": clinicId } } : {}),
  });
  return data;
};

/**
 * 指定年月の休診日一覧を取得する（共有フック）。
 * features/shifts と同一 query key を使用し React Query キャッシュを共有。
 * clinicId を渡すと X-Clinic-ID を明示し query key を医院で分離する
 * （Codex PR606 再指摘: 拠点横断 tentative 確定フォーム用）。
 */
export function useGetClinicHolidays(yearMonth: string, clinicId?: string) {
  return useQuery({
    queryKey: clinicId
      ? [...queryKeys.clinicHolidays.byMonth(yearMonth), clinicId]
      : queryKeys.clinicHolidays.byMonth(yearMonth),
    queryFn: () => getClinicHolidays(yearMonth, clinicId),
    enabled: !!yearMonth,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
}
