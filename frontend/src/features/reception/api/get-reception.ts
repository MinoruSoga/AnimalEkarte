import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { DAY_VIEW_FETCH_LIMIT } from "@/config/fetch-limits";
import { todayJSTISO } from "@/lib/jst-date";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformReservationsToReceptionColumns } from "./transforms";
import type { Reservation as BackendReceptionReservation } from "@/types/generated/models";
import type { ReceptionColumn } from "./types";

interface ReservationsResponse {
  data: BackendReceptionReservation[];
  total: number;
  page: number;
  limit: number;
}

/** 今日の日付を YYYY-MM-DD 形式で返す */
export function todayISO(): string {
  return todayJSTISO();
}

/** 指定日の予約を当日受付用カラム配列として取得 */
async function getReception(date: string): Promise<ReceptionColumn[]> {
  const { data } = await axios.get<ReservationsResponse>("/v1/reservations", {
    params: { date, limit: DAY_VIEW_FETCH_LIMIT },
  });
  return transformReservationsToReceptionColumns(data.data);
}

/** 当日の受付用 React Query hook */
export function useGetReception(date: string = todayISO()) {
  return useQuery({
    queryKey: queryKeys.reception.byDate(date),
    queryFn: () => getReception(date),
    // 30秒ごとにポーリングしてリアルタイム性を確保
    refetchInterval: 30_000,
    // client-swr-dedup: staleTimeをrefetchIntervalより短く設定。
    // staleTime=0（デフォルト）のままではコンポーネント再マウント時に即時再取得が走り
    // ポーリングとは別に重複リクエストが発生する。
    staleTime: QUERY_STALE_TIMES.REALTIME,
    gcTime: QUERY_GC_TIMES.STANDARD,
    // ウィンドウフォーカス時も再取得
    refetchOnWindowFocus: true,
  });
}
