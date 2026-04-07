import { useQuery } from "@tanstack/react-query";
import { format } from "date-fns";
import { axios } from "@/lib/axios";
import { transformReservationsToReceptionColumns } from "./transforms";
import type { ReservationAppointment as BackendDashboardReservation } from "@/types/generated/models";
import type { ReceptionColumn } from "./types";

interface ReservationsResponse {
  data: BackendDashboardReservation[];
  total: number;
  page: number;
  limit: number;
}

/** 今日の日付を YYYY-MM-DD 形式で返す */
export function todayISO(): string {
  return format(new Date(), "yyyy-MM-dd");
}

/** 指定日の予約をダッシュボード用カラム配列として取得 */
export async function getReception(date: string): Promise<ReceptionColumn[]> {
  const { data } = await axios.get<ReservationsResponse>(
    "/v1/reservations",
    { params: { date } }
  );
  return transformReservationsToReceptionColumns(data.data);
}

/** ダッシュボード用 React Query hook */
export function useGetReception(date: string = todayISO()) {
  return useQuery({
    queryKey: ["reception", date],
    queryFn: () => getReception(date),
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
