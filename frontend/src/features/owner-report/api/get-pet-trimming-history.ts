import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { transformTrimming, type TrimmingUI } from "@/features/trimming";
import type { TrimmingListResponse } from "@/types/trimming";

/**
 * #158 トリミング履歴（飼主レポート⑦）。
 * GET /v1/trimmings?pet_id（appointments ベース）から当該ペットの予約を取得する。
 * 一覧 API は TrimmingDetail.Course / Doctor を preload するため、コース名・担当が埋まる。
 * ステータス "完了" が施術実施済みを表す。実施日（start_time 由来）の降順で並べる。
 */
const getPetTrimmingHistory = async (petId: string): Promise<TrimmingUI[]> => {
  const { data } = await axios.get<TrimmingListResponse>("/v1/trimmings", {
    params: { pet_id: Number(petId), page: 1, limit: 100 },
  });
  // date は "YYYY-MM-DD"（未設定は ""）。辞書順 = 時系列順なので降順ソートで新しい順。
  return (data.data ?? [])
    .map(transformTrimming)
    .sort((a, b) => b.date.localeCompare(a.date));
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
