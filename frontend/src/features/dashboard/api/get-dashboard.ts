import { useQuery } from "@tanstack/react-query";
import { format } from "date-fns";
import { axios } from "@/lib/axios";
import { transformReservationsToDashboardColumns } from "./transforms";
import type { BackendDashboardReservation, DashboardColumn } from "./types";

/** 今日の日付を YYYY-MM-DD 形式で返す */
export function todayISO(): string {
  return format(new Date(), "yyyy-MM-dd");
}

/** 指定日の予約をダッシュボード用カラム配列として取得 */
export async function getDashboard(date: string): Promise<DashboardColumn[]> {
  const { data } = await axios.get<BackendDashboardReservation[]>(
    "/v1/reservations",
    { params: { date } }
  );
  return transformReservationsToDashboardColumns(data);
}

/** ダッシュボード用 React Query hook */
export function useDashboardData(date: string = todayISO()) {
  return useQuery({
    queryKey: ["dashboard", date],
    queryFn: () => getDashboard(date),
    // 30秒ごとにポーリングしてリアルタイム性を確保
    refetchInterval: 30_000,
    // ウィンドウフォーカス時も再取得
    refetchOnWindowFocus: true,
  });
}
