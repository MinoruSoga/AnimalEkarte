import { useQuery } from "@tanstack/react-query";
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
   */
  isTruncated: boolean;
}

const getPetExaminations = async (petId: string): Promise<PetExaminationHistoryResult> => {
  const { data } = await axios.get<{ data: Examination[]; total?: number }>("/v1/examinations", {
    params: { pet_id: petId, limit: HISTORY_FETCH_LIMIT, include_items: true },
  });
  const rawRows = data.data ?? [];
  const items = rawRows.flatMap((row) => {
    const examination = transformExamination(row);
    return DRAFT_STATUS_LABELS.has(examination.status) ? [] : [examination];
  });
  const isTruncated = typeof data.total === "number" && data.total > rawRows.length;
  return { items, isTruncated };
};

export const useGetPetExaminations = (petId?: string) => {
  return useQuery({
    queryKey: queryKeys.petExaminationsReport(petId!),
    queryFn: () => getPetExaminations(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
