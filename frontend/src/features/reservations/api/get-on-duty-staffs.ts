import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";

// バックエンドの staffSummaryResponse 型（id + name のみ）
export interface OnDutyStaff {
  id: number;
  name: string;
}

// GET /api/v1/shifts/on-duty-staffs?date=YYYY-MM-DD
const fetchOnDutyStaffs = async (date: string): Promise<OnDutyStaff[]> => {
  const { data } = await axios.get<OnDutyStaff[]>("/v1/shifts/on-duty-staffs", {
    params: { date },
  });
  return data;
};

/**
 * 指定日に出勤しているスタッフ一覧を取得する。
 * BUG-344: ReservationFormFields の担当者選択フィルタリングに使用。
 * date が null の場合は無効（全アクティブスタッフを表示する側でフォールバック）。
 */
export function useGetOnDutyStaffs(date: string | null) {
  return useQuery({
    queryKey: ["shifts", "on-duty-staffs", date],
    queryFn: () => fetchOnDutyStaffs(date!),
    enabled: date !== null,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.SHORT,
  });
}
