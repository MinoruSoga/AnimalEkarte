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
    // client-swr-dedup: staleTimeをrefetchIntervalより短く設定。
    // staleTime=0（デフォルト）のままではコンポーネント再マウント時に即時再取得が走り
    // ポーリングとは別に重複リクエストが発生する。
    staleTime: 20_000,
    // ウィンドウフォーカス時も再取得
    refetchOnWindowFocus: true,
  });
}
