import { useInfiniteQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { HISTORY_FETCH_LIMIT } from "@/config/fetch-limits";
import { queryKeys } from "@/lib/query-keys";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformExamination, type ExaminationRecord } from "@/lib/transforms/examination";
import type { Examination } from "@/types/generated/models";

// #158 §計画補足: 履歴表示は下書き（依頼中/検査中）を除外する。
// 既存 transformExamination は status を日本語ラベルに変換するため、そのラベルで除外する。
const DRAFT_STATUS_LABELS = new Set(["依頼中", "検査中"]);

export interface PetExaminationHistoryResult {
  items: ExaminationRecord[];
  /**
   * SD-18: 取得上限(HISTORY_FETCH_LIMIT)により実件数より少ない可能性がある場合 true。
   * バックエンド total（下書き除外前の生件数）が実際に fetch した件数を上回るかで判定する
   * （下書き除外フィルタ後の件数で判定すると、生件数側で truncate されていても除外分が
   * 相殺されて見かけ上 limit 未満になり検知漏れするため）。
   * EMR-242: 追加ページ読み込み後も同じ判定を維持する（total > 全ページの累積 raw 行数）。
   */
  isTruncated: boolean;
}

/** 1 ページ分の取得結果。rawCount は下書き除外フィルタ前の生行数。 */
interface PetExaminationsPage {
  items: ExaminationRecord[];
  rawCount: number;
  total?: number;
}

const getPetExaminations = async (petId: string, page: number): Promise<PetExaminationsPage> => {
  const { data } = await axios.get<{ data: Examination[]; total?: number }>("/v1/examinations", {
    params: { pet_id: petId, page, limit: HISTORY_FETCH_LIMIT, include_items: true },
  });
  const rawRows = data.data ?? [];
  const items = rawRows.flatMap((row) => {
    const examination = transformExamination(row);
    return DRAFT_STATUS_LABELS.has(examination.status) ? [] : [examination];
  });
  return { items, rawCount: rawRows.length, total: data.total };
};

/** 累積 raw 行数が total に届いていなければ true（可視件数では判定しない）。 */
function historyHasMore(pages: ReadonlyArray<{ rawCount: number; total?: number }>): boolean {
  const fetched = pages.reduce((sum, page) => sum + page.rawCount, 0);
  const total = pages[pages.length - 1]?.total;
  return typeof total === "number" && total > fetched;
}

export const useGetPetExaminations = (petId?: string) => {
  return useInfiniteQuery({
    queryKey: queryKeys.petExaminationsReport(petId!),
    queryFn: ({ pageParam }) => getPetExaminations(petId!, pageParam),
    initialPageParam: 1,
    getNextPageParam: (_lastPage, allPages) =>
      historyHasMore(allPages) ? allPages.length + 1 : undefined,
    // 消費側の既存 shape ({items, isTruncated}) を維持しつつ全ページを累積する。
    select: (data): PetExaminationHistoryResult => ({
      items: data.pages.flatMap((page) => page.items),
      isTruncated: historyHasMore(data.pages),
    }),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
