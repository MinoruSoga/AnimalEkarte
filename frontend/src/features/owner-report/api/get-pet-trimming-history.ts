import { useInfiniteQuery } from "@tanstack/react-query";
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
  return items.filter((t) => t.status === "完了").sort((a, b) => b.date.localeCompare(a.date));
}

export interface PetTrimmingHistoryResult {
  items: TrimmingUI[];
  /**
   * SD-18: 取得上限(HISTORY_FETCH_LIMIT)により実件数より少ない可能性がある場合 true。
   * total はステータス問わない生の予約件数のため、完了のみに絞った items.length と直接比較せず
   * 「fetch した生行数(rawRows.length) を total が上回るか」で判定する（フィルタ後件数で判定すると、
   * 生件数側で truncate されていても完了以外が除外されて見かけ上 limit 未満になり検知漏れするため）。
   * EMR-242: 追加ページ読み込み後も同じ判定を維持する（total > 全ページの累積 raw 行数）。
   */
  isTruncated: boolean;
}

/** 1 ページ分の取得結果。rawCount は完了フィルタ前の生行数。 */
interface PetTrimmingHistoryPage {
  items: TrimmingUI[];
  rawCount: number;
  total?: number;
}

/**
 * GET /v1/trimmings?pet_id（appointments ベース）から当該ペットの予約を取得し、
 * 施術実施済み（完了）のみを実施日降順で返す。
 * 一覧 API は TrimmingDetail.Course / Doctor を preload するため、コース名・担当が埋まる。
 * EMR-242: page を引数化し、追加読み込みではページ単位で完了抽出・降順ソート済みの結果を連結する。
 * マトリクスは日付キーで再グループ化するため、ページ境界の順序差は表示に影響しない。
 */
const getPetTrimmingHistory = async (
  petId: string,
  page: number,
): Promise<PetTrimmingHistoryPage> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: petId, page, limit: HISTORY_FETCH_LIMIT },
  });
  const rawRows = data.data ?? [];
  return {
    items: selectCompletedTrimmingHistory(rawRows.map(transformTrimming)),
    rawCount: rawRows.length,
    total: data.total,
  };
};

/** 累積 raw 行数が total に届いていなければ true（完了フィルタ後の可視件数では判定しない）。 */
function historyHasMore(pages: ReadonlyArray<{ rawCount: number; total?: number }>): boolean {
  const fetched = pages.reduce((sum, page) => sum + page.rawCount, 0);
  const total = pages[pages.length - 1]?.total;
  return typeof total === "number" && total > fetched;
}

export const useGetPetTrimmingHistory = (petId?: string) => {
  return useInfiniteQuery({
    queryKey: queryKeys.petTrimmingHistory(petId!),
    queryFn: ({ pageParam }) => getPetTrimmingHistory(petId!, pageParam),
    initialPageParam: 1,
    getNextPageParam: (_lastPage, allPages) =>
      historyHasMore(allPages) ? allPages.length + 1 : undefined,
    // 消費側の既存 shape ({items, isTruncated}) を維持しつつ全ページを累積する。
    select: (data): PetTrimmingHistoryResult => ({
      items: data.pages.flatMap((page) => page.items),
      isTruncated: historyHasMore(data.pages),
    }),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
