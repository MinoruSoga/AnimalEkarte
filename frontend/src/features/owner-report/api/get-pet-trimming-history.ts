import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { transformTrimming, type TrimmingUI } from "@/lib/transforms/trimming";
import type { TrimmingListResponse } from "@/types/trimming";

/**
 * #158 トリミング履歴（飼主レポート⑦）= 施術の「実施履歴」。
 * status="完了"（施術実施済み）のみを実施日降順で返す。
 * 予約（未実施）・進行中・キャンセルは実施履歴ではないため除外する。
 * date は "YYYY-MM-DD"（未設定は ""）。辞書順 = 時系列順なので降順ソートで新しい順。
 */
export function selectCompletedTrimmingHistory(items: TrimmingUI[]): TrimmingUI[] {
  return items
    .filter((t) => t.status === "完了")
    .sort((a, b) => b.date.localeCompare(a.date));
}

export interface PetTrimmingHistoryResult {
  items: TrimmingUI[];
  /**
   * SD-18: 取得上限(HISTORY_FETCH_LIMIT)により実件数より少ない可能性がある場合 true。
   * total はステータス問わない生の予約件数のため、完了のみに絞った items.length と直接比較せず
   * 「fetch した生行数(rawRows.length) を total が上回るか」で判定する（フィルタ後件数で判定すると、
   * 生件数側で truncate されていても完了以外が除外されて見かけ上 limit 未満になり検知漏れするため）。
   */
  isTruncated: boolean;
}

/**
 * GET /v1/trimmings?pet_id（appointments ベース）から当該ペットの予約を取得し、
 * 施術実施済み（完了）のみを実施日降順で返す。
 * 一覧 API は TrimmingDetail.Course / Doctor を preload するため、コース名・担当が埋まる。
 */
const getPetTrimmingHistory = async (petId: string): Promise<PetTrimmingHistoryResult> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: petId, page: 1, limit: HISTORY_FETCH_LIMIT },
  });
  const rawRows = data.data ?? [];
  return {
    items: selectCompletedTrimmingHistory(rawRows.map(transformTrimming)),
    isTruncated: typeof data.total === "number" && data.total > rawRows.length,
  };
};

export const useGetPetTrimmingHistory = (petId?: string) => {
  return useQuery({
    queryKey: queryKeys.petTrimmingHistory(petId!),
    queryFn: () => getPetTrimmingHistory(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
