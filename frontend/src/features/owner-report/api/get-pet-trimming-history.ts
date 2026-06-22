import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformTrimming, type TrimmingUI } from "@/features/trimming";
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

/**
 * GET /v1/trimmings?pet_id（appointments ベース）から当該ペットの予約を取得し、
 * 施術実施済み（完了）のみを実施日降順で返す。
 * 一覧 API は TrimmingDetail.Course / Doctor を preload するため、コース名・担当が埋まる。
 */
const getPetTrimmingHistory = async (petId: string): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: Number(petId), page: 1, limit: 100 },
  });
  return selectCompletedTrimmingHistory((data.data ?? []).map(transformTrimming));
};

export const useGetPetTrimmingHistory = (petId?: string) => {
  return useQuery({
    queryKey: ["pet-trimmings", "report", petId],
    queryFn: () => getPetTrimmingHistory(petId!),
    enabled: !!petId,
    staleTime: QUERY_STALE_TIMES.MEDIUM,
    gcTime: QUERY_GC_TIMES.STANDARD,
  });
};
